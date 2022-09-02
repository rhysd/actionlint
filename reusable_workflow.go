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
	if c.proj == nil || !strings.HasPrefix(spec, "./") {
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
	}

	m, err := parseReusableWorkflowMetadata(src)
	if err != nil {
		c.writeCache(spec, nil) // Remember the workflow file was invalid
		msg := strings.ReplaceAll(err.Error(), "\n", " ")
		return nil, fmt.Errorf("reusable workflow file %q is invalid: %s", file, msg)
	}

	c.debug("New reusable workflow metadata at %s: %v", file, m)
	c.writeCache(spec, m)
	return m, nil
}

func parseReusableWorkflowMetadata(src []byte) (*ReusableWorkflowMetadata, error) {
	type workflow struct {
		On struct {
			WorkflowCall *ReusableWorkflowMetadata `yaml:"workflow_call"`
		} `yaml:"on"`
	}

	var w workflow
	if err := yaml.Unmarshal(src, &w); err != nil {
		type workflowEventAsValue struct {
			On struct {
				WorkflowCall *ReusableWorkflowMetadata `yaml:"workflow_call"`
			} `yaml:"on"`
		}
		return nil, err
	}
	if w.On.WorkflowCall == nil {
		// When the workflow call is empty like:
		//   on:
		//     workflow_call:
		return &ReusableWorkflowMetadata{}, nil
	}
	return w.On.WorkflowCall, nil
}
