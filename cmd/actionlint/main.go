package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

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

func getwd() (string, error) {
	d, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("could not get current working directory: %w", err)
	}
	return d, nil
}

func lint(args []string, opts *actionlint.LinterOptions) ([]*actionlint.Error, error) {
	l, err := actionlint.NewLinter(os.Stdout, opts)
	if err != nil {
		return nil, err
	}

	if len(args) == 0 {
		// Find nearest workflows directory
		d, err := getwd()
		if err != nil {
			return nil, err
		}
		return l.LintRepository(d)
	}

	if len(args) == 1 && args[0] == "-" {
		b, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("could not read stdin: %w", err)
		}
		return l.Lint("<stdin>", b, nil)
	}

	return l.LintFiles(args)
}

func generateDefaultConfig() error {
	wd, err := getwd()
	if err != nil {
		return err
	}
	proj := actionlint.FindProject(wd)
	if proj == nil {
		return errors.New("project is not found. check current project is initialized as Git repository and \".github/workflows\" directory exists")
	}
	return proj.WriteDefaultConfig()
}

func main() {
	var ver bool
	var opts actionlint.LinterOptions
	var ignorePat string
	var generateConfig bool

	flag.StringVar(&ignorePat, "ignore", "", "Regular expression matching to error messages which you want to ignore")
	flag.StringVar(&opts.Shellcheck, "shellcheck", "shellcheck", "Command name or file path of \"shellcheck\" external command")
	flag.BoolVar(&opts.Oneline, "oneline", false, "Use one line per one error. Useful for reading error messages from programs")
	flag.StringVar(&opts.ConfigFilePath, "config-file", "", "File path to config file")
	flag.BoolVar(&generateConfig, "generate-config", false, "Generate default config file at .github/actionlint.yaml in current project")
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

	if generateConfig {
		if err := generateDefaultConfig(); err != nil {
			fmt.Fprintln(os.Stderr)
			os.Exit(1)
		}
		return
	}

	if ignorePat != "" {
		r, err := regexp.Compile(ignorePat)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid regular expression %q: %s", ignorePat, err.Error())
			os.Exit(1)
		}
		opts.IgnorePattern = r
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
