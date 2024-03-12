package actionlint

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

//go:generate go run ./scripts/generate-popular-actions ./popular_actions.go

// ActionMetadataInput is input metadata in "inputs" section in action.yml metadata file.
// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#inputs
type ActionMetadataInput struct {
	// Name is a name of this input.
	Name string `json:"name"`
	// Required is true when this input is mandatory to run the action.
	Required bool `json:"required"`
}

// ActionMetadataInputs is a map from input ID to its metadata. Keys are in lower case since input
// names are case-insensitive.
// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#inputs
type ActionMetadataInputs map[string]*ActionMetadataInput

// UnmarshalYAML implements yaml.Unmarshaler.
func (inputs *ActionMetadataInputs) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind != yaml.MappingNode {
		return unexpectedYAMLKind("inputs", n, yaml.MappingNode)
	}

	type actionInputMetadata struct {
		Required bool    `yaml:"required"`
		Default  *string `yaml:"default"`
	}

	md := make(ActionMetadataInputs, len(n.Content)/2)
	for i := 0; i < len(n.Content); i += 2 {
		k, v := n.Content[i].Value, n.Content[i+1]

		var m actionInputMetadata
		if err := v.Decode(&m); err != nil {
			return err
		}

		id := strings.ToLower(k)
		if _, ok := md[id]; ok {
			return fmt.Errorf("input %q is duplicated", k)
		}

		md[id] = &ActionMetadataInput{k, m.Required && m.Default == nil}
	}

	*inputs = md
	return nil
}

// ActionMetadataOutput is output metadata in "outputs" section in action.yml metadata file.
// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#outputs-for-composite-actions
type ActionMetadataOutput struct {
	Name string `json:"name"`
}

// ActionMetadataOutputs is a map from output ID to its metadata. Keys are in lower case since output
// names are case-insensitive.
// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#outputs-for-composite-actions
type ActionMetadataOutputs map[string]*ActionMetadataOutput

// UnmarshalYAML implements yaml.Unmarshaler.
func (inputs *ActionMetadataOutputs) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind != yaml.MappingNode {
		return unexpectedYAMLKind("outputs", n, yaml.MappingNode)
	}

	md := make(ActionMetadataOutputs, len(n.Content)/2)
	for i := 0; i < len(n.Content); i += 2 {
		k := n.Content[i].Value
		id := strings.ToLower(k)
		if _, ok := md[id]; ok {
			return fmt.Errorf("output %q is duplicated", k)
		}
		md[id] = &ActionMetadataOutput{k}
	}

	*inputs = md
	return nil
}

// ActionMetadataRunnerKind ... TODO
type ActionMetadataRunnerKind uint8

const (
	// ActionMetadataRunnerKindJS ... TODO
	ActionMetadataRunnerKindJS ActionMetadataRunnerKind = iota
	// ActionMetadataRunnerKindComposite ... TODO
	ActionMetadataRunnerKindComposite
	// ActionMetadataRunnerKindDocker ... TODO
	ActionMetadataRunnerKindDocker
)

func (k ActionMetadataRunnerKind) String() string {
	switch k {
	case ActionMetadataRunnerKindJS:
		return "JavaScript"
	case ActionMetadataRunnerKindComposite:
		return "Composite"
	case ActionMetadataRunnerKindDocker:
		return "Docker"
	default:
		panic(fmt.Sprintf("Unknown kind: %d", k))
	}
}

// ActionRunnerMetadata ... TODO
type ActionRunnerMetadata interface {
	// Kind ... TODO
	Kind() ActionMetadataRunnerKind
}

// ActionRunnerJS ... TODO
type ActionRunnerJS struct {
	// Version ... TODO
	Version uint8
	// Main ... TODO
	// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#runsmain
	Main string
	// Pre ... TODO
	// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#runspre
	Pre string
	// PreIf ... TODO
	// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#runspre-if
	PreIf string
	// Post ... TODO
	// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#runspost
	Post string
	// PostIf ... TODO
	// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#runspost-if
	PostIf string
}

// Kind implements ActionRunnerMetadata interface
func (r *ActionRunnerJS) Kind() ActionMetadataRunnerKind {
	return ActionMetadataRunnerKindJS
}

// ActionRunnerComposite ... TODO
type ActionRunnerComposite struct {
	// TODO: Include "steps" array in this node
}

