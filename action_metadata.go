package actionlint

import (
	"fmt"
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

type cachedLocalActionResult struct {
	meta *ActionMetadata
	err  error
}

// LocalActionsCache is cache for local actions' metadata. It avoids repeating to find/read/parse
// local action's metadata file (action.yml).
type LocalActionsCache struct {
	mu    sync.RWMutex
	proj  *Project // might be nil
	cache map[string]cachedLocalActionResult
	errs  map[string]error
}

// NewLocalActionsCache creates new LocalActionsCache instance.
func NewLocalActionsCache(proj *Project) *LocalActionsCache {
	return &LocalActionsCache{
		proj:  proj,
		cache: map[string]cachedLocalActionResult{},
	}
}

func (c *LocalActionsCache) cached(key string) (cachedLocalActionResult, bool) {
	c.mu.RLock()
	r, ok := c.cache[key]
	c.mu.RUnlock()
	return r, ok
}

func (c *LocalActionsCache) remember(key string, m *ActionMetadata, err error) {
	c.mu.Lock()
	c.cache[key] = cachedLocalActionResult{m, err}
	c.mu.Unlock()
}

// FindMetadata finds metadata for given spec. The spec should indicate for local action hence it
// should start with "./". LocalActionCache caches the both results that it was found and it was
// not found. When the local action was not found, the same error is returned at the next call.
func (c *LocalActionsCache) FindMetadata(spec string) (*ActionMetadata, error) {
	if c.proj == nil || !strings.HasPrefix(spec, "./") {
		return nil, nil
	}

	if r, ok := c.cached(spec); ok {
		return r.meta, r.err
	}

	dir := filepath.Join(c.proj.RootDir(), filepath.FromSlash(spec))
	b, err := readLocalActionMetadataFile(dir)
	if err != nil {
		c.remember(spec, nil, err)
		return nil, err
	}

	var meta ActionMetadata
	if err := yaml.Unmarshal(b, &meta); err != nil {
		msg := strings.ReplaceAll(err.Error(), "\n", " ")
		err := fmt.Errorf("action.yml in %q is invalid: %s", dir, msg)
		c.remember(spec, nil, err)
		return nil, err
	}

	c.remember(spec, &meta, nil)
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
