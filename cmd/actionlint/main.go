package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/rhysd/actionlint"
)

func lint(args []string) ([]*actionlint.Error, error) {
	l := actionlint.NewLinter(os.Stdout)
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
	var version bool

	flag.BoolVar(&version, "version", false, "Show version")
	flag.Parse()

	if version {
		fmt.Println("0.0.0")
		return
	}

	errs, err := lint(flag.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	if len(errs) > 0 {
		os.Exit(1) // Linter found some issues, yay!
	}
}
