package actionlint

import (
	"slices"
	"strings"
)

// Numeric permission levels for comparison. Higher means broader access.
const (
	permLevelNone  = 0
	permLevelRead  = 1
	permLevelWrite = 2
)

// permissionLevel returns the numeric level (0=none, 1=read, 2=write) for the given permission
// value. Unknown values fall back to none.
func permissionLevel(value string) int {
	switch strings.ToLower(value) {
	case "write":
		return permLevelWrite
	case "read":
		return permLevelRead
	}
	return permLevelNone
}

// permissionLevelName returns the canonical name for the given numeric level.
func permissionLevelName(level int) string {
	switch level {
	case permLevelWrite:
		return "write"
	case permLevelRead:
		return "read"
	default:
		return "none"
	}
}

// effectiveLevel returns the granted level for a scope under an explicit (non-nil) caller
// permissions block. "read-all"/"write-all" wins (clamped by the scope's allowed levels);
// otherwise the per-scope mapping is consulted, and a scope not present is "none" per GitHub
// semantics (declaring permissions: opts every scope out unless listed).
func effectiveLevel(p *ReusableWorkflowPermissions, scope string) int {
	if p.All != "" {
		switch p.All {
		case "write-all":
			return clampLevelForScope(scope, permLevelWrite)
		case "read-all":
			return clampLevelForScope(scope, permLevelRead)
		}
	}
	if v, ok := p.Scopes[scope]; ok {
		return clampLevelForScope(scope, permissionLevel(v))
	}
	// Explicit permissions block but this scope omitted → none.
	return permLevelNone
}

// silentDefaultLevel returns the level GitHub grants for the given scope when the caller declares
// no `permissions:` block at all (workflow- or job-level). Behavior depends on the configured
// assumption: "restricted" matches GitHub's restricted default token (contents/packages: read,
// everything else: none); "permissive" matches the permissive default (write on every scope) with
// the exception of `id-token`, which always requires an explicit opt-in regardless of the
// repo-level Workflow permissions setting.
func silentDefaultLevel(mode, scope string) int {
	if mode == AssumeDefaultPermissionsPermissive {
		if scope == "id-token" {
			return permLevelNone
		}
		return clampLevelForScope(scope, permLevelWrite)
	}
	switch scope {
	case "contents", "packages":
		return permLevelRead
	}
	return permLevelNone
}

// clampLevelForScope drops a level down to the highest level the scope supports.
// e.g. read-all + id-token → none (id-token only allows write/none), models only read/none.
func clampLevelForScope(scope string, level int) int {
	allowed, ok := allPermissionScopes[scope]
	if !ok {
		return level
	}
	if level == permLevelWrite && !slices.Contains(allowed, "write") {
		level = permLevelRead
	}
	if level == permLevelRead && !slices.Contains(allowed, "read") {
		level = permLevelNone
	}
	return level
}

var allPermissionScopes = map[string][]string{
	"actions":             {"read", "write", "none"},
	"artifact-metadata":   {"read", "write", "none"},
	"attestations":        {"read", "write", "none"},
	"checks":              {"read", "write", "none"},
	"contents":            {"read", "write", "none"},
	"deployments":         {"read", "write", "none"},
	"discussions":         {"read", "write", "none"},
	"id-token":            {"write", "none"},
	"issues":              {"read", "write", "none"},
	"models":              {"read", "none"},
	"packages":            {"read", "write", "none"},
	"pages":               {"read", "write", "none"},
	"pull-requests":       {"read", "write", "none"},
	"repository-projects": {"read", "write", "none"},
	"security-events":     {"read", "write", "none"},
	"statuses":            {"read", "write", "none"},
}

// RulePermissions is a rule checker to check permission configurations in a workflow.
// https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax#defining-access-for-the-github_token-scopes
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
			rule.Errorf(p.Value.Pos, "%q is invalid as permission of scope %q. available values are %s", p.Value.Value, n, quotes(s))
		}
	}
}
