package actionlint

import (
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

// ActionSpec represents structure of action.yaml.
type ActionSpec struct {
	Inputs map[string]struct {
		Required bool    `yaml:"required"`
		Default  *string `yaml:"default"`
	} `yaml:"inputs"`
	Outputs map[string]struct{} `yaml:"outputs"`
}

func (s *ActionSpec) describeInputs() string {
	qs := make([]string, 0, len(s.Inputs))
	for k := range s.Inputs {
		qs = append(qs, strconv.Quote(k))
	}
	return strings.Join(qs, ", ")
}

// NewRuleAction creates new RuleAction instance.
func NewRuleAction(file string) *RuleAction {
	return &RuleAction{
		RuleBase: RuleBase{name: "action"},
		repoPath: findRepositoryRoot(file),
	}
}

func (rule *RuleAction) VisitStep(n *Step) {
	e, ok := n.Exec.(*ExecAction)
	if !ok || e.Uses == nil {
		return
	}

	spec := e.Uses.Value

	if strings.HasPrefix(spec, "./") {
		// Relative to repository root
		rule.checkActionInSameRepo(spec, e)
		return
	}

	if strings.HasPrefix(spec, "docker://") {
		rule.checkDockerAction(spec, e)
		return
	}

	rule.checkRepoAction(spec, e)
}

// Parse {owner}/{repo}@{ref} or {owner}/{repo}/{path}@{ref}
func (rule *RuleAction) checkRepoAction(spec string, exec *ExecAction) {
	s := spec
	idx := strings.IndexRune(s, '/')
	if idx == -1 {
		rule.invalidActionFormat(exec.Uses.Pos, spec)
		return
	}

	// Consume owner name
	owner := s[:idx]
	s = s[idx+1:]

	repo := ""

	if idx := strings.IndexRune(s, '/'); idx >= 0 {
		repo = s[:idx]
		s = s[idx+1:]
		idx = strings.IndexRune(s, '@')
		if idx == -1 {
			rule.invalidActionFormat(exec.Uses.Pos, spec)
			return
		}
		// path = s[:idx]
		s = s[idx+1:]
	} else if idx := strings.IndexRune(s, '@'); idx >= 0 {
		repo = s[:idx]
		s = s[idx+1:]
	} else {
		rule.invalidActionFormat(exec.Uses.Pos, spec)
		return
	}
	tag := s

	if owner == "" || repo == "" || tag == "" {
		rule.invalidActionFormat(exec.Uses.Pos, spec)
	}

	// TODO?: Fetch action.yaml from GitHub and check the specification with checkAction()
}

func (rule *RuleAction) invalidActionFormat(pos *Pos, spec string) {
	rule.errorf(pos, "specifying action %q in invalid format. available formats are \"{owner}/{repo}@{ref}\" or \"{owner}/{repo}/{path}@{ref}\"", spec)
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#example-using-the-github-packages-container-registry
func (rule *RuleAction) checkDockerAction(uri string, exec *ExecAction) {
	tag := ""
	if idx := strings.IndexRune(uri[len("docker://"):], ':'); idx != -1 {
		idx += len("docker://")
		uri = uri[:idx]
		tag = uri[idx+1:]
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

	rule.checkAction(dir, &spec, action)
}

func (rule *RuleAction) checkAction(action string, spec *ActionSpec, exec *ExecAction) {
	// Check specified inputs are defined in action's inputs spec
	for name, val := range exec.Inputs {
		if _, ok := spec.Inputs[name]; !ok {
			rule.errorf(
				val.Name.Pos,
				"input %q is not defined in action %q. available inputs are %s",
				name,
				action,
				spec.describeInputs(),
			)
		}
	}

	// Check mandatory inputs are specified
	for name, input := range spec.Inputs {
		if input.Required && input.Default == nil {
			if _, ok := exec.Inputs[name]; !ok {
				rule.errorf(
					exec.Uses.Pos,
					"missing input %q which is required by action %q",
					name,
					action,
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
	rule.errorf(pos, "Neither action.yaml nor action.yml is found in %q", dir)
	return nil
}

func findRepositoryRoot(file string) string {
	if file == "<stdin>" {
		return ""
	}

	if p, err := filepath.Abs(file); err == nil {
		file = p
	}

	d := filepath.Dir(file)
	for {
		// Note: .git might be a file
		if _, err := os.Stat(filepath.Join(d, ".git")); err == nil {
			return d
		}

		p := filepath.Dir(d)
		if p == d {
			return "" // Not found
		}

		d = p
	}
}
