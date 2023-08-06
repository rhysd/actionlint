package actionlint

import (
	"os"
	"path/filepath"
	"strings"
)

// Project represents one GitHub project. One Git repository corresponds to one project.
type Project struct {
	root   string
	config *Config
}

func absPath(path string) string {
	if p, err := filepath.Abs(path); err == nil {
		path = p
	}
	return path
}

// findProject creates new Project instance by finding a project which the given path belongs to.
// A project must be a Git repository and have ".github/workflows" directory.
func findProject(path string) (*Project, error) {
	d := absPath(path)
	for {
		if s, err := os.Stat(filepath.Join(d, ".github", "workflows")); err == nil && s.IsDir() {
			if _, err := os.Stat(filepath.Join(d, ".git")); err == nil { // Note: .git may be a file
				return NewProject(d)
			}
		}

		p := filepath.Dir(d)
		if p == d {
			return nil, nil
		}
		d = p
	}
}

// NewProject creates a new instance with a file path to the root directory of the repository.
// This function returns an error when failing to parse an actionlint config file in the repository.
func NewProject(root string) (*Project, error) {
	c, err := loadRepoConfig(root)
	if err != nil {
		return nil, err
	}
	return &Project{root, c}, nil
}

// RootDir returns a root directory path of the GitHub project repository.
func (p *Project) RootDir() string {
	return p.root
}

// WorkflowsDir returns a ".github/workflows" directory path of the GitHub project repository.
// This method does not check if the directory exists.
func (p *Project) WorkflowsDir() string {
	return filepath.Join(p.root, ".github", "workflows")
}

// Knows returns true when the project knows the given file. When a file is included in the
// project's directory, the project knows the file.
func (p *Project) Knows(path string) bool {
	// TODO: strings.HasPrefix is not perfect to check file path
	return strings.HasPrefix(absPath(path), p.root)
}

// Config returns config object of the GitHub project repository. The config file was read from
// ".github/actionlint.yaml" or ".github/actionlint.yml" when this Project instance was created.
// When no config was found, this method returns nil.
func (p *Project) Config() *Config {
	// Note: Calling this method must be thread safe (#333)
	return p.config
}

// Projects represents set of projects. It caches Project instances which was created previously
// and reuses them.
type Projects struct {
	known []*Project
}

// NewProjects creates new Projects instance.
func NewProjects() *Projects {
	return &Projects{}
}

// At returns the Project instance which the path belongs to. It returns nil if no project is found
// from the path.
func (ps *Projects) At(path string) (*Project, error) {
	for _, p := range ps.known {
		if p.Knows(path) {
			return p, nil
		}
	}

	p, err := findProject(path)
	if err != nil {
		return nil, err
	}
	if p != nil {
		ps.known = append(ps.known, p)
	}

	return p, nil
}
