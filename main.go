package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/rhysd/actionlint/linter"
)

func lint(args []string) ([]*linter.LintError, error) {
	l := linter.NewLinter(os.Stdout)
	if len(args) > 0 {
		return l.LintFiles(args)
	}

	// Find nearest workflows directory
	d, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("Could not get current working directory: %w", err)
	}
	return l.LintRepoDir(d)
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
