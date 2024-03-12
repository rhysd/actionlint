package actionlint

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// MinimumNodeRunnerVersion is the minimum supported Node.js version for JavaScript action runner.
// This constant will be updated when GitHub bumps the version.
// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#runs-for-javascript-actions
//
// Note: "node16" runner is deprecated but still available: https://github.blog/changelog/2023-09-22-github-actions-transitioning-from-node-16-to-node-20/
const MinimumNodeRunnerVersion uint8 = 16

// RuleAction is a rule to check running action in steps of jobs.
// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idstepsuses
type RuleAction struct {
	RuleBase
	cache *LocalActionsCache
}

// NewRuleAction creates new RuleAction instance.
func NewRuleAction(cache *LocalActionsCache) *RuleAction {
	return &RuleAction{
		RuleBase: RuleBase{
			name: "action",
			desc: "Checks for popular actions released on GitHub, local actions, and action calls at \"uses:\"",
		},
		cache: cache,
	}
}

// VisitStep is callback when visiting Step node.
func (rule *RuleAction) VisitStep(n *Step) error {
	e, ok := n.Exec.(*ExecAction)
	if !ok || e.Uses == nil {
		return nil
	}

	if e.Uses.ContainsExpression() {
		// Cannot parse specification made with interpolation. Give up
		return nil
	}

	spec := e.Uses.Value

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
	}

	if owner == "" || repo == "" || ref == "" {
		rule.invalidActionFormat(exec.Uses.Pos, spec, "owner and repo and ref should not be empty")
	}

	meta, ok := PopularActions[spec]
	if !ok {
		rule.Debug("This action is not found in popular actions data set: %s", spec)
		return
	}
	if meta.SkipInputs {
		rule.Debug("This action skips to check inputs: %s", spec)
		return
	}

	rule.checkAction(meta, exec, func(m *ActionMetadata) string {
		return strconv.Quote(spec)
	})
}

func (rule *RuleAction) invalidActionFormat(pos *Pos, spec string, why string) {
	rule.Errorf(pos, "specifying action %q in invalid format because %s. available formats are \"{owner}/{repo}@{ref}\" or \"{owner}/{repo}/{path}@{ref}\"", spec, why)
}

// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#runs-for-javascript-actions
// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#runs-for-docker-container-actions
// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#runs-for-composite-actions
func (rule *RuleAction) checkLocalActionRunner(path string, meta *ActionMetadata, pos *Pos) {
	if meta.Runs.Runner == nil {
		rule.Errorf(pos, `"runs.using" is missing in the local action %q defined at %q`, meta.Name, path)
		return
	}

	// TODO: Add more checks
	switch r := meta.Runs.Runner.(type) {
	case *ActionRunnerJS:
		if r.Version < MinimumNodeRunnerVersion {
			rule.Errorf(
				pos,
				`the Node.js version %d is too old. the minimum version is %d. the local action %q defined at %q is invalid. see https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#runs-for-javascript-actions`,
				r.Version,
				MinimumNodeRunnerVersion,
				meta.Name,
				path,
			)
		}
	}
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
		rule.Errorf(
			exec.Uses.Pos,
			"URI for Docker container %q is invalid: %s (tag=%s)",
			uri,
			err.Error(),
			tag,
		)
	}

	if tagExists && tag == "" {
		rule.Errorf(exec.Uses.Pos, "tag of Docker action should not be empty: %q", uri)
	}
}

// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#example-using-action-in-the-same-repository-as-the-workflow
func (rule *RuleAction) checkLocalAction(path string, action *ExecAction) {
	meta, cached, err := rule.cache.FindMetadata(path)
	if err != nil {
		rule.Error(action.Uses.Pos, err.Error())
		return
	}
	if meta == nil {
		return
	}

	if !cached {
		rule.Debug("Checking runner metadata of %s action %q at %q", meta.Runs, meta.Name, path)
		rule.checkLocalActionRunner(path, meta, action.Uses.Pos)
	}

	rule.checkAction(meta, action, func(m *ActionMetadata) string {
		return fmt.Sprintf("%q defined at %q", meta.Name, path)
	})
}

func (rule *RuleAction) checkAction(meta *ActionMetadata, exec *ExecAction, describe func(*ActionMetadata) string) {
	// Check specified inputs are defined in action's inputs spec
	for id, i := range exec.Inputs {
		if _, ok := meta.Inputs[id]; !ok {
			ns := make([]string, 0, len(meta.Inputs))
			for _, i := range meta.Inputs {
				ns = append(ns, i.Name)
			}
			rule.Errorf(
				i.Name.Pos,
				"input %q is not defined in action %s. available inputs are %s",
				i.Name.Value,
				describe(meta),
				sortedQuotes(ns),
			)
		}
	}

	// Check mandatory inputs are specified
	for id, i := range meta.Inputs {
		if i.Required {
			if _, ok := exec.Inputs[id]; !ok {
				ns := make([]string, 0, len(meta.Inputs))
				for _, i := range meta.Inputs {
					if i.Required {
						ns = append(ns, i.Name)
					}
				}
				rule.Errorf(
					exec.Uses.Pos,
					"missing input %q which is required by action %s. all required inputs are %s",
					i.Name,
					describe(meta),
					sortedQuotes(ns),
				)
			}
		}
	}
}
