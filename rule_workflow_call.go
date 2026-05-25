package actionlint

import (
	"fmt"
	"sort"
	"strings"
)

// RuleWorkflowCall is a rule checker to check workflow call at jobs.<job_id>.
type RuleWorkflowCall struct {
	RuleBase
	workflowCallEventPos *Pos
	workflowPath         string
	cache                *LocalReusableWorkflowCache
	workflow             *Workflow
	curJob               *Job
}

// NewRuleWorkflowCall creates a new RuleWorkflowCall instance. 'workflowPath' is a file path to
// the workflow which is relative to a project root directory or an absolute path.
func NewRuleWorkflowCall(workflowPath string, cache *LocalReusableWorkflowCache) *RuleWorkflowCall {
	return &RuleWorkflowCall{
		RuleBase: RuleBase{
			name: "workflow-call",
			desc: "Checks for reusable workflow calls. Inputs and outputs of called reusable workflow are checked",
		},
		workflowCallEventPos: nil,
		workflowPath:         workflowPath,
		cache:                cache,
	}
}

// VisitWorkflowPre is callback when visiting Workflow node before visiting its children.
func (rule *RuleWorkflowCall) VisitWorkflowPre(n *Workflow) error {
	rule.workflow = n
	for _, e := range n.On {
		if e, ok := e.(*WorkflowCallEvent); ok {
			rule.workflowCallEventPos = e.Pos
			// Register this reusable workflow in cache so that it does not need to parse this workflow
			// file again when this workflow is called by other workflows.
			rule.cache.WriteWorkflowCallEvent(rule.workflowPath, e, n)
			break
		}
	}
	return nil
}

// VisitJobPre is callback when visiting Job node before visiting its children.
func (rule *RuleWorkflowCall) VisitJobPre(n *Job) error {
	rule.curJob = n
	if n.WorkflowCall == nil {
		return nil
	}

	u := n.WorkflowCall.Uses
	if u == nil || u.Value == "" || u.ContainsExpression() {
		return nil
	}

	if isWorkflowCallUsesLocalFormat(u.Value) {
		rule.checkWorkflowCallUsesLocal(n.WorkflowCall)
		return nil
	}

	if isWorkflowCallUsesRepoFormat(u.Value) {
		return nil
	}

	if strings.HasPrefix(u.Value, "./") {
		// When the specification is invalid and it is local reusable workflow call, remember it caused
		// an error by setting `nil` to cache. This can prevent redundant 'could not read workflow call'
		// error.
		rule.cache.writeCache(u.Value, nil)
	}

	rule.Errorf(
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
		rule.Error(u.Pos, err.Error())
		return
	}
	if m == nil {
		rule.Debug("Skip workflow call %q since no metadata was found", u.Value)
		return
	}

	// Validate inputs
	for n, i := range m.Inputs {
		if i != nil && i.Required {
			if _, ok := call.Inputs[n]; !ok {
				rule.Errorf(u.Pos, "input %q is required by %q reusable workflow", i.Name, u.Value)
			}
		}
	}
	for n, i := range call.Inputs {
		if _, ok := m.Inputs[n]; !ok {
			note := "no input is defined"
			if len(m.Inputs) > 0 {
				is := make([]string, 0, len(m.Inputs))
				for _, i := range m.Inputs {
					is = append(is, i.Name)
				}
				if len(is) == 1 {
					note = fmt.Sprintf("defined input is %q", is[0])
				} else {
					note = "defined inputs are " + sortedQuotes(is)
				}
			}
			rule.Errorf(i.Name.Pos, "input %q is not defined in %q reusable workflow. %s", i.Name.Value, u.Value, note)
		}
	}

	// Validate secrets
	if !call.InheritSecrets {
		for n, s := range m.Secrets {
			if s.Required {
				if _, ok := call.Secrets[n]; !ok {
					rule.Errorf(u.Pos, "secret %q is required by %q reusable workflow", s.Name, u.Value)
				}
			}
		}
		for n, s := range call.Secrets {
			if _, ok := m.Secrets[n]; !ok {
				note := "no secret is defined"
				if len(m.Secrets) > 0 {
					ss := make([]string, 0, len(m.Secrets))
					for _, s := range m.Secrets {
						ss = append(ss, s.Name)
					}
					if len(ss) == 1 {
						note = fmt.Sprintf("defined secret is %q", ss[0])
					} else {
						note = "defined secrets are " + sortedQuotes(ss)
					}
				}
				rule.Errorf(s.Name.Pos, "secret %q is not defined in %q reusable workflow. %s", s.Name.Value, u.Value, note)
			}
		}
	}

	// Validate permissions
	rule.checkWorkflowCallPermissions(call, m)

	rule.Debug("Validated reusable workflow %q", u.Value)
}

