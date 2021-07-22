package actionlint

import (
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// RuleAction is a rule to check running action in steps of jobs.
// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idstepsuses
type RuleAction struct {
	RuleBase
	repoPath string
}

// ActionInput is input metadata of action.
// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#inputs
type ActionInput struct {
	// Required is whether the input is required.
	Required bool `yaml:"required" json:"required"`
	// Default is a default value of the input. This is optional field. nil is set when it is
	// missing.
	Default *string `yaml:"default" json:"default"`
	// Description is description of the input.
	Description string `yaml:"description" json:"description"`
}

// ActionOutput is output metadata of action.
// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#outputs
type ActionOutput struct {
	Description string `yaml:"description" json:"description"`
}

// ActionSpec represents structure of action.yaml.
// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions
type ActionSpec struct {
	// Name is "name" field of action.yaml
	Name string `yaml:"name" json:"name"`
	// Inputs is "inputs" field of action.yaml
	Inputs map[string]*ActionInput `yaml:"inputs" json:"inputs"`
	// Outputs is "outputs" field of action.yaml. Key is name of output.
	Outputs map[string]*ActionOutput `yaml:"outputs" json:"outputs"`
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
		rule.checkActionInSameRepo(spec, e)
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

	// TODO?: Fetch action.yaml from GitHub and check the specification with checkAction()
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
func (rule *RuleAction) checkActionInSameRepo(path string, action *ExecAction) {
	if rule.repoPath == "" {
		return // Give up
	}

	dir := filepath.Join(rule.repoPath, filepath.FromSlash(path))
	b := rule.readActionSpecFile(dir, action.Uses.Pos)
	if len(b) == 0 {
		return
	}

	var spec ActionSpec
	if err := yaml.Unmarshal(b, &spec); err != nil {
		rule.errorf(action.Uses.Pos, "action.yaml in %q is invalid: %s", dir, err.Error())
		return
	}

	rule.checkAction(path, &spec, action)
}

func (rule *RuleAction) checkAction(path string, spec *ActionSpec, exec *ExecAction) {
	// Check specified inputs are defined in action's inputs spec
	for name, val := range exec.Inputs {
		if _, ok := spec.Inputs[name]; !ok {
			ss := make([]string, 0, len(spec.Inputs))
			for k := range spec.Inputs {
				ss = append(ss, k)
			}
			rule.errorf(
				val.Name.Pos,
				"input %q is not defined in action %q defined at %q. available inputs are %s",
				name,
				path,
				spec.Name,
				sortedQuotes(ss),
			)
		}
	}

	// Check mandatory inputs are specified
	for name, input := range spec.Inputs {
		if input.Required && input.Default == nil {
			if _, ok := exec.Inputs[name]; !ok {
				rule.errorf(
					exec.Uses.Pos,
					"missing input %q which is required by action %q defined at %q",
					name,
					spec.Name,
					path,
				)
			}
		}
	}
}

func (rule *RuleAction) readActionSpecFile(dir string, pos *Pos) []byte {
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
