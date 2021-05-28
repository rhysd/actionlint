package actionlint

import (
	"fmt"
	"io"
	"io/ioutil"
)

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
	// TODO: Find nearest workflows directory
	// TODO: All YAML files in the directory
	// TODO: Call LintFiles method with the file paths
	panic("TODO")
}

// LintFiles lints YAML workflow files and outputs the errors to given writer.
// It applies lint rules to all given files.
func (l *Linter) LintFiles(filepaths []string) ([]*Error, error) {
	all := []*Error{}

	for _, p := range filepaths {
		errs, err := l.LintFile(p)
		if err != nil {
			return all, err
		}
		all = append(all, errs...)
	}

	return all, nil
}

func (l *Linter) LintFile(filepath string) ([]*Error, error) {
	b, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("Could not read %q: %w", filepath, err)
	}

	_, errs := Parse(b)
	for _, e := range errs {
		// Populate filename in the error
		e.Filename = filepath // TODO: Use canonical path
	}

	// TODO: Check workflow syntax tree

	for _, e := range errs {
		fmt.Fprintln(l.out, e)
	}

	return errs, nil
}
