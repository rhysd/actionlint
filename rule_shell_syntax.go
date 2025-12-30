package actionlint

import (
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

// RuleShellSyntax is a rule to check syntax of shell scripts in 'run:' using mvdan/sh parser.
// This provides built-in shell syntax validation without requiring external tools like shellcheck.
type RuleShellSyntax struct {
	RuleBase
	workflowShell string
	jobShell      string
	runnerShell   string
}

// NewRuleShellSyntax creates new RuleShellSyntax instance.
func NewRuleShellSyntax() *RuleShellSyntax {
	return &RuleShellSyntax{
		RuleBase: RuleBase{
			name: "shell-syntax",
			desc: "Checks for syntax errors in shell scripts in \"run:\"",
		},
	}
}

// VisitStep is callback when visiting Step node.
func (rule *RuleShellSyntax) VisitStep(n *Step) error {
	run, ok := n.Exec.(*ExecRun)
	if !ok || run.Run == nil {
		return nil
	}

	shell := rule.getShellName(run)
	rule.checkShellSyntax(run.Run.Value, shell, run.RunPos)
	return nil
}

// VisitJobPre is callback when visiting Job node before visiting its children.
func (rule *RuleShellSyntax) VisitJobPre(n *Job) error {
	if n.Defaults != nil && n.Defaults.Run != nil && n.Defaults.Run.Shell != nil {
		rule.jobShell = n.Defaults.Run.Shell.Value
	}

	if n.RunsOn != nil {
		for _, label := range n.RunsOn.Labels {
			l := strings.ToLower(label.Value)
			if l == "windows" || strings.HasPrefix(l, "windows-") {
				rule.runnerShell = "pwsh"
				break
			}
		}
	}

	return nil
}

// VisitJobPost is callback when visiting Job node after visiting its children.
func (rule *RuleShellSyntax) VisitJobPost(n *Job) error {
	rule.jobShell = ""
	rule.runnerShell = ""
	return nil
}

// VisitWorkflowPre is callback when visiting Workflow node before visiting its children.
func (rule *RuleShellSyntax) VisitWorkflowPre(n *Workflow) error {
	if n.Defaults != nil && n.Defaults.Run != nil && n.Defaults.Run.Shell != nil {
		rule.workflowShell = n.Defaults.Run.Shell.Value
	}
	return nil
}

// VisitWorkflowPost is callback when visiting Workflow node after visiting its children.
func (rule *RuleShellSyntax) VisitWorkflowPost(n *Workflow) error {
	rule.workflowShell = ""
	return nil
}

func (rule *RuleShellSyntax) getShellName(exec *ExecRun) string {
	if exec.Shell != nil {
		return exec.Shell.Value
	}
	if rule.jobShell != "" {
		return rule.jobShell
	}
	if rule.workflowShell != "" {
		return rule.workflowShell
	}
	if rule.runnerShell != "" {
		return rule.runnerShell
	}
	return "bash"
}

func (rule *RuleShellSyntax) checkShellSyntax(src, shell string, pos *Pos) {
	// Determine the shell variant for parsing
	var variant syntax.LangVariant
	if shell == "bash" || strings.HasPrefix(shell, "bash ") {
		variant = syntax.LangBash
	} else if shell == "sh" || strings.HasPrefix(shell, "sh ") {
		variant = syntax.LangPOSIX
	} else {
		// Skip checking for non-bash/sh shells (pwsh, python, etc.)
		return
	}

	// Sanitize expressions in script to avoid parse errors due to ${{ }}
	src = sanitizeExpressionsInScript(src)
	rule.Debug("%s: Checking shell syntax for %s script:\n%s", pos, shell, src)

	parser := syntax.NewParser(syntax.Variant(variant))
	_, err := parser.Parse(strings.NewReader(src), "")
	if err != nil {
		// Extract the error message without file position since we'll report our own position
		errMsg := err.Error()
		rule.Errorf(pos, "shell script syntax error: %s", errMsg)
	}
}
