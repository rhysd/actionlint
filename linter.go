package actionlint

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/kr/pretty"
	"github.com/mattn/go-colorable"
)

// LogLevel is log level of logger used in Linter instance.
type LogLevel int

const (
	// LogLevelNone does not output any log output.
	LogLevelNone LogLevel = 0
	// LogLevelVerbose shows verbose log output. This is equivalent to specifying -verbose option
	// to actionlint command.
	LogLevelVerbose = 1
	// LogLevelDebug shows all log output including debug information.
	LogLevelDebug = 2
)

// LinterOptions is set of options for Linter instance. This struct should be created by user and
// given to NewLinter factory function.
type LinterOptions struct {
	// Verbose is flag if verbose log output is enabled.
	Verbose bool
	// Debug is flag if debug log output is enabled.
	Debug bool
	// LogWriter is io.Writer object to use to print log outputs. Note that error outputs detected
	// by the linter are not included in the log outputs.
	LogWriter io.Writer
	// NoColor is flag if colorful output is enabled.
	NoColor bool
	// Oneline is flag if one line output is enabled. When enabling it, one error is output per one
	// line. It is useful when reading outputs from programs.
	Oneline bool
	// Shellcheck is executable for running shellcheck external command. It can be command name like
	// "shellcheck" or file path like "/path/to/shellcheck", "path/to/shellcheck". When this value
	// is empty, shellcheck won't run to check scripts in workflow file.
	Shellcheck string
	// Pyflakes is executable for running pyflakes external command. It can be command name like "pyflakes"
	// or file path like "/path/to/pyflakes", "path/to/pyflakes". When this value is empty, pyflakes
	// won't run to check scripts in workflow file.
	Pyflakes string
	// IgnorePatterns is list of regular expression to filter errors. The pattern is applied to error
	// messages. When an error is matched, the error is ignored.
	IgnorePatterns []string
	// ConfigFile is a path to config file. Empty string means no config file path is given. In
	// the case, actionlint will try to read config from .github/actionlint.yaml.
	ConfigFile string
	// More options will come here
}

// Linter is struct to lint workflow files.
type Linter struct {
	projects      *Projects
	out           io.Writer
	logOut        io.Writer
	logLevel      LogLevel
	noColor       bool
	oneline       bool
	shellcheck    string
	pyflakes      string
	ignorePats    []*regexp.Regexp
	defaultConfig *Config
}

// NewLinter creates a new Linter instance.
// The out parameter is used to output errors from Linter instance. Set io.Discard if you don't
// want the outputs.
// The opts parameter is LinterOptions instance which configures behavior of linting.
func NewLinter(out io.Writer, opts *LinterOptions) (*Linter, error) {
	l := LogLevelNone
	if opts.Verbose {
		l = LogLevelVerbose
	} else if opts.Debug {
		l = LogLevelDebug
	}
	if opts.NoColor {
		color.NoColor = true
	} else {
		// Allow colorful output on Windows
		if f, ok := out.(*os.File); ok {
			out = colorable.NewColorable(f)
		}
	}

	var lout io.Writer = os.Stderr
	if opts.LogWriter != nil {
		lout = opts.LogWriter
	}

	var cfg *Config
	if opts.ConfigFile != "" {
		c, err := readConfigFile(opts.ConfigFile)
		if err != nil {
			return nil, err
		}
		cfg = c
	}

	ignore := make([]*regexp.Regexp, 0, len(opts.IgnorePatterns))
	for _, s := range opts.IgnorePatterns {
		r, err := regexp.Compile(s)
		if err != nil {
			return nil, fmt.Errorf("invalid regular expression for ignore pattern %q: %s", s, err.Error())
		}
		ignore = append(ignore, r)
	}

	return &Linter{
		NewProjects(),
		out,
		lout,
		l,
		opts.NoColor,
		opts.Oneline,
		opts.Shellcheck,
		opts.Pyflakes,
		ignore,
		cfg,
	}, nil
}

func (l *Linter) log(args ...interface{}) {
	if l.logLevel < LogLevelVerbose {
		return
	}
	fmt.Fprint(l.logOut, "verbose: ")
	fmt.Fprintln(l.logOut, args...)
}

func (l *Linter) debug(format string, args ...interface{}) {
	if l.logLevel < LogLevelDebug {
		return
	}
	format = fmt.Sprintf("[Linter] %s\n", format)
	fmt.Fprintf(l.logOut, format, args...)
}

// GenerateDefaultConfig generates default config file at ".github/actionlint.yaml" in project
// which the given directory path belongs to.
func (l *Linter) GenerateDefaultConfig(dir string) error {
	l.log("Generating default actionlint.yaml in repository:", dir)

	p := l.projects.At(dir)
	if p == nil {
		return errors.New("project is not found. check current project is initialized as Git repository and \".github/workflows\" directory exists")
	}

	path := filepath.Join(p.RootDir(), ".github", "actionlint.yaml")
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("config file already exists at %q", path)
	}

	if err := writeDefaultConfigFile(path); err != nil {
		return err
	}

	fmt.Fprintf(l.out, "Config file was generated at %q\n", path)
	return nil
}