// Kind implements ActionRunnerMetadata interface
func (r *ActionRunnerComposite) Kind() ActionMetadataRunnerKind {
	return ActionMetadataRunnerKindComposite
}

// ActionRunnerDocker ... TODO
// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#runs-for-docker-container-actions
type ActionRunnerDocker struct {
	// Image ... TODO
	// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#runsimage
	Image string
	// PreEntrypoint ... TODO
	// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#runspre-entrypoint
	PreEntrypoint string
	// Entrypoint ... TODO
	// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#runsentrypoint
	Entrypoint string
	// PostEntrypoint ... TODO
	// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#runspost-entrypoint
	PostEntrypoint string
}

// Kind implements ActionRunnerMetadata interface
func (r *ActionRunnerDocker) Kind() ActionMetadataRunnerKind {
	return ActionMetadataRunnerKindDocker
}

// ActionMetadataRuns is "runs" section of action.yaml. It defines how the action is run.
// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#runs
type ActionMetadataRuns struct {
	Runner ActionRunnerMetadata
}

// TODO Move parsing logic to parse.go and make method to parse action.yml there

func findNodeVersionInRunnerName(using string) (uint8, bool) {
	if !strings.HasPrefix(using, "node") {
		return 0, false
	}
	v, err := strconv.ParseUint(using[len("node"):], 10, 0)
	return uint8(v), err == nil
}

func (runs *ActionMetadataRuns) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind != yaml.MappingNode {
		return unexpectedYAMLKind("runs", n, yaml.MappingNode)
	}

	var using *yaml.Node
	for i := 0; i < len(n.Content); i += 2 {
		k := n.Content[i]
		if strings.EqualFold(k.Value, "using") {
			using = n.Content[i+1]
			break
		}
	}
	if using == nil {
		return fmt.Errorf(`"using" is missing in "runs" section at line:%d, col:%d`, n.Line, n.Column)
	}
	if using.Kind != yaml.ScalarNode {
		return unexpectedYAMLKind("runs.using", using, yaml.ScalarNode)
	}
	u := strings.ToLower(using.Value)

	var md ActionRunnerMetadata
	switch u {
	case "composite":
		// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#runs-for-composite-actions
		steps := false
		for i := 0; i < len(n.Content); i += 2 {
			k, v := strings.ToLower(n.Content[i].Value), n.Content[i+1]
			switch k {
			case "using":
				// Do nothing
			case "steps":
				steps = true
				if v.Kind != yaml.SequenceNode {
					return unexpectedYAMLKind("runs.steps", v, yaml.SequenceNode)
				}
			default:
				return fmt.Errorf(`unexpected %q key for "composite" runner in "runs" section at line:%d, col:%d. available keys are "using" and "steps"`, n.Content[i].Value, n.Line, n.Column)
			}
		}
		if !steps {
			return fmt.Errorf(`"steps" key is required for "composite" runner in "runs" section at line%d, col:%d`, n.Line, n.Column)
		}
		md = &ActionRunnerComposite{}
	case "docker":
		// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#runs-for-docker-container-actions
		r := &ActionRunnerDocker{}
		for i := 0; i < len(n.Content); i += 2 {
			k, v := strings.ToLower(n.Content[i].Value), n.Content[i+1]
			switch k {
			case "using":
				// Do nothing
			case "image":
				if v.Kind != yaml.ScalarNode {
					return unexpectedYAMLKind("runs.image", v, yaml.ScalarNode)
				}
				r.Image = v.Value
			case "env":
				if v.Kind != yaml.MappingNode {
					return unexpectedYAMLKind("runs.env", v, yaml.MappingNode)
				}
			case "pre-entrypoint":
				if v.Kind != yaml.ScalarNode {
					return unexpectedYAMLKind("runs.pre-entrypoint", v, yaml.ScalarNode)
				}
				r.PreEntrypoint = v.Value
			case "entrypoint":
				if v.Kind != yaml.ScalarNode {
					return unexpectedYAMLKind("runs.entrypoint", v, yaml.ScalarNode)
				}
				r.Entrypoint = v.Value
			case "post-entrypoint":
				if v.Kind != yaml.ScalarNode {
					return unexpectedYAMLKind("runs.post-entrypoint", v, yaml.ScalarNode)
				}
				r.PostEntrypoint = v.Value
			case "args":
				if v.Kind != yaml.SequenceNode {
					return unexpectedYAMLKind("runs.args", v, yaml.SequenceNode)
				}
			default:
				return fmt.Errorf(`unexpected %q key for "docker" runner in "runs" section at line:%d, col:%d. available keys are "using", "image", "env", "pre-entrypoint", "entrypoint", "post-entrypoint", and "args"`, n.Content[i].Value, n.Line, n.Column)
			}
		}
		if r.Image == "" {
			return fmt.Errorf(`"image" key is required for "docker" runner in "runs" section at line%d, col:%d`, n.Line, n.Column)
		}
		md = r
	default:
		// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#runs-for-javascript-actions
		v, ok := findNodeVersionInRunnerName(u)
		if !ok {
			return fmt.Errorf(`invalid runner name %q at "using" in "runs" section at line:%d, col:%d`, u, n.Line, n.Column)
		}
		r := &ActionRunnerJS{Version: v}
		for i := 0; i < len(n.Content); i += 2 {
			k, v := strings.ToLower(n.Content[i].Value), n.Content[i+1]
			switch k {
			case "using":
				// Do nothing
			case "main":
				if v.Kind != yaml.ScalarNode {
					return unexpectedYAMLKind("runs.main", v, yaml.ScalarNode)
				}
				r.Main = v.Value
			case "pre":
				if v.Kind != yaml.ScalarNode {
					return unexpectedYAMLKind("runs.pre", v, yaml.ScalarNode)
				}
				r.Pre = v.Value
			case "post":
				if v.Kind != yaml.ScalarNode {
					return unexpectedYAMLKind("runs.post", v, yaml.ScalarNode)
				}
				r.Post = v.Value
			case "pre-if":
				if v.Kind != yaml.ScalarNode {
					return unexpectedYAMLKind("runs.pre-if", v, yaml.ScalarNode)
				}
				r.PreIf = v.Value
			case "post-if":
				if v.Kind != yaml.ScalarNode {
					return unexpectedYAMLKind("runs.post-if", v, yaml.ScalarNode)
				}
				r.PostIf = v.Value
			default:
				return fmt.Errorf(`unexpected %q key for %q runner in "runs" section at line:%d, col:%d. available keys are "using", "main", "pre", "pre-if", "post", and "post-if"`, n.Content[i].Value, u, n.Line, n.Column)
			}
		}
		if r.Main == "" {
			return fmt.Errorf(`"main" key is required for %q runner in "runs" section at line%d, col:%d`, u, n.Line, n.Column)
		}
		if r.PreIf != "" && r.Pre == "" {
			return fmt.Errorf(`"pre" key is required when "pre-if" key is specified in "runs" section at line%d, col:%d`, n.Line, n.Column)
		}
		if r.PostIf != "" && r.Post == "" {
			return fmt.Errorf(`"post" key is required when "post-if" key is specified in "runs" section at line%d, col:%d`, n.Line, n.Column)
		}
		md = r
	}

	runs.Runner = md
	return nil
}

