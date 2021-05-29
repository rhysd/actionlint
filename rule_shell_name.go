package actionlint

import (
	"strconv"
	"strings"
)

type platformKind int

const (
	platformKindAny platformKind = iota
	platformKindMacOrLinux
	platformKindWindows
)

// RuleShellName is a rule to check 'shell' field. For more details, see
// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#using-a-specific-shell
type RuleShellName struct {
	RuleBase
	platform platformKind
}

// NewRuleShellName creates new RuleShellName instance.
func NewRuleShellName() *RuleShellName {
	return &RuleShellName{
		RuleBase: RuleBase{name: "shell-name"},
		platform: platformKindAny,
	}
}

func (rule *RuleShellName) VisitStep(n *Step) {
	if run, ok := n.Exec.(*ExecRun); ok {
		rule.checkShellName(run.Shell)
	}
}

func (rule *RuleShellName) VisitJobPre(n *Job) {
	if n.RunsOn == nil {
		return
	}
	rule.platform = getPlatformFromRunner(n.RunsOn)
	if n.Defaults != nil && n.Defaults.Run != nil {
		rule.checkShellName(n.Defaults.Run.Shell)
	}
}

func (rule *RuleShellName) VisitJobPost(n *Job) {
	rule.platform = platformKindAny // Clear
}

func (rule *RuleShellName) VisitWorkflowPre(n *Workflow) {
	if n.Defaults != nil && n.Defaults.Run != nil {
		rule.checkShellName(n.Defaults.Run.Shell)
	}
}

func (rule *RuleShellName) checkShellName(name *String) {
	if name == nil {
		return
	}

	available := getAvailableShellNames(rule.platform)

	for _, s := range available {
		if name.Value == s {
			return // ok
		}
	}

	onPlatform := ""
	switch rule.platform {
	case platformKindWindows:
		for _, p := range getAvailableShellNames(platformKindAny) {
			if name.Value == p {
				onPlatform = " on Windows" // only when the shell is unavailable on Windows
			}
		}
	case platformKindMacOrLinux:
		for _, p := range getAvailableShellNames(platformKindAny) {
			if name.Value == p {
				onPlatform = " on macOS or Linux" // only when the shell is unavailable on macOS or Linux
			}
		}
	}

	qs := make([]string, 0, len(available))
	for _, s := range available {
		qs = append(qs, strconv.Quote(s))
	}

	rule.errorf(
		name.Pos,
		"shell name %q is invalid%s. available names are %s",
		name.Value,
		onPlatform,
		strings.Join(qs, ", "),
	)
}

func getAvailableShellNames(kind platformKind) []string {
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#using-a-specific-shell
	switch kind {
	case platformKindAny:
		return []string{
			"bash",
			"pwsh",
			"python",
			"sh",
			"cmd",
			"powershell",
		}
	case platformKindWindows:
		return []string{
			"bash",
			"pwsh",
			"python",
			"cmd",
			"powershell",
		}
	case platformKindMacOrLinux:
		return []string{
			"bash",
			"pwsh",
			"python",
			"sh",
		}
	default:
		panic("unreachable")
	}
}

func getPlatformFromRunner(runner Runner) platformKind {
	switch r := runner.(type) {
	case *GitHubHostedRunner:
		if r.Label == nil {
			return platformKindAny
		}
		if strings.HasPrefix(r.Label.Value, "windows-") {
			return platformKindWindows
		} else if strings.HasPrefix(r.Label.Value, "macos-") || strings.HasPrefix(r.Label.Value, "ubuntu-") {
			return platformKindMacOrLinux
		} else {
			return platformKindAny
		}
	case *SelfHostedRunner:
		if len(r.Labels) == 0 {
			return platformKindAny
		}
		// https://docs.github.com/en/actions/hosting-your-own-runners/using-self-hosted-runners-in-a-workflow#using-default-labels-to-route-jobs
		switch r.Labels[0].Value {
		case "windows":
			return platformKindWindows
		case "linux", "macOS", "macos":
			return platformKindMacOrLinux
		default:
			return platformKindAny
		}
	default:
		panic("unreachable")
	}
}
