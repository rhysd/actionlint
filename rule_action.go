package actionlint

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// RuleAction is a rule to check running action in steps of jobs.
// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idstepsuses
type RuleAction struct {
	RuleBase
	cache *LocalActionsCache
}

// NewRuleAction creates new RuleAction instance.
func NewRuleAction(cache *LocalActionsCache) *RuleAction {
	return &RuleAction{
		RuleBase: RuleBase{name: "action"},
		cache:    cache,
	}
}

// VisitStep is callback when visiting Step node.
func (rule *RuleAction) VisitStep(n *Step) error {
	e, ok := n.Exec.(*ExecAction)
	if !ok || e.Uses == nil {
		return nil
	}

	spec := e.Uses.Value

	if strings.Contains(spec, "${{") && strings.Contains(spec, "}}") {
		// Cannot parse specification made with interpolation. Give up
		return nil
	}

	if strings.HasPrefix(spec, "./") {
		// Relative to repository root
		rule.checkLocalAction(spec, e)
		return nil
	}

	if strings.HasPrefix(spec, "docker://") {
		rule.checkDockerAction(spec, e)
		return nil
	}

	rule.checkRepoAction(spec, e)
	return nil
}

// Parse {owner}/{repo}@{ref} or {owner}/{repo}/{path}@{ref}
func (rule *RuleAction) checkRepoAction(spec string, exec *ExecAction) {
	s := spec
	idx := strings.IndexRune(s, '@')
	if idx == -1 {
		rule.invalidActionFormat(exec.Uses.Pos, spec, "ref is missing")
		return
	}
	ref := s[idx+1:]
	s = s[:idx] // remove {ref}

	idx = strings.IndexRune(s, '/')
	if idx == -1 {
		rule.invalidActionFormat(exec.Uses.Pos, spec, "owner is missing")
		return
	}

	owner := s[:idx]
	s = s[idx+1:] // eat {owner}

	repo := s
	if idx := strings.IndexRune(s, '/'); idx >= 0 {
		repo = s[:idx]
		// path = s[idx+1:]
	}

	if owner == "" || repo == "" || ref == "" {
		rule.invalidActionFormat(exec.Uses.Pos, spec, "owner and repo and ref should not be empty")
	}

	meta, ok := PopularActions[spec]
	if !ok {
		rule.debug("This action is not found in popular actions data set: %s", spec)
		return
	}
	if meta.SkipInputs {
		rule.debug("This action skips to check inputs: %s", spec)
		return
	}

	rule.checkAction(meta, exec, func(m *ActionMetadata) string {
		return strconv.Quote(spec)
	})
}

func (rule *RuleAction) invalidActionFormat(pos *Pos, spec string, why string) {
	rule.errorf(pos, "specifying action %q in invalid format because %s. available formats are \"{owner}/{repo}@{ref}\" or \"{owner}/{repo}/{path}@{ref}\"", spec, why)
}

// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#example-using-the-github-packages-container-registry
func (rule *RuleAction) checkDockerAction(uri string, exec *ExecAction) {
	tag := ""
	tagExists := false
	if idx := strings.IndexRune(uri[len("docker://"):], ':'); idx != -1 {
		idx += len("docker://")
		if idx < len(uri) {
			tag = uri[idx+1:]
			uri = uri[:idx]
			tagExists = true
		}
	}

	if _, err := url.Parse(uri); err != nil {
		rule.errorf(
			exec.Uses.Pos,
			"URI for Docker container %q is invalid: %s (tag=%s)",
			uri,
			err.Error(),
			tag,
		)
	}

	if tagExists && tag == "" {
		rule.errorf(exec.Uses.Pos, "tag of Docker action should not be empty: %q", uri)
	}
}

// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#example-using-action-in-the-same-repository-as-the-workflow
func (rule *RuleAction) checkLocalAction(path string, action *ExecAction) {
	meta, err := rule.cache.FindMetadata(path)
	if err != nil {
		rule.errorf(action.Uses.Pos, "error while parsing local action metadata: %s", err)
		return
	}
	if meta == nil {
		return
	}

	rule.checkAction(meta, action, func(m *ActionMetadata) string {
		return fmt.Sprintf("%q defined at %q", meta.Name, path)
	})
}

func (rule *RuleAction) checkAction(meta *ActionMetadata, exec *ExecAction, describe func(*ActionMetadata) string) {
	// Check specified inputs are defined in action's inputs spec
Outer:
	for name, val := range exec.Inputs {
		// XXX: This is O(n) workaround for #31
		for n := range meta.Inputs {
			// Input name is in lower case because parser.go converts all keys into lower case.
			// But keys of ActionMetadata.Inputs are as-is defined in action.yml.
			if strings.EqualFold(n, name) {
				continue Outer // found
			}
		}
		ns := make([]string, 0, len(meta.Inputs))
		for n := range meta.Inputs {
			ns = append(ns, n)
		}
		rule.errorf(
			val.Name.Pos,
			"input %q is not defined in action %s. available inputs are %s",
			name,
			describe(meta),
			sortedQuotes(ns),
		)
	}

	// Check mandatory inputs are specified
	for name, required := range meta.Inputs {
		if required {
			if _, ok := exec.Inputs[strings.ToLower(name)]; !ok {
				ns := make([]string, 0, len(meta.Inputs))
				for n, req := range meta.Inputs {
					if req {
						ns = append(ns, n)
					}
				}
				rule.errorf(
					exec.Uses.Pos,
					"missing input %q which is required by action %s. all required inputs are %s",
					name,
					describe(meta),
					sortedQuotes(ns),
				)
			}
		}
	}
}