// LintRepository lints YAML workflow files and outputs the errors to given writer. It finds the nearest
// `.github/workflow` directory based on `dir` and applies lint rules to all YAML worflow files
// under the directory.
func (l *Linter) LintRepository(dir string) ([]*Error, error) {
	l.log("Linting all workflow files in repository:", dir)

	proj := l.projects.At(dir)
	if proj == nil {
		return nil, fmt.Errorf("no project was found in any parent directories of %q. check workflows directory is put correctly in your Git repository", dir)
	}

	l.log("Detected project:", proj.RootDir())
	wd := proj.WorkflowsDir()

	files := []string{}
	if err := filepath.Walk(wd, func(path string, info os.FileInfo, err error) error {
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
		return nil, fmt.Errorf("could not read files in %q: %w", wd, err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no YAML file was found in %q", wd)
	}
	l.log("Collected", len(files), "YAML files")

	return l.LintFiles(files, proj)
}

// LintFiles lints YAML workflow files and outputs the errors to given writer. It applies lint
// rules to all given files. The project parameter can be nil. In the case, a project is detected
// from the file path.
func (l *Linter) LintFiles(filepaths []string, project *Project) ([]*Error, error) {
	n := len(filepaths)
	if n > 1 {
		l.log("Linting", n, "files")
	}

	all := []*Error{}

	// TODO: Use multiple threads (per file)
	for _, p := range filepaths {
		errs, err := l.LintFile(p, project)
		if err != nil {
			return all, err
		}
		all = append(all, errs...)
	}
	if n > 1 {
		l.log("Found", len(all), "errors in", n, "files")
	}

	return all, nil
}

// LintFile lints one YAML workflow file and outputs the errors to given writer. The project
//parameter can be nil. In the case, the project is detected from the given path.
func (l *Linter) LintFile(path string, project *Project) ([]*Error, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read %q: %w", path, err)
	}

	if project == nil {
		project = l.projects.At(path)
	}

	// Use relative path if possible
	if wd, err := os.Getwd(); err == nil {
		if r, err := filepath.Rel(wd, path); err == nil {
			path = r
		}
	}

	return l.Lint(path, b, project)
}

// Lint lints YAML workflow file content given as byte sequence. The path parameter is used as file
// path the content came from. Setting "<stdin>" to path parameter indicates the output came from
// STDIN.
// Note that only given Project instance is used for configuration. No config is automatically loaded
// based on path parameter.
func (l *Linter) Lint(path string, content []byte, project *Project) ([]*Error, error) {
	var start time.Time
	if l.logLevel >= LogLevelVerbose {
		start = time.Now()
	}

	l.log("Linting", path)
	if project != nil {
		l.log("Using project at", project.RootDir())
	}

	var cfg *Config
	if l.defaultConfig != nil {
		cfg = l.defaultConfig
	} else if project != nil {
		c, err := project.Config()
		if err != nil {
			return nil, err
		}
		cfg = c
	}
	if l.logLevel >= LogLevelDebug {
		if cfg != nil {
			pretty.Fprintf(l.logOut, "[Linter] Config: %# v\n", cfg)
		} else {
			l.debug("No config was found")
		}
	}

	w, all := Parse(content)

	if l.logLevel >= LogLevelVerbose {
		elapsed := time.Since(start)
		l.log("Found", len(all), "parse errors in", elapsed.Milliseconds(), "ms for", path)
	}

	if l.logLevel >= LogLevelDebug {
		fmt.Fprintln(l.logOut, "========== WORKFLOW TREE START ==========")
		pretty.Fprintf(l.logOut, "%# v\n", w)
		fmt.Fprintln(l.logOut, "=========== WORKFLOW TREE END ===========")
	}

	if w != nil {
		dbg := l.logOut
		if l.logLevel < LogLevelDebug {
			dbg = nil
		}

		var labels []string
		if cfg != nil {
			labels = cfg.SelfHostedRunner.Labels
		}

		var root string
		if project != nil {
			root = project.RootDir()
		}

		rules := []Rule{
			NewRuleMatrix(),
			NewRuleCredentials(),
			NewRuleShellName(),
			NewRuleRunnerLabel(labels),
			NewRuleEvents(),
			NewRuleJobNeeds(),
			NewRuleAction(root),
			NewRuleEnvVar(),
			NewRuleStepID(),
			NewRuleExpression(),
		}
		if l.shellcheck != "" {
			r, err := NewRuleShellcheck(l.shellcheck)
			if err == nil {
				rules = append(rules, r)
			} else {
				l.log("Rule \"shellcheck\" was disabled:", err)
			}
		} else {
			l.log("Rule \"shellcheck\" was disabled since shellcheck command name was empty")
		}
		if l.pyflakes != "" {
			r, err := NewRulePyflakes(l.pyflakes)
			if err == nil {
				rules = append(rules, r)
			} else {
				l.log("Rule \"pyflakes\" was disabled:", err)
			}
		} else {
			l.log("Rule \"pyflakes\" was disabled since pyflakes command name was empty")
		}

		v := NewVisitor()
		for _, rule := range rules {
			v.AddPass(rule)
		}
		if dbg != nil {
			v.EnableDebug(dbg)
			for _, r := range rules {
				r.EnableDebug(dbg)
			}
		}

		if err := v.Visit(w); err != nil {
			l.debug("error occurred while visiting workflow syntax tree: %v", err)
			return nil, err
		}

		for _, rule := range rules {
			errs := rule.Errs()
			l.debug("%s found %d errors", rule.Name(), len(errs))
			all = append(all, errs...)
		}
	}

	if len(l.ignorePats) > 0 {
		filtered := make([]*Error, 0, len(all))
	Loop:
		for _, err := range all {
			for _, pat := range l.ignorePats {
				if pat.MatchString(err.Message) {
					continue Loop
				}
			}
			filtered = append(filtered, err)
		}
		all = filtered
	}

	sort.Sort(ByErrorPosition(all))

	src := content
	if l.oneline {
		src = nil
	}
	for _, err := range all {
		err.Filepath = path // Populate filename in the error
		err.PrettyPrint(l.out, src)
	}

	if l.logLevel >= LogLevelVerbose {
		elapsed := time.Since(start)
		l.log("Found total", len(all), "errors in", elapsed.Milliseconds(), "ms for", path)
	}

	return all, nil
}
