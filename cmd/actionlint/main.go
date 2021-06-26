package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"runtime/debug"

	"github.com/rhysd/actionlint"
)

// These variables might be modified by ldflags on building release binaries by GoReleaser. Do not modify manually
var (
	version = ""
	gotFrom = "built from source"
)

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

Configuration:

  Configuration file can be put at:

    .github/actionlint.yaml
    .github/actionlint.yml

  Please generate default configuration file and check comments in the file for
  more details.

    $ actionlint -init-config

Flags:`

func usage() {
	fmt.Fprintln(os.Stderr, usageHeader)
	flag.PrintDefaults()
}

func getVersion() string {
	if version != "" {
		return version
	}

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown" // Should be unreachable though
	}

	return info.Main.Version
}

func run(args []string, opts *actionlint.LinterOptions, initConfig bool) ([]*actionlint.Error, error) {
	l, err := actionlint.NewLinter(os.Stdout, opts)
	if err != nil {
		return nil, err
	}

	if initConfig {
		return nil, l.GenerateDefaultConfig(".")
	}

	if len(args) == 0 {
		return l.LintRepository(".")
	}

	if len(args) == 1 && args[0] == "-" {
		b, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("could not read stdin: %w", err)
		}
		return l.Lint("<stdin>", b, nil)
	}

	return l.LintFiles(args, nil)
}

type ignorePatterns []string

func (i *ignorePatterns) String() string {
	return "option for ignore patterns"
}
func (i *ignorePatterns) Set(v string) error {
	*i = append(*i, v)
	return nil
}

func main() {

	var ver bool
	var opts actionlint.LinterOptions
	var ignorePats ignorePatterns
	var initConfig bool

	flag.Var(&ignorePats, "ignore", "Regular expression matching to error messages you want to ignore. This flag can be specified multiple times")
	flag.StringVar(&opts.Shellcheck, "shellcheck", "shellcheck", "Command name or file path of \"shellcheck\" external command")
	flag.StringVar(&opts.Pyflakes, "pyflakes", "pyflakes", "Command name or file path of \"pyflakes\" external command")
	flag.BoolVar(&opts.Oneline, "oneline", false, "Use one line per one error. Useful for reading error messages from programs")
	flag.StringVar(&opts.ConfigFile, "config-file", "", "File path to config file")
	flag.BoolVar(&initConfig, "init-config", false, "Generate default config file at .github/actionlint.yaml in current project")
	flag.BoolVar(&opts.NoColor, "no-color", false, "Disable colorful output")
	flag.BoolVar(&opts.Verbose, "verbose", false, "Enable verbose output")
	flag.BoolVar(&opts.Debug, "debug", false, "Enable debug output (for development)")
	flag.BoolVar(&ver, "version", false, "Show version and how this binary was installed")
	flag.Usage = usage
	flag.Parse()

	if ver {
		fmt.Printf("%s\n%s\n", getVersion(), gotFrom)
		return
	}

	opts.IgnorePatterns = ignorePats

	errs, err := run(flag.Args(), &opts, initConfig)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	if len(errs) > 0 {
		os.Exit(1) // Linter found some issues, yay!
	}
}