// ActionMetadata represents structure of action.yaml.
// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions
type ActionMetadata struct {
	// Name is "name" field of action.yaml.
	Name string `yaml:"name" json:"name"`
	// Inputs is "inputs" field of action.yaml.
	Inputs ActionMetadataInputs `yaml:"inputs" json:"inputs"`
	// Outputs is "outputs" field of action.yaml. Key is name of output. Description is omitted
	// since actionlint does not use it.
	Outputs ActionMetadataOutputs `yaml:"outputs" json:"outputs"`
	// SkipInputs is flag to specify behavior of inputs check. When it is true, inputs for this
	// action will not be checked.
	SkipInputs bool `json:"skip_inputs"`
	// SkipOutputs is flag to specify a bit loose typing to outputs object. If it is set to
	// true, the outputs object accepts any properties along with strictly typed props.
	SkipOutputs bool `json:"skip_outputs"`
	// Runs is "runs" field of action.yaml.
	Runs ActionMetadataRuns `yaml:"runs" json:"-"`
}

// LocalActionsCache is cache for local actions' metadata. It avoids repeating to find/read/parse
// local action's metadata file (action.yml).
// This cache is not available across multiple repositories. One LocalActionsCache instance needs
// to be created per one repository.
type LocalActionsCache struct {
	mu    sync.RWMutex
	proj  *Project // might be nil
	cache map[string]*ActionMetadata
	dbg   io.Writer
}

