package actionlint

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

//go:generate go run ./scripts/generate-popular-actions -s remote -f go ./popular_actions.go

// ActionMetadataInputRequired represents if the action input is required to be set or not.
// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#inputs
type ActionMetadataInputRequired bool

// UnmarshalYAML implements yaml.Unmarshaler.
func (required *ActionMetadataInputRequired) UnmarshalYAML(n *yaml.Node) error {
	// Name this local type for better error message on unmarshaling
	type actionInputMetadata struct {
		Required bool    `yaml:"required"`
		Default  *string `yaml:"default"`
	}
	var input actionInputMetadata
	if err := n.Decode(&input); err != nil {
		return err
	}
	*required = ActionMetadataInputRequired(input.Required && input.Default == nil)
	return nil
}

// ActionMetadata represents structure of action.yaml.
// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions
type ActionMetadata struct {
	// Name is "name" field of action.yaml
	Name string `yaml:"name" json:"name"`
	// Inputs is "inputs" field of action.yaml
	Inputs map[string]ActionMetadataInputRequired `yaml:"inputs" json:"inputs"`
	// Outputs is "outputs" field of action.yaml. Key is name of output. Description is omitted
	// since actionlint does not use it.
	Outputs map[string]struct{} `yaml:"outputs" json:"outputs"`
	// SkipInputs is flag to specify behavior of inputs check. When it is true, inputs for this
	// action will not be checked.
	SkipInputs bool `json:"skip_inputs"`
	// SkipOutputs is flag to specify a bit loose typing to outputs object. If it is set to
	// true, the outputs object accepts any properties along with strictly typed props.
	SkipOutputs bool `json:"skip_outputs"`
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
func (c *LocalActionsCache) FindMetadata(spec string) (*ActionMetadata, error) {
	if c.proj == nil || !strings.HasPrefix(spec, "./") {
		return nil, nil
	}

	if m, ok := c.readCache(spec); ok {
		c.debug("Cache hit for %s: %v", spec, m)
		return m, nil
	}

	dir := filepath.Join(c.proj.RootDir(), filepath.FromSlash(spec))
	b, err := readLocalActionMetadataFile(dir)
	if err != nil {
		c.writeCache(spec, nil) // Remember action was not found
		return nil, err
	}

	var meta ActionMetadata
	if err := yaml.Unmarshal(b, &meta); err != nil {
		c.writeCache(spec, nil) // Remember action was invalid
		msg := strings.ReplaceAll(err.Error(), "\n", " ")
		return nil, fmt.Errorf("action.yml in %q is invalid: %s", dir, msg)
	}

	c.debug("New metadata parsed from action %s: %v", dir, &meta)

	c.writeCache(spec, &meta)
	return &meta, nil
}

func readLocalActionMetadataFile(dir string) ([]byte, error) {
	for _, p := range []string{
		filepath.Join(dir, "action.yaml"),
		filepath.Join(dir, "action.yml"),
	} {
		if b, err := ioutil.ReadFile(p); err == nil {
			return b, nil
		}
	}

	if wd, err := os.Getwd(); err == nil {
		if p, err := filepath.Rel(wd, dir); err == nil {
			dir = p
		}
	}
	return nil, fmt.Errorf("neither action.yaml nor action.yml is found in directory \"%s\"", dir)
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
