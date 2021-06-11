package actionlint

import (
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

type shellcheckError struct {
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Level   string `json:"level"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// RuleShellcheck is a rule to check scripts using shellcheck.
// https://github.com/koalaman/shellcheck
type RuleShellcheck struct {
	RuleBase
	cmd           string
	workflowShell string
	jobShell      string
}

// NewRuleShellcheck craetes new RuleShellcheck instance. Parameter executable can be command name
// or relative/absolute file path. When the given executable is not found in system, it returns an
// error as 2nd return value.
func NewRuleShellcheck(executable string, debug io.Writer) (*RuleShellcheck, error) {
	p, err := exec.LookPath(executable)
	if err != nil {
		return nil, err
	}
	r := &RuleShellcheck{
		RuleBase:      RuleBase{name: "shellcheck", dbg: debug},
		cmd:           p,
		workflowShell: "",
		jobShell:      "",
	}
	return r, nil
}

// VisitStep is callback when visiting Step node.
func (rule *RuleShellcheck) VisitStep(n *Step) {
	if rule.cmd == "" {
		return
	}

	run, ok := n.Exec.(*ExecRun)
	if !ok || run.Run == nil {
		return
	}

	name := rule.getShellName(run)
	if name != "bash" && name != "sh" {
		return
	}

	rule.runShellcheck(run.Run.Value, name, run.RunPos)
}

// VisitJobPre is callback when visiting Job node before visiting its children.
func (rule *RuleShellcheck) VisitJobPre(n *Job) {
	if n.Defaults != nil && n.Defaults.Run != nil && n.Defaults.Run.Shell != nil {
		rule.jobShell = n.Defaults.Run.Shell.Value
		return
	}

	if n.RunsOn == nil {
		return
	}

	for _, label := range n.RunsOn.Labels {
		l := strings.ToLower(label.Value)
		// Default shell on Windows is PowerShell.
		// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#using-a-specific-shell
		if l == "windows" || strings.HasPrefix(l, "windows-") {
			return
		}
	}

	// TODO: When bash is not found, GitHub-hosted runner fallbacks to sh. What OSes require this behavior?
	rule.jobShell = "bash"
}

// VisitJobPost is callback when visiting Job node after visiting its children.
func (rule *RuleShellcheck) VisitJobPost(n *Job) {
	rule.jobShell = ""
}

// VisitWorkflowPre is callback when visiting Workflow node before visiting its children.
func (rule *RuleShellcheck) VisitWorkflowPre(n *Workflow) {
	if n.Defaults != nil && n.Defaults.Run != nil && n.Defaults.Run.Shell != nil {
		rule.workflowShell = n.Defaults.Run.Shell.Value
	}
}

// VisitWorkflowPost is callback when visiting Workflow node after visiting its children.
func (rule *RuleShellcheck) VisitWorkflowPost(n *Workflow) {
	rule.workflowShell = ""
}

func (rule *RuleShellcheck) getShellName(exec *ExecRun) string {
	if exec.Shell != nil {
		return exec.Shell.Value
	}
	if rule.jobShell != "" {
		return rule.jobShell
	}
	return rule.workflowShell
}

// Replace ${{ ... }} with spaces
func sanitizeExpressionsInScript(src string) string {
	for {
		s := strings.Index(src, "${{")
		if s == -1 {
			return src
		}
		e := strings.Index(src, "}}")
		if e == -1 || e < s {
			return src
		}
		e += 2 // offset for len("}}")
		// Note: If ${{ ... }} includes newline, line and column reported by shellcheck will be
		// shifted.
		src = src[:s] + strings.Repeat(" ", e-s) + src[e:]
	}
}

func (rule *RuleShellcheck) runShellcheck(src string, sh string, pos *Pos) {
	src = sanitizeExpressionsInScript(src)

	cmd := exec.Command(rule.cmd, "-f", "json", "-x", "--shell", sh, "-")
	cmd.Stderr = nil
	rule.debug("%s: Running shellcheck: %s", pos, cmd)

	// Use same options to run shell process described at document
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#using-a-specific-shell
	setup := "set -e"
	if sh == "bash" {
		setup = "set -eo pipefail"
	}
	script := fmt.Sprintf("%s\n%s\n", setup, src)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		rule.debug("%s: Could not make stdin pipe: %v", pos, err)
		return
	}
	if _, err := io.WriteString(stdin, script); err != nil {
		rule.debug("%s: Could not write stdin: %v", pos, err)
		return
	}
	stdin.Close()

	b, err := cmd.Output()
	if err != nil {
		rule.debug("Command %s failed: %v", rule.cmd, err)
	}
	if len(b) == 0 {
		return
	}

	errs := []shellcheckError{}
	if err := json.Unmarshal(b, &errs); err != nil {
		rule.debug("Could not unmarshal JSON input from shellcheck: %q: %v", b, err)
		return
	}

	// It's better to show source location in the script as position of error, but it's not
	// possible easily. YAML has multiple block styles with '|', '>', '|+', '>+', '|-', '>-'. Some
	// of them remove indentation and/or blank lines. So restoring source position in block string
	// is not possible. Sourcemap is necessary to do it.
	// Instead, actionlint shows position of 'run:' as position of error. And separately show
	// location in script which is reported by shellcheck in error message.
	for _, err := range errs {
		// Consider the first line is setup for running shell which was implicitly added for better check
		line := err.Line - 1
		msg := strings.TrimSuffix(err.Message, ".") // Trim period aligning style of error message
		rule.errorf(pos, "shellcheck reported issue in this script: SC%d:%s:%d:%d: %s", err.Code, err.Level, line, err.Column, msg)
	}
}
