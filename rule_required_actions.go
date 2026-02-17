// Package actionlint provides linting functionality for GitHub Actions workflows.
package actionlint

import (
	"fmt"
	"strings"
)

// RequiredActionRule represents a rule that enforces the usage of a specific GitHub Action
// with an optional version constraint in workflows.
type RequiredActionRule struct {
	Action  string `yaml:"Action"`  // Name of the required GitHub Action (e.g., "actions/checkout")
	Version string `yaml:"Version"` // Optional version constraint (e.g., "v3")
}

// RuleRequiredActions implements a linting rule that checks for the presence and version
// of required GitHub Actions within workflows.
type RuleRequiredActions struct {
	RuleBase
	required []RequiredActionRule
}

// NewRuleRequiredActions creates a new instance of RuleRequiredActions with the specified
// required actions. Returns nil if no required actions are provided.
func NewRuleRequiredActions(required []RequiredActionRule) *RuleRequiredActions {
	if len(required) == 0 {
		return nil
	}
	return &RuleRequiredActions{
		RuleBase: RuleBase{
			name: "required-actions",
			desc: "Checks that required GitHub Actions are used in workflows",
		},
		required: required,
	}
}

// VisitWorkflowPre analyzes the workflow to ensure all required actions are present
// with correct versions. It reports errors for missing or mismatched versions.
func (rule *RuleRequiredActions) VisitWorkflowPre(workflow *Workflow) error {
	if workflow == nil {
		return nil
	}

	pos := &Pos{Line: 1, Col: 1}
	foundActions := make(map[string]string)

	if workflow != nil && len(workflow.Jobs) > 0 {
		// Get first job's position
		for _, job := range workflow.Jobs {
			if job != nil && job.Pos != nil {
				pos = job.Pos
				break
			}
		}

		// Check steps in all jobs
		for _, job := range workflow.Jobs {
			if job == nil || len(job.Steps) == 0 {
				continue
			}
			for _, step := range job.Steps {
				if step != nil && step.Exec != nil {
					if exec, ok := step.Exec.(*ExecAction); ok && exec.Uses != nil {
						name, ver := parseActionRef(exec.Uses.Value)
						if name != "" {
							foundActions[name] = ver
						}
					}
				}
			}
		}
	}

	// Check required actions
	for _, req := range rule.required {
		ver, found := foundActions[req.Action]
		if !found {
			rule.Error(pos, fmt.Sprintf("required action %q (version %q) is not used in this workflow",
				req.Action, req.Version))
			continue
		}
		if req.Version != "" && ver != req.Version {
			rule.Error(pos, fmt.Sprintf("action %q must use version %q but found version %q",
				req.Action, req.Version, ver))
		}
	}

	return nil
}

// parseActionRef extracts the action name and version from a GitHub Action reference.
// Returns empty strings for invalid references like Docker images or malformed strings.
// Example: "actions/checkout@v3" returns ("actions/checkout", "v3")
func parseActionRef(uses string) (name string, version string) {
	if uses == "" || !strings.Contains(uses, "/") || strings.HasPrefix(uses, "docker://") {
		return "", ""
	}
	parts := strings.SplitN(uses, "@", 2)
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}
