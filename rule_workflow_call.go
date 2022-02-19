package actionlint

import (
	"strings"
)

// RuleWorkflowCall is a rule checker to check workflow call at jobs.<job_id>.
type RuleWorkflowCall struct {
	RuleBase
	workflowCallEventPos *Pos
}

// NewRuleWorkflowCall creates a new RuleWorkflowCall instance.
func NewRuleWorkflowCall() *RuleWorkflowCall {
	return &RuleWorkflowCall{
		RuleBase:             RuleBase{name: "workflow-call"},
		workflowCallEventPos: nil,
	}
}

// VisitWorkflowPre is callback when visiting Workflow node before visiting its children.
func (rule *RuleWorkflowCall) VisitWorkflowPre(n *Workflow) error {
	for _, e := range n.On {
		if e, ok := e.(*WorkflowCallEvent); ok {
			rule.workflowCallEventPos = e.Pos
			break
		}
	}
	return nil
}

// VisitJobPre is callback when visiting Job node before visiting its children.
func (rule *RuleWorkflowCall) VisitJobPre(n *Job) error {
	if n.WorkflowCall == nil {
		return nil
	}

	u := n.WorkflowCall.Uses
	if u == nil || u.Value == "" {
		return nil
	}

	if rule.workflowCallEventPos != nil {
		rule.errorf(u.Pos, "reusable workflow cannot be nested. but this workflow hooks \"workflow_call\" event at %s", rule.workflowCallEventPos)
	}

	if !strings.Contains(u.Value, "${{") && !checkWorkflowCallUsesFormat(u.Value) {
		rule.errorf(u.Pos, "reusable workflow call %q at \"uses\" is not following the format \"owner/repo/path/to/workflow.yml@ref\". see https://docs.github.com/en/actions/learn-github-actions/reusing-workflows for more details", u.Value)
	}

	return nil
}

// Parse {owner}/{repo}/{path to workflow.yml}@{ref} or ./{path to workflow.yml}
// https://docs.github.com/en/actions/learn-github-actions/reusing-workflows#calling-a-reusable-workflow
func checkWorkflowCallUsesFormat(u string) bool {
	if strings.HasPrefix(u, ".") {
		return true // Local paths have no known path components to validate.
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
