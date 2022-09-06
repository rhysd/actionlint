package actionlint

import (
	"fmt"
	"strings"
)

// RuleWorkflowCall is a rule checker to check workflow call at jobs.<job_id>.
type RuleWorkflowCall struct {
	RuleBase
	workflowCallEventPos *Pos
	workflowPath         string
	cache                *LocalReusableWorkflowCache
}

// NewRuleWorkflowCall creates a new RuleWorkflowCall instance. 'workflowPath' is a file path to
// the workflow which is relative to a project root directory or an absolute path.
func NewRuleWorkflowCall(workflowPath string, cache *LocalReusableWorkflowCache) *RuleWorkflowCall {
	return &RuleWorkflowCall{
		RuleBase:             RuleBase{name: "workflow-call"},
		workflowCallEventPos: nil,
		workflowPath:         workflowPath,
		cache:                cache,
	}
}

// VisitWorkflowPre is callback when visiting Workflow node before visiting its children.
func (rule *RuleWorkflowCall) VisitWorkflowPre(n *Workflow) error {
	for _, e := range n.On {
		if e, ok := e.(*WorkflowCallEvent); ok {
			rule.workflowCallEventPos = e.Pos
			// Register this reusable workflow in cache so that it does not need to parse this workflow
			// file again when this workflow is called by other workflows.
			rule.cache.WriteWorkflowCallEvent(rule.workflowPath, e)
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

	if strings.Contains(u.Value, "${{") {
		return nil
	}

	if isWorkflowCallUsesLocalFormat(u.Value) {
		rule.checkWorkflowCallUsesLocal(n.WorkflowCall)
		return nil
	}

	if isWorkflowCallUsesRepoFormat(u.Value) {
		return nil
	}

	rule.errorf(
		u.Pos,
		"reusable workflow call %q at \"uses\" is not following the format \"owner/repo/path/to/workflow.yml@ref\" nor \"./path/to/workflow.yml\". see https://docs.github.com/en/actions/learn-github-actions/reusing-workflows for more details",
		u.Value,
	)
	return nil
}

func (rule *RuleWorkflowCall) checkWorkflowCallUsesLocal(call *WorkflowCall) {
	u := call.Uses
	m, err := rule.cache.FindMetadata(u.Value)
	if err != nil {
		rule.errorf(u.Pos, "error while checking reusable workflow call %q: %s", u.Value, err.Error())
		return
	}
	if m == nil {
		rule.debug("Skip workflow call %q since no metadata was found", u.Value)
		return
	}

	// Validate inputs
	for n, i := range m.Inputs {
		if i.Required {
			if _, ok := call.Inputs[n]; !ok {
				rule.errorf(u.Pos, "input %q is required by %q reusable workflow", n, u.Value)
			}
		}
	}
	for n, i := range call.Inputs {
		if _, ok := m.Inputs[n]; !ok {
			note := "no input is defined"
			if len(m.Inputs) > 0 {
				i := make([]string, 0, len(m.Inputs))
				for n := range m.Inputs {
					i = append(i, n)
				}
				if len(i) == 1 {
					note = fmt.Sprintf("defined input is %q", i[0])
				} else {
					note = "defined inputs are " + sortedQuotes(i)
				}
			}
			rule.errorf(i.Name.Pos, "input %q is not defined in %q reusable workflow. %s", n, u.Value, note)
		}
	}

	// Validate secrets
	if !call.InheritSecrets {
		for n, r := range m.Secrets {
			if r {
				if _, ok := call.Secrets[n]; !ok {
					rule.errorf(u.Pos, "secret %q is required by %q reusable workflow", n, u.Value)
				}
			}
		}
		for n, s := range call.Secrets {
			if _, ok := m.Secrets[n]; !ok {
				note := "no secret is defined"
				if len(m.Secrets) > 0 {
					s := make([]string, 0, len(m.Secrets))
					for n := range m.Secrets {
						s = append(s, n)
					}
					if len(s) == 1 {
						note = fmt.Sprintf("defined secret is %q", s[0])
					} else {
						note = "defined secrets are " + sortedQuotes(s)
					}
				}
				rule.errorf(s.Name.Pos, "secret %q is not defined in %q reusable workflow. %s", n, u.Value, note)
			}
		}
	}

	rule.debug("Validated reusable workflow %q", u.Value)
}

// Parse ./{path/{filename}
// https://docs.github.com/en/actions/learn-github-actions/reusing-workflows#calling-a-reusable-workflow
func isWorkflowCallUsesLocalFormat(u string) bool {
	if !strings.HasPrefix(u, "./") {
		return false
	}
	u = strings.TrimPrefix(u, "./")

	// Cannot container a ref
	idx := strings.IndexRune(u, '@')
	if idx > 0 {
		return false
	}

	return len(u) > 0
}

// Parse {owner}/{repo}/{path to workflow.yml}@{ref}
// https://docs.github.com/en/actions/learn-github-actions/reusing-workflows#calling-a-reusable-workflow
func isWorkflowCallUsesRepoFormat(u string) bool {
	// Repo reference must start with owner
	if strings.HasPrefix(u, ".") {
		return false
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
