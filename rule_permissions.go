package actionlint

import (
	"slices"
)

var allPermissionScopes = map[string][]string{
	"actions":             []string{"read", "write", "none"},
	"attestations":        []string{"read", "write", "none"},
	"checks":              []string{"read", "write", "none"},
	"contents":            []string{"read", "write", "none"},
	"deployments":         []string{"read", "write", "none"},
	"discussions":         []string{"read", "write", "none"},
	"id-token":            []string{"write", "none"},
	"issues":              []string{"read", "write", "none"},
	"models":              []string{"read", "none"},
	"packages":            []string{"read", "write", "none"},
	"pages":               []string{"read", "write", "none"},
	"pull-requests":       []string{"read", "write", "none"},
	"repository-projects": []string{"read", "write", "none"},
	"security-events":     []string{"read", "write", "none"},
	"statuses":            []string{"read", "write", "none"},
}

// RulePermissions is a rule checker to check permission configurations in a workflow.
// https://docs.github.com/en/actions/security-for-github-actions/security-guides/automatic-token-authentication#permissions-for-the-github_token
type RulePermissions struct {
	RuleBase
}

// NewRulePermissions creates new RulePermissions instance.
func NewRulePermissions() *RulePermissions {
	return &RulePermissions{
		RuleBase: RuleBase{
			name: "permissions",
			desc: "Checks for permissions configuration in \"permissions:\". Permission names and permission scopes are checked",
		},
	}
}

// VisitJobPre is callback when visiting Job node before visiting its children.
func (rule *RulePermissions) VisitJobPre(n *Job) error {
	rule.checkPermissions(n.Permissions)
	return nil
}

// VisitWorkflowPre is callback when visiting Workflow node before visiting its children.
func (rule *RulePermissions) VisitWorkflowPre(n *Workflow) error {
	rule.checkPermissions(n.Permissions)
	return nil
}

func (rule *RulePermissions) checkPermissions(p *Permissions) {
	if p == nil {
		return
	}

	if p.All != nil {
		switch p.All.Value {
		case "write-all", "read-all":
			// OK
		default:
			rule.Errorf(p.All.Pos, "%q is invalid for permission for all the scopes. available values are \"read-all\", \"write-all\" or {}", p.All.Value)
		}
		return
	}

	for _, p := range p.Scopes {
		n := p.Name.Value // Permission names are case-sensitive
		s, ok := allPermissionScopes[n]
		if !ok {
			ss := make([]string, 0, len(allPermissionScopes))
			for s := range allPermissionScopes {
				ss = append(ss, s)
			}
			rule.Errorf(p.Name.Pos, "unknown permission scope %q. all available permission scopes are %s", n, sortedQuotes(ss))
			continue
		}

		if !slices.Contains(s, p.Value.Value) {
			rule.Errorf(p.Value.Pos, "%q is invalid as scope of permission %q. available values are %s", p.Value.Value, n, quotes(s))
		}
	}
}
