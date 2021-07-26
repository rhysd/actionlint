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

// LocalActionsCache is cache for local actions' metadata. It avoids repeating to find/read/parse
// local action's metadata file (action.yml).
type LocalActionsCache struct {
	mu    sync.RWMutex
	proj  *Project // might be nil
	cache map[string]*ActionMetadata
	dbg   io.Writer
}

// NewLocalActionsCache creates new LocalActionsCache instance.
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