// checkWorkflowCallPermissions compares each callee job's effective `permissions:` requirement
// against the caller job's effective grant. Emits an error per missing scope. The check ignores
// `if:` on callee jobs because GitHub validates permissions at workflow load time regardless of
// the runtime gate.
func (rule *RuleWorkflowCall) checkWorkflowCallPermissions(call *WorkflowCall, m *ReusableWorkflowMetadata) {
	if len(m.JobPermissions) == 0 {
		return
	}

	cfg := rule.Config()
	mode := AssumeDefaultPermissionsRestricted
	if cfg != nil && cfg.AssumeDefaultPermissions != nil {
		mode = *cfg.AssumeDefaultPermissions
	}

	// Caller's effective permissions: job-level wins over workflow-level.
	var callerPerm *ReusableWorkflowPermissions
	callerSilent := false
	if rule.curJob != nil && rule.curJob.Permissions != nil {
		callerPerm = convertASTPermissions(rule.curJob.Permissions)
	} else if rule.workflow != nil && rule.workflow.Permissions != nil {
		callerPerm = convertASTPermissions(rule.workflow.Permissions)
	} else {
		callerSilent = true
		if mode == AssumeDefaultPermissionsPermissive {
			// Caller is silent and the configured assumption is permissive — skip the check
			// entirely. GitHub's permissive default grants everything (except id-token, which
			// always needs an explicit opt-in), so we have nothing to flag.
			return
		}
	}

	defaultLevel := func(scope string) int {
		// Used for scopes a *non-nil* caller permission doesn't mention (treated as 'none' per GitHub),
		// or — when the caller is fully silent and mode is restricted — the GitHub restricted default.
		if !callerSilent {
			return permLevelNone
		}
		// Restricted default: only contents and packages are granted at read.
		// https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/enabling-features-for-your-repository/managing-github-actions-settings-for-a-repository
		switch scope {
		case "contents", "packages":
			return permLevelRead
		}
		return permLevelNone
	}

	// Iterate callee jobs in a stable order so error messages are deterministic.
	ids := make([]string, 0, len(m.JobPermissions))
	for id := range m.JobPermissions {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	u := call.Uses
	for _, id := range ids {
		jp := m.JobPermissions[id]
		if jp == nil {
			continue
		}
		// Determine the set of scopes the callee actually requires (non-none).
		required := map[string]int{}
		if jp.All != "" {
			level := permLevelNone
			switch jp.All {
			case "write-all":
				level = permLevelWrite
			case "read-all":
				level = permLevelRead
			}
			if level > permLevelNone {
				for scope := range allPermissionScopes {
					l := clampLevelForScope(scope, level)
					if l > permLevelNone {
						required[scope] = l
					}
				}
			}
		} else {
			for scope, val := range jp.Scopes {
				if _, ok := allPermissionScopes[scope]; !ok {
					continue
				}
				l := permissionLevel(scope, val)
				if l > permLevelNone {
					required[scope] = l
				}
			}
		}

		scopes := make([]string, 0, len(required))
		for s := range required {
			scopes = append(scopes, s)
		}
		sort.Strings(scopes)
		for _, scope := range scopes {
			want := required[scope]
			have := effectiveLevel(callerPerm, scope, defaultLevel(scope))
			if have < want {
				rule.Errorf(
					u.Pos,
					"nested job %q of %q requires %q but the calling job grants %q",
					id, u.Value,
					scope+": "+permissionLevelName(want),
					scope+": "+permissionLevelName(have),
				)
			}
		}
	}
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
