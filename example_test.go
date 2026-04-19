package actionlint_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rhysd/actionlint"
)

func ExampleLinter() {
	// Specify linter options
	o := &actionlint.LinterOptions{
		IgnorePatterns: []string{`'label ".+" is unknown'`},
		// Other options...
	}

	// Create Linter instance which outputs errors to stdout
	l, err := actionlint.NewLinter(os.Stdout, o)
	if err != nil {
		panic(err)
	}

	// File to check
	f := filepath.Join("testdata", "examples", "main.yaml")

	// First return value is an array of lint errors found in the workflow files. The second return
	// value is an error of actionlint itself. This call outputs the lint errors to stdout. Use
	// io.Discard to prevent the output.
	//
	// There are several methods to run linter.
	// - LintFile: Check the given single file
	// - LintFiles: Check the given multiple files
	// - LintDir: Check all workflow files in the given single directory recursively
	// - LintRepository: Check all workflow files under .github/workflows in the given repository
	// - LintStdin: Check the given workflow content read from STDIN
	// - Lint: Check the given workflow content assuming the given file path
	errs, err := l.LintFile(f, nil)
	if err != nil {
		panic(err)
	}

	fmt.Println(len(errs), "lint errors found by actionlint")
}

func ExampleErrorFormatter() {
	// Errors returned from Linter methods
	errs := []*actionlint.Error{
		{
			Message:  "error message 1",
			Filepath: "foo.yaml",
			Line:     1,
			Column:   4,
			Kind:     "rule1",
		},
		{
			Message:  "error message 2",
			Filepath: "foo.yaml",
			Line:     3,
			Column:   1,
			Kind:     "rule2",
		},
	}

	// Create ErrorFormatter instance with template
	f, err := actionlint.NewErrorFormatter(`{{range $ := .}}{{$.Filepath}}:{{$.Line}}:{{$.Column}}: {{$.Message}}\n{{end}}`)
	if err != nil {
		// Some error happened while creating the formatter (e.g. syntax error)
		panic(err)
	}

	src := `on: push

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: echo
`

	// Prints all errors to stdout following the template
	if err := f.PrintErrors(os.Stdout, errs, []byte(src)); err != nil {
		panic(err)
	}

	// Output:
	// foo.yaml:1:4: error message 1
	// foo.yaml:3:1: error message 2
}

func ExampleCommand() {
	// Write command output to this buffer
	var output bytes.Buffer

	// Create command instance populating stdin/stdout/stderr
	cmd := actionlint.Command{
		Stdin:  os.Stdin,
		Stdout: &output,
		Stderr: &output,
	}

	// Run the command end-to-end. Note that given args should contain program name
	workflow := filepath.Join(".github", "workflows", "release.yaml")
	status := cmd.Main([]string{"actionlint", "-shellcheck=", "-pyflakes=", workflow})

	fmt.Println("Exited with status", status)
	// Output: Exited with status 0

	if status != 0 {
		panic("actionlint command failed: " + output.String())
	}
}
