package actionlint

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"runtime"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// These variables might be modified by ldflags on building release binaries by GoReleaser. Do not modify manually
var (
	version       = ""
	installedFrom = "installed by building from source"
)

const (
	// ExitStatusSuccessNoProblem is the exit status when the command ran successfully with no problem found.
	ExitStatusSuccessNoProblem = 0
	// ExitStatusSuccessProblemFound is the exit status when the command ran successfully with some problem found.
	ExitStatusSuccessProblemFound = 1
	// ExitStatusInvalidCommandOption is the exit status when parsing command line options failed.
	ExitStatusInvalidCommandOption = 2
	// ExitStatusFailure is the exit status when the command stopped due to some fatal error while checking workflows.
	ExitStatusFailure = 3
)

// errLintIssuesFound is a sentinel error returned from RunE when the linter found issues.
// It signals that the process should exit with ExitStatusSuccessProblemFound.
var errLintIssuesFound = errors.New("lint issues found")

func getCommandVersion() string {
	if version != "" {
		return version
	}

	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Version == "" {
		return "unknown" // Reaches only when actionlint package is built outside module
	}

	return info.Main.Version
}

// Command represents entire actionlint command. Given stdin/stdout/stderr are used for input/output.
type Command struct {
	// Stdin is a reader to read input from stdin
	Stdin io.Reader
	// Stdout is a writer to write output to stdout
	Stdout io.Writer
	// Stderr is a writer to write output to stderr
	Stderr io.Writer
}

func (cmd *Command) runLinter(args []string, opts *LinterOptions, initConfig bool) ([]*Error, error) {
	l, err := NewLinter(cmd.Stdout, opts)
	if err != nil {
		return nil, err
	}

	if initConfig {
		return nil, l.GenerateDefaultConfig("")
	}

	if len(args) == 0 {
		return l.LintRepository("")
	}

	if len(args) == 1 && args[0] == "-" {
		return l.LintStdin(cmd.Stdin)
	}

	return l.LintFiles(args, nil)
}

func (cmd *Command) newRootCmd() *cobra.Command {
	var opts LinterOptions
	var ignorePats []string
	var initConfig bool
	var noColor bool
	var color bool
	var ver bool

	v := getCommandVersion()
	b := "main"
	if regexp.MustCompile(`^\d+\.\d+\.\d+$`).MatchString(v) {
		b = "v" + v
	}

	root := &cobra.Command{
		Use:   "actionlint [FLAGS] [FILES...] [-]",
		Short: "actionlint is a linter for GitHub Actions workflow files",
		Args:  cobra.ArbitraryArgs,
		Long: fmt.Sprintf(`actionlint is a linter for GitHub Actions workflow files.

  To check all YAML files in current repository, just run actionlint without
  arguments. It automatically finds the nearest '.github/workflows' directory:

    $ actionlint

  To check specific files, pass the file paths as arguments:

    $ actionlint file1.yaml file2.yaml

  To check content which is not saved in file yet (e.g. output from some
  command), pass - argument. It reads stdin and checks it as workflow file:

    $ actionlint -

  To serialize errors into JSON, use --format option. It allows to format error
  messages flexibly with Go template syntax.

    $ actionlint --format '{{json .}}'

Documents:

  - List of checks: https://github.com/rhysd/actionlint/tree/%s/docs/checks.md
  - Usage:          https://github.com/rhysd/actionlint/tree/%s/docs/usage.md
  - Configuration:  https://github.com/rhysd/actionlint/tree/%s/docs/config.md`, b, b, b),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(c *cobra.Command, args []string) error {
			if ver {
				fmt.Fprintf(
					cmd.Stdout,
					"%s\n%s\nbuilt with %s compiler for %s/%s\n",
					getCommandVersion(),
					installedFrom,
					runtime.Version(),
					runtime.GOOS,
					runtime.GOARCH,
				)
				return nil
			}

			opts.IgnorePatterns = ignorePats
			opts.LogWriter = cmd.Stderr

			if color {
				opts.Color = ColorOptionKindAlways
			}
			if noColor {
				opts.Color = ColorOptionKindNever
			}

			errs, err := cmd.runLinter(args, &opts, initConfig)
			if err != nil {
				return err
			}
			if len(errs) > 0 {
				return errLintIssuesFound
			}
			return nil
		},
	}

	root.SetOut(cmd.Stdout)
	root.SetErr(cmd.Stderr)

	f := root.Flags()
	f.StringArrayVar(&ignorePats, "ignore", nil, "Regular expression matching to error messages you want to ignore. This flag is repeatable")
	f.StringVar(&opts.Shellcheck, "shellcheck", "shellcheck", "Command name or file path of \"shellcheck\" external command. If empty, shellcheck integration will be disabled")
	f.StringVar(&opts.Pyflakes, "pyflakes", "pyflakes", "Command name or file path of \"pyflakes\" external command. If empty, pyflakes integration will be disabled")
	f.BoolVar(&opts.Oneline, "oneline", false, "Use one line per one error. Useful for reading error messages from programs")
	f.StringVar(&opts.Format, "format", "", "Custom template to format error messages in Go template syntax. See the usage documentation for more details")
	f.StringVar(&opts.ConfigFile, "config-file", "", "File path to config file")
	f.BoolVar(&initConfig, "init-config", false, "Generate default config file at .github/actionlint.yaml in current project")
	f.BoolVar(&noColor, "no-color", false, "Disable colorful output")
	f.BoolVar(&color, "color", false, "Always enable colorful output. This is useful to force colorful outputs")
	f.BoolVar(&opts.Verbose, "verbose", false, "Enable verbose output")
	f.BoolVar(&opts.Debug, "debug", false, "Enable debug output (for development)")
	f.BoolVar(&ver, "version", false, "Show version and how this binary was installed")
	f.StringVar(&opts.StdinFileName, "stdin-filename", "<stdin>", "File name when reading input from stdin")

	root.AddCommand(newCompletionCmd(cmd.Stdout))

	return root
}

func newCompletionCmd(out io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:           "completion [bash|zsh|fish|powershell]",
		Short:         "Generate shell completion script",
		ValidArgs:     []string{"bash", "zsh", "fish", "powershell"},
		Args:          cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(c *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return c.Root().GenBashCompletion(out)
			case "zsh":
				return c.Root().GenZshCompletion(out)
			case "fish":
				return c.Root().GenFishCompletion(out, true)
			case "powershell":
				return c.Root().GenPowerShellCompletionWithDesc(out)
			}
			return nil
		},
	}
}

// Main is main function of actionlint. It takes command line arguments as string slice and returns
// exit status. The args should be entire arguments including the program name, usually given via
// os.Args.
func (cmd *Command) Main(args []string) int {
	root := cmd.newRootCmd()
	root.SetArgs(args[1:])

	if err := root.Execute(); err != nil {
		if errors.Is(err, errLintIssuesFound) {
			return ExitStatusSuccessProblemFound
		}
		// Cobra argument/flag parsing errors and cobra.OnlyValidArgs errors
		// are considered invalid command options.
		fmt.Fprintln(cmd.Stderr, err.Error())
		return ExitStatusInvalidCommandOption
	}

	return ExitStatusSuccessNoProblem
}
