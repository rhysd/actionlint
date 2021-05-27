package linter

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/rhysd/actionlint/syntax"
)

type LintError struct {
	Message  string
	Filename string
	Line     int
	Column   int
}

func (e *LintError) Error() string {
	return fmt.Sprintf("%s:%d:%d: %s", e.Filename, e.Line, e.Column, e.Message)
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
func (l *Linter) LintRepoDir(dir string) ([]*LintError, error) {
	// TODO: Find nearest workflows directory
	// TODO: All YAML files in the directory
	// TODO: Call LintFiles method with the file paths
	panic("TODO")
}

// LintFiles lints YAML workflow files and outputs the errors to given writer.
// It applies lint rules to all given files.
func (l *Linter) LintFiles(filepaths []string) ([]*LintError, error) {
	all := []*LintError{}

	for _, p := range filepaths {
		errs, err := l.LintFile(p)
		if err != nil {
			return all, err
		}
		all = append(all, errs...)
	}

	return all, nil
}

func (l *Linter) LintFile(filepath string) ([]*LintError, error) {
	b, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("Could not read %q: %w", filepath, err)
	}

	_, errs := syntax.Parse(b)

	all := make([]*LintError, 0, len(errs))
	for _, e := range errs {
		all = append(all, &LintError{
			Filename: filepath, // TODO: Use canonical path
			Message:  e.Message,
			Line:     e.Line,
			Column:   e.Column,
		})
	}

	// TODO: Check workflow syntax tree

	for _, e := range all {
		fmt.Fprintln(l.out, e)
	}

	return all, nil
}
