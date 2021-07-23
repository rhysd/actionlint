package actionlint

//go:generate go run ./scripts/generate-popular-actions -s remote -f go ./popular_actions.go

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// RuleAction is a rule to check running action in steps of jobs.
// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idstepsuses
type RuleAction struct {
	RuleBase
	repoPath string
}

// ActionMetadataInput is input metadata of action. Description is omitted because it is unused by
// actionlint.
// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#inputs
type ActionMetadataInput struct {
	// Required is whether the input is required.
	Required bool `yaml:"required" json:"required"`
	// Default is a default value of the input. This is optional field. nil is set when it is
	// missing.
	Default *string `yaml:"default" json:"default"`
}

// IsRequired returns whether this input is required
func (input *ActionMetadataInput) IsRequired() bool {
	return input.Required && input.Default == nil
}

// ActionMetadata represents structure of action.yaml.
// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions
type ActionMetadata struct {
	// Name is "name" field of action.yaml
	Name string `yaml:"name" json:"name"`
	// Inputs is "inputs" field of action.yaml
	Inputs map[string]*ActionMetadataInput `yaml:"inputs" json:"inputs"`
	// Outputs is "outputs" field of action.yaml. Key is name of output. Description is omitted
	// since actionlint does not use it.
	Outputs map[string]struct{} `yaml:"outputs" json:"outputs"`
}

// NewRuleAction creates new RuleAction instance.
func NewRuleAction(repoDir string) *RuleAction {
	return &RuleAction{
		RuleBase: RuleBase{name: "action"},
		repoPath: repoDir,
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
		return
	}

	rule.checkAction(meta, exec, func(m *ActionMetadata) string {
		return strconv.Quote(spec)
	})
}

func (rule *RuleAction) invalidActionFormat(pos *Pos, spec string, why string) {
	rule.errorf(pos, "specifying action %q in invalid format because %s. available formats are \"{owner}/{repo}@{ref}\" or \"{owner}/{repo}/{path}@{ref}\"", spec, why)
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#example-using-the-github-packages-container-registry
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

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#example-using-action-in-the-same-repository-as-the-workflow
func (rule *RuleAction) checkLocalAction(path string, action *ExecAction) {
	if rule.repoPath == "" {
		return // Give up
	}

	dir := filepath.Join(rule.repoPath, filepath.FromSlash(path))
	b := rule.readActionMetadataFile(dir, action.Uses.Pos)
	if len(b) == 0 {
		return
	}

	var meta ActionMetadata
	if err := yaml.Unmarshal(b, &meta); err != nil {
		rule.errorf(action.Uses.Pos, "action.yml in %q is invalid: %s", dir, err.Error())
		return
	}

	rule.checkAction(&meta, action, func(m *ActionMetadata) string {
		return fmt.Sprintf("%q defined at %q", meta.Name, path)
	})
}

func (rule *RuleAction) checkAction(meta *ActionMetadata, exec *ExecAction, describe func(*ActionMetadata) string) {
	// Check specified inputs are defined in action's inputs spec
	for name, val := range exec.Inputs {
		if _, ok := meta.Inputs[name]; !ok {
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
	}

	// Check mandatory inputs are specified
	for name, input := range meta.Inputs {
		if input.IsRequired() {
			if _, ok := exec.Inputs[name]; !ok {
				ns := make([]string, 0, len(meta.Inputs))
				for n, i := range meta.Inputs {
					if i.IsRequired() {
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

func (rule *RuleAction) readActionMetadataFile(dir string, pos *Pos) []byte {
	for _, p := range []string{
		filepath.Join(dir, "action.yaml"),
		filepath.Join(dir, "action.yml"),
	} {
		if b, err := ioutil.ReadFile(p); err == nil {
			return b
		}
	}

	if wd, err := os.Getwd(); err == nil {
		if p, err := filepath.Rel(wd, dir); err == nil {
			dir = p
		}
	}
	rule.errorf(pos, "Neither action.yaml nor action.yml is found in directory \"%s\"", dir)
	return nil
}
