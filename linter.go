package actionlint

import (
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func findNearestWorkflowsDir(from string) (string, error) {
	d := from
	for {
		p := filepath.Join(d, ".github", "workflows")
		if s, err := os.Stat(p); err == nil && s.IsDir() {
			return p, nil
		}

		n := filepath.Dir(d)
		if n == d {
			return "", fmt.Errorf("No .github/workflows directory was found in any parent directories of %q", from)
		}
		d = n
	}
}

type Linter struct {
	out io.Writer
	// More options will come here
}

func NewLinter(out io.Writer) *Linter {
	return &Linter{out}
}

// LintRepoDir lints YAML workflow files and outputs the errors to given writer. It finds the nearest
// `.github/workflow` directory based on `dir` and applies lint rules to all YAML worflow files
// under the directory.
func (l *Linter) LintRepoDir(dir string) ([]*Error, error) {
	d, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("Could not get absolute path of %q: %w", dir, err)
	}

	wd, err := findNearestWorkflowsDir(d)
	if err != nil {
		return nil, err
	}

	files := []string{}
	if err := filepath.Walk(wd, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".yaml") {
			files = append(files, path)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("Could not read files in %q: %w", wd, err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("No YAML file was found in %q", wd)
	}

	return l.LintFiles(files)
}

// LintFiles lints YAML workflow files and outputs the errors to given writer.
// It applies lint rules to all given files.
func (l *Linter) LintFiles(filepaths []string) ([]*Error, error) {
	all := []*Error{}

	// TODO: Use multiple threads (per file)
	for _, p := range filepaths {
		errs, err := l.LintFile(p)
		if err != nil {
			return all, err
		}
		all = append(all, errs...)
	}

	return all, nil
}

func (l *Linter) LintFile(path string) ([]*Error, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Could not read %q: %w", path, err)
	}

	// Use relative path if possible
	if wd, err := os.Getwd(); err == nil {
		if r, err := filepath.Rel(wd, path); err == nil {
			path = r
		}
	}

	return l.Lint(path, b)
}

func (l *Linter) Lint(path string, content []byte) ([]*Error, error) {
	_, errs := Parse(content)
	for _, e := range errs {
		// Praser doesn't know where the content came from. Populate filename in the error
		e.Filepath = path
	}

	// TODO: Check workflow syntax tree

	for _, e := range errs {
		fmt.Fprintln(l.out, e)
	}

	return errs, nil
}
