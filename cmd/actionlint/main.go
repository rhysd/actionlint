package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/rhysd/actionlint"
)

const version = "0.0.0"
const usageHeader = `Usage: actionlint [FLAGS] [FILES...] [-]

  actionlint is a linter for GitHub Actions workflow files.

  To check all YAML files in current repository, just run actionlint without
  arguments. It automatically finds the nearest '.github/workflows' directory:

    $ actionlint

  To check specific files, pass the file paths as arguments:

    $ actionlint file1.yaml file2.yaml

  To check content which is not saved in file yet (e.g. output from some
  command), pass - argument. It reads stdin and checks it as workflow file:

    $ actionlint -

Flags:`

func usage() {
	fmt.Fprintln(os.Stderr, usageHeader)
	flag.PrintDefaults()
}

func lint(args []string, opts *actionlint.LinterOptions) ([]*actionlint.Error, error) {
	l := actionlint.NewLinter(os.Stdout, opts)
	if len(args) == 0 {
		// Find nearest workflows directory
		d, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("Could not get current working directory: %w", err)
		}
		return l.LintRepoDir(d)
	}

	if len(args) == 1 && args[0] == "-" {
		b, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("Could not read stdin: %w", err)
		}
		return l.Lint("<stdin>", b)
	}

	return l.LintFiles(args)
}

func main() {
	var ver bool
	var opts actionlint.LinterOptions

	flag.BoolVar(&opts.Oneline, "oneline", false, "Use one line per one error. Useful for reading error messages from programs")
	flag.BoolVar(&opts.NoColor, "no-color", false, "Disable colorful output")
	flag.BoolVar(&opts.Verbose, "verbose", false, "Enable verbose output")
	flag.BoolVar(&opts.Debug, "debug", false, "Enable debug output (for development)")
	flag.BoolVar(&ver, "version", false, "Show version")
	flag.Usage = usage
	flag.Parse()

	if ver {
		fmt.Println(version)
		return
	}

	errs, err := lint(flag.Args(), &opts)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	if len(errs) > 0 {
		os.Exit(1) // Linter found some issues, yay!
	}
}
