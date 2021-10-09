package actionlint

import (
	"strings"
)

// RuleWorkflowCall is a rule checker to check workflow call at jobs.<job_id>.
type RuleWorkflowCall struct {
	RuleBase
}

// NewRuleWorkflowCall creates a new RuleWorkflowCall instance.
func NewRuleWorkflowCall() *RuleWorkflowCall {
	return &RuleWorkflowCall{
		RuleBase: RuleBase{name: "workflow-call"},
	}
}

// VisitJobPre is callback when visiting Job node before visiting its children.
func (rule *RuleWorkflowCall) VisitJobPre(n *Job) error {
	if n.WorkflowCall == nil {
		return nil
	}

	u := n.WorkflowCall.Uses
	if u == nil || u.Value == "" || strings.Contains(u.Value, "${{") {
		return nil
	}

	if !checkWorkflowCallUsesFormat(u.Value) {
		rule.errorf(u.Pos, "reusable workflow call %q at \"uses\" is not following the format \"owner/repo/path/to/workflow.yml@ref\". see https://docs.github.com/en/actions/learn-github-actions/reusing-workflows for more details", u.Value)
	}

	return nil
}

// Parse {owner}/{repo}/{path to workflow.yml}@{ref}
// https://docs.github.com/en/actions/learn-github-actions/reusing-workflows#calling-a-reusable-workflow
func checkWorkflowCallUsesFormat(u string) bool {
	if strings.HasPrefix(u, ".") {
		return false // Local path is not supported.
	}

	idx := strings.IndexRune(u, '/')
	if idx <= 0 {
		return false
	}
	u = u[idx+1:] // Eat owner

	idx = strings.IndexRune(u, '/')
	if idx <= 0 {
		return false
	}
	u = u[idx+1:] // Eat repo

	idx = strings.IndexRune(u, '@')
	if idx <= 0 {
		return false
	}
	u = u[idx+1:] // Eat workflow path

	return len(u) > 0
}
