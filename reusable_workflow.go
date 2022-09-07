package actionlint

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// ReusableWorkflowMetadataInput is input metadata for validating local reusable workflow file.
type ReusableWorkflowMetadataInput struct {
	// Required is true when 'required' field of the input is set to true and no default value is set.
	Required bool
	// Type is a type of the input. When the input type is unknown, 'any' type is set.
	Type ExprType
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (input *ReusableWorkflowMetadataInput) UnmarshalYAML(n *yaml.Node) error {
	type metadata struct {
		Required bool    `yaml:"required"`
		Default  *string `yaml:"default"`
		Type     string  `yaml:"type"`
	}

	var md metadata
	if err := n.Decode(&md); err != nil {
		return err
	}

	input.Required = md.Required && md.Default == nil
	switch md.Type {
	case "boolean":
		input.Type = BoolType{}
	case "number":
		input.Type = NumberType{}
	case "string":
		input.Type = StringType{}
	default:
		input.Type = AnyType{}
	}

	return nil
}

// ReusableWorkflowMetadataSecretRequired is metadata to indicate a secret of reusable workflow is
// required or not.
type ReusableWorkflowMetadataSecretRequired bool

// UnmarshalYAML implements yaml.Unmarshaler.
func (required *ReusableWorkflowMetadataSecretRequired) UnmarshalYAML(n *yaml.Node) error {
	type metadata struct {
		Required bool `yaml:"required"`
	}

	var md metadata
	if err := n.Decode(&md); err != nil {
		return err
	}

	*required = ReusableWorkflowMetadataSecretRequired(md.Required)

	return nil
}

// ReusableWorkflowMetadata is metadata to validate local reusable workflows. This struct does not
// contain all metadata from YAML file. It only contains metadata which is necessary to validate
// reusable workflow files by actionlint.
type ReusableWorkflowMetadata struct {
	Inputs  map[string]*ReusableWorkflowMetadataInput         `yaml:"inputs"`
	Outputs map[string]struct{}                               `yaml:"outputs"`
	Secrets map[string]ReusableWorkflowMetadataSecretRequired `yaml:"secrets"`
}

// LocalReusableWorkflowCache is a cache for local reusable workflow metadata files. It avoids find/read/parse
// local reusable workflow YAML files. This cache is dedicated for a single project (repository)
// indicated by 'proj' field. One LocalReusableWorkflowCache instance needs to be created per one
// project.
type LocalReusableWorkflowCache struct {
	mu    sync.RWMutex
	proj  *Project // maybe nil
	cache map[string]*ReusableWorkflowMetadata
	cwd   string
	dbg   io.Writer
}

func (c *LocalReusableWorkflowCache) debug(format string, args ...interface{}) {
	if c.dbg == nil {
		return
	}
	format = "[LocalReusableWorkflowCache] " + format + "\n"
	fmt.Fprintf(c.dbg, format, args...)
}

func (c *LocalReusableWorkflowCache) readCache(key string) (*ReusableWorkflowMetadata, bool) {
	c.mu.RLock()
	m, ok := c.cache[key]
	c.mu.RUnlock()
	return m, ok
}

func (c *LocalReusableWorkflowCache) writeCache(key string, val *ReusableWorkflowMetadata) {
	c.mu.Lock()
	c.cache[key] = val
	c.mu.Unlock()
}

// FindMetadata finds/parses a reusable workflow metadata located by the 'spec' argument. When project
// is not set to 'proj' field or the spec does not start with "./", this method immediately returns with nil.
//
// Note that an error is not cached. At first search, let's say this method returned an error since
// the reusable workflow is invalid. In this case, calling this method with the same spec later will
// not return the error again. It just will return nil. This behavior prevents repeating to report
// the same error from multiple places.
//
// Calling this method is thread-safe.
func (c *LocalReusableWorkflowCache) FindMetadata(spec string) (*ReusableWorkflowMetadata, error) {
	if c.proj == nil || !strings.HasPrefix(spec, "./") || strings.Contains(spec, "${{") {
		return nil, nil
	}

	if m, ok := c.readCache(spec); ok {
		c.debug("Cache hit for %s: %v", spec, m)
		return m, nil
	}

	file := filepath.Join(c.proj.RootDir(), filepath.FromSlash(spec))
	src, err := os.ReadFile(file)
	if err != nil {
		c.debug("File %q was not found: %s", file, err.Error())
		c.writeCache(spec, nil) // Remember the workflow file was not found
		return nil, nil
	}

	m, err := parseReusableWorkflowMetadata(src)
	if err != nil {
		c.writeCache(spec, nil) // Remember the workflow file was invalid
		msg := strings.ReplaceAll(err.Error(), "\n", " ")
		return nil, fmt.Errorf("error while parsing reusable workflow %q: %s", spec, msg)
	}

	c.debug("New reusable workflow metadata at %s: %v", file, m)
	c.writeCache(spec, m)
	return m, nil
}

func (c *LocalReusableWorkflowCache) convWorkflowPathToSpec(p string) (string, bool) {
	if c.proj == nil {
		return "", false
	}
	if !filepath.IsAbs(p) {
		p = filepath.Join(c.cwd, p)
	}
	r := c.proj.RootDir()
	if !strings.HasPrefix(p, r) {
		return "", false
	}
	p, err := filepath.Rel(r, p)
	if err != nil {
		return "", false // Unreachable
	}
	p = filepath.ToSlash(p)
	if !strings.HasPrefix(p, "./") {
		p = "./" + p
	}
	return p, true
}

// WriteWorkflowCallEvent writes reusable workflow metadata by converting from WorkflowCallEvent AST
// node. The 'wpath' parameter is a path to the workflow file of the AST, which is a relative to the
// project root directory or an absolute path.
// This method does nothing when (1) no project is set, (2) it could not convert the workflow path
// to workflow call spec, (3) some cache for the workflow is already existing.
// This method is thread safe.
func (c *LocalReusableWorkflowCache) WriteWorkflowCallEvent(wpath string, event *WorkflowCallEvent) {
	// Convert workflow path to workflow call spec
	spec, ok := c.convWorkflowPathToSpec(wpath)
	if !ok {
		return
	}
	c.debug("Workflow call spec from workflow path %s: %s", wpath, spec)

	c.mu.RLock()
	_, ok = c.cache[spec]
	c.mu.RUnlock()
	if ok {
		return
	}

	m := &ReusableWorkflowMetadata{
		Inputs:  map[string]*ReusableWorkflowMetadataInput{},
		Outputs: map[string]struct{}{},
		Secrets: map[string]ReusableWorkflowMetadataSecretRequired{},
	}

	for n, i := range event.Inputs {
		var t ExprType = AnyType{}
		switch i.Type {
		case WorkflowCallEventInputTypeBoolean:
			t = BoolType{}
		case WorkflowCallEventInputTypeNumber:
			t = NumberType{}
		case WorkflowCallEventInputTypeString:
			t = StringType{}
		}
		m.Inputs[n.Value] = &ReusableWorkflowMetadataInput{
			Type:     t,
			Required: i.Required != nil && i.Required.Value && i.Default == nil,
		}
	}

	for n := range event.Outputs {
		m.Outputs[n.Value] = struct{}{}
	}

	for n, s := range event.Secrets {
		r := s.Required != nil && s.Required.Value
		m.Secrets[n.Value] = ReusableWorkflowMetadataSecretRequired(r)
	}

	c.mu.Lock()
	c.cache[spec] = m
	c.mu.Unlock()

	c.debug("Workflow call metadata from workflow %s: %v", wpath, m)
}

func parseReusableWorkflowMetadata(src []byte) (*ReusableWorkflowMetadata, error) {
	type workflow struct {
		On yaml.Node `yaml:"on"`
	}

	var w workflow
	if err := yaml.Unmarshal(src, &w); err != nil {
		return nil, err
	}

	n := &w.On
	if n.Line == 0 && n.Column == 0 {
		return nil, fmt.Errorf("\"on:\" is not found")
	}

	switch n.Kind {
	case yaml.MappingNode:
		// on:
		//   workflow_call:
		for i := 0; i < len(n.Content); i += 2 {
			k := strings.ToLower(n.Content[i].Value)
			if k == "workflow_call" {
				var m ReusableWorkflowMetadata
				if err := n.Content[i+1].Decode(&m); err != nil {
					return nil, err
				}
				return &m, nil
			}
		}
	case yaml.ScalarNode:
		// on: workflow_call
		if v := strings.ToLower(n.Value); v == "workflow_call" {
			return &ReusableWorkflowMetadata{}, nil
		}
	case yaml.SequenceNode:
		// on: [workflow_call]
		for _, c := range n.Content {
			e := strings.ToLower(c.Value)
			if e == "workflow_call" {
				return &ReusableWorkflowMetadata{}, nil
			}
		}
	}

	return nil, fmt.Errorf("\"workflow_call\" event trigger is not found in \"on:\" at line:%d, column:%d", n.Line, n.Column)
}

// NewLocalReusableWorkflowCache creates a new LocalReusableWorkflowCache instance for the given
// project. 'cwd' is a current working directory as an absolute file path. The 'Local' means that
// the cache instance is project-local. It is not available accross multiple projects.
func NewLocalReusableWorkflowCache(proj *Project, cwd string, dbg io.Writer) *LocalReusableWorkflowCache {
	return &LocalReusableWorkflowCache{
		proj:  proj,
		cache: map[string]*ReusableWorkflowMetadata{},
		cwd:   cwd,
		dbg:   dbg,
	}
}

// LocalReusableWorkflowCacheFactory is a factory object to create a LocalReusableWorkflowCache
// instance per project.
type LocalReusableWorkflowCacheFactory struct {
	caches map[string]*LocalReusableWorkflowCache
	cwd    string
	dbg    io.Writer
}

// NewLocalReusableWorkflowCacheFactory creates a new LocalReusableWorkflowCacheFactory instance.
func NewLocalReusableWorkflowCacheFactory(cwd string, dbg io.Writer) *LocalReusableWorkflowCacheFactory {
	return &LocalReusableWorkflowCacheFactory{map[string]*LocalReusableWorkflowCache{}, cwd, dbg}
}

// GetCache returns a new or existing LocalReusableWorkflowCache instance per project. When a instance
// was already created for the project, this method returns the existing instance. Otherwise it creates
// a new instance and returns it.
func (f *LocalReusableWorkflowCacheFactory) GetCache(p *Project) *LocalReusableWorkflowCache {
	r := p.RootDir()
	if c, ok := f.caches[r]; ok {
		return c
	}
	c := NewLocalReusableWorkflowCache(p, f.cwd, f.dbg)
	f.caches[r] = c
	return c
}