// NewLocalActionsCache creates new LocalActionsCache instance for the given project.
func NewLocalActionsCache(proj *Project, dbg io.Writer) *LocalActionsCache {
	return &LocalActionsCache{
		proj:  proj,
		cache: map[string]*ActionMetadata{},
		dbg:   dbg,
	}
}

func newNullLocalActionsCache(dbg io.Writer) *LocalActionsCache {
	// Null cache. Cache never hits. It is used when project is not found
	return &LocalActionsCache{dbg: dbg}
}

func (c *LocalActionsCache) debug(format string, args ...interface{}) {
	if c.dbg == nil {
		return
	}
	format = "[LocalActionsCache] " + format + "\n"
	fmt.Fprintf(c.dbg, format, args...)
}

func (c *LocalActionsCache) readCache(key string) (*ActionMetadata, bool) {
	c.mu.RLock()
	m, ok := c.cache[key]
	c.mu.RUnlock()
	return m, ok
}

func (c *LocalActionsCache) writeCache(key string, val *ActionMetadata) {
	c.mu.Lock()
	c.cache[key] = val
	c.mu.Unlock()
}

// FindMetadata finds metadata for given spec. The spec should indicate for local action hence it
// should start with "./". The first return value can be nil even if error did not occur.
// LocalActionCache caches that the action was not found. At first search, it returns an error that
// the action was not found. But at the second search, it does not return an error even if the result
// is nil. This behavior prevents repeating to report the same error from multiple places.
// Calling this method is thread-safe.
func (c *LocalActionsCache) FindMetadata(spec string) (*ActionMetadata, bool, error) {
	if c.proj == nil || !strings.HasPrefix(spec, "./") {
		return nil, false, nil
	}

	if m, ok := c.readCache(spec); ok {
		c.debug("Cache hit for %s: %v", spec, m)
		return m, true, nil
	}

	dir := filepath.Join(c.proj.RootDir(), filepath.FromSlash(spec))
	b, ok := c.readLocalActionMetadataFile(dir)
	if !ok {
		c.debug("No action metadata found in %s", dir)
		// Remember action was not found
		c.writeCache(spec, nil)
		// Do not complain about the action does not exist (#25, #40).
		// It seems a common pattern that the local action does not exist in the repository
		// (e.g. Git submodule) and it is cloned at running workflow (due to a private repository).
		return nil, false, nil
	}

	var meta ActionMetadata
	if err := yaml.Unmarshal(b, &meta); err != nil {
		c.writeCache(spec, nil) // Remember action was invalid
		msg := strings.ReplaceAll(err.Error(), "\n", " ")
		return nil, false, fmt.Errorf("could not parse action metadata in %q: %s", dir, msg)
	}

	c.debug("New metadata parsed from action %s: %v", dir, &meta)

	c.writeCache(spec, &meta)
	return &meta, false, nil
}

func (c *LocalActionsCache) readLocalActionMetadataFile(dir string) ([]byte, bool) {
	for _, p := range []string{
		filepath.Join(dir, "action.yaml"),
		filepath.Join(dir, "action.yml"),
	} {
		if b, err := os.ReadFile(p); err == nil {
			return b, true
		}
	}

	return nil, false
}

// LocalActionsCacheFactory is a factory to create LocalActionsCache instances. LocalActionsCache
// should be created for each repositories. LocalActionsCacheFactory creates new LocalActionsCache
// instance per repository (project).
type LocalActionsCacheFactory struct {
	caches map[string]*LocalActionsCache
	dbg    io.Writer
}

// GetCache returns LocalActionsCache instance for the given project. One LocalActionsCache is
// created per one repository. Created instances are cached and will be used when caches are
// requested for the same projects. This method is not thread safe.
func (f *LocalActionsCacheFactory) GetCache(p *Project) *LocalActionsCache {
	if p == nil {
		return newNullLocalActionsCache(f.dbg)
	}
	r := p.RootDir()
	if c, ok := f.caches[r]; ok {
		return c
	}
	c := NewLocalActionsCache(p, f.dbg)
	f.caches[r] = c
	return c
}

// NewLocalActionsCacheFactory creates a new LocalActionsCacheFactory instance.
func NewLocalActionsCacheFactory(dbg io.Writer) *LocalActionsCacheFactory {
	return &LocalActionsCacheFactory{map[string]*LocalActionsCache{}, dbg}
}
