package actionlint

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
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

// ColorOptionKind is kind of colorful output behavior.
type ColorOptionKind int

const (
	// ColorOptionKindAuto is kind to determine to colorize errors output automatically. It is
	// determined based on pty and $NO_COLOR environment variable. See document of fatih/color
	// for more details.
	ColorOptionKindAuto ColorOptionKind = iota
	// ColorOptionKindAlways is kind to always colorize errors output.
	ColorOptionKindAlways
	// ColorOptionKindNever is kind never to colorize errors output.
	ColorOptionKindNever
)

// LinterOptions is set of options for Linter instance. This struct is used for NewLinter factory
// function call. The zero value LinterOptions{} represents the default behavior.
type LinterOptions struct {
	// Verbose is flag if verbose log output is enabled.
	Verbose bool
	// Debug is flag if debug log output is enabled.
	Debug bool
	// LogWriter is io.Writer object to use to print log outputs. Note that error outputs detected
	// by the linter are not included in the log outputs.
	LogWriter io.Writer
	// Color is option for colorizing error outputs. See ColorOptionKind document for each enum values.
	Color ColorOptionKind
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
	// Format is a custom template to format error messages. It must follow Go Template format and
	// contain at least one {{ }} placeholder. https://pkg.go.dev/text/template
	Format string
	// StdinFileName is a file name when reading input from stdin. When this value is empty, "<stdin>"
	// is used as the default value.
	StdinFileName string
	// WorkingDir is a file path to the current working directory. When this value is empty, os.Getwd
	// will be used to get a working directory.
	WorkingDir string
	// OnRulesCreated is a hook to add or remove the check rules. This function is called on checking
	// every workflow files. Rules created by Linter instance are passed to the argument and the
	// function should return the modified rules.
	// Note that syntax errors may be reported even if this function returns nil or an empty slice.
	OnRulesCreated func([]Rule) []Rule
	// InputFormat decides whether directories or files should be linted as workflow files or Action
	// metadata
	InputFormat string
	// More options will come here
}

// Linter is struct to lint workflow files.
type Linter struct {
	projects       *Projects
	out            io.Writer
	logOut         io.Writer
	logLevel       LogLevel
	oneline        bool
	shellcheck     string
	pyflakes       string
	ignorePats     []*regexp.Regexp
	defaultConfig  *Config
	errFmt         *ErrorFormatter
	cwd            string
	onRulesCreated func([]Rule) []Rule
	inputFormat    InputFormat
}

// NewLinter creates a new Linter instance.
// The out parameter is used to output errors from Linter instance. Set io.Discard if you don't
// want the outputs.
// The opts parameter is LinterOptions instance which configures behavior of linting.
func NewLinter(out io.Writer, opts *LinterOptions) (*Linter, error) {
	level := LogLevelNone
	if opts.Verbose {
		level = LogLevelVerbose
	} else if opts.Debug {
		level = LogLevelDebug
	}

	if opts.Color == ColorOptionKindNever {
		color.NoColor = true
	} else {
		if opts.Color == ColorOptionKindAlways {
			color.NoColor = false
		}
		// Allow colorful output on Windows
		if f, ok := out.(*os.File); ok {
			out = colorable.NewColorable(f)
		}
	}

	var lout io.Writer = io.Discard
	if opts.LogWriter != nil {
		lout = opts.LogWriter
	}

	var cfg *Config
	if opts.ConfigFile != "" {
		c, err := ReadConfigFile(opts.ConfigFile)
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

	var formatter *ErrorFormatter
	if opts.Format != "" {
		f, err := NewErrorFormatter(opts.Format)
		if err != nil {
			return nil, err
		}
		formatter = f
	}

	cwd := opts.WorkingDir
	if cwd == "" {
		if d, err := os.Getwd(); err == nil {
			cwd = d
		}
	}
	var inputFormat InputFormat
	switch opts.InputFormat {
	case "workflow":
		inputFormat = FileWorkflow
	case "action":
		inputFormat = FileAction
	case "":
		fallthrough
	case "auto-detect":
		inputFormat = FileAutoDetect
	default:
		return nil, fmt.Errorf("unknown file syntax choose 'workflow', 'action' or 'auto-detect'")
	}

	return &Linter{
		NewProjects(),
		out,
		lout,
		level,
		opts.Oneline,
		opts.Shellcheck,
		opts.Pyflakes,
		ignore,
		cfg,
		formatter,
		cwd,
		opts.OnRulesCreated,
		inputFormat,
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

func (l *Linter) debugWriter() io.Writer {
	if l.logLevel < LogLevelDebug {
		return nil
	}
	return l.logOut
}

// GenerateDefaultConfig generates default config file at ".github/actionlint.yaml" in project
// which the given directory path belongs to.
func (l *Linter) GenerateDefaultConfig(dir string) error {
	l.log("Generating default actionlint.yaml in repository:", dir)

	p, err := l.projects.At(dir)
	if err != nil {
		return err
	}
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

// LintRepository lints YAML workflow and action files and outputs the errors to given writer.
// It finds the nearest `.github/workflows` directory based on `dir` and applies lint rules to
// all YAML workflow files under the directory.
// Action metadata files are searched for in all other directories of the repository.
// Note, the InputFormat option can be used to filter this behavior.
func (l *Linter) LintRepository(dir string) ([]*Error, error) {
	l.log("Linting all", inputFormatString(l.inputFormat), "files in repository:", dir)

	p, err := l.projects.At(dir)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, fmt.Errorf("no project was found in any parent directories of %q. check workflows directory is put correctly in your Git repository", dir)
	}

	l.log("Detected project:", p.RootDir())
	var parentDir string
	if l.inputFormat == FileWorkflow {
		parentDir = p.WorkflowsDir()
	} else {
		parentDir = p.RootDir()
	}
	return l.LintDir(parentDir, p)
}

// LintDirInRepository lints YAML workflow or actions files and outputs the errors to given writer.
// It finds the projected based on the nearest `.github/workflows` directory based on `dir`.
// However, only files within `dir` are linted.
func (l *Linter) LintDirInRepository(dir string) ([]*Error, error) {
	l.log("Linting all", inputFormatString(l.inputFormat), "files in repository:", dir)

	p, err := l.projects.At(dir)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, fmt.Errorf("no project was found in any parent directories of %q. check workflows directory is put correctly in your Git repository", dir)
	}

	l.log("Detected project:", p.RootDir())
	return l.LintDir(dir, p)
}

// LintDir lints all YAML files in the given directory recursively.
func (l *Linter) LintDir(dir string, project *Project) ([]*Error, error) {
	files := []string{}
	workflow_dir := project.WorkflowsDir()
	if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		includeAnyYaml := false
		switch l.inputFormat {
		case FileAction:
			includeAnyYaml = false
		case FileWorkflow:
			includeAnyYaml = true
		case FileAutoDetect:
			includeAnyYaml = strings.HasPrefix(path, workflow_dir)
		}
		if !includeAnyYaml && (strings.HasSuffix(path, "action.yml") || strings.HasSuffix(path, "action.yaml")) {
			files = append(files, path)
		}
		if includeAnyYaml && (strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".yaml")) {
			files = append(files, path)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("could not read files in %q: %w", dir, err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no YAML file was found in %q", dir)
	}
	l.log("Collected", len(files), "YAML files")

	// To make output deterministic, sort order of file paths
	sort.Strings(files)

	return l.LintFiles(files, project)
}

// LintFiles lints YAML workflow files and outputs the errors to given writer. It applies lint
// rules to all given files. The project parameter can be nil. In the case, a project is detected
// from the file path.
func (l *Linter) LintFiles(filepaths []string, project *Project) ([]*Error, error) {
	n := len(filepaths)
	switch n {
	case 0:
		return []*Error{}, nil
	case 1:
		return l.LintFile(filepaths[0], project)
	}

	l.log("Linting", n, "files")

	cwd := l.cwd
	proc := newConcurrentProcess(runtime.NumCPU())
	sema := semaphore.NewWeighted(int64(runtime.NumCPU()))
	ctx := context.Background()
	dbg := l.debugWriter()
	acf := NewLocalActionsCacheFactory(dbg)
	rwcf := NewLocalReusableWorkflowCacheFactory(cwd, dbg)

	type workspace struct {
		path string
		errs []*Error
		src  []byte
	}

	ws := make([]workspace, 0, len(filepaths))
	for _, p := range filepaths {
		ws = append(ws, workspace{path: p})
	}

	eg := errgroup.Group{}
	for i := range ws {
		// Each element of ws is accessed by single goroutine so mutex is unnecessary
		w := &ws[i]
		proj := project
		if proj == nil {
			// This method modifies state of l.projects so it cannot be called in parallel.
			// Before entering goroutine, resolve project instance.
			p, err := l.projects.At(w.path)
			if err != nil {
				return nil, err
			}
			proj = p
		}
		ac := acf.GetCache(proj) // #173
		rwc := rwcf.GetCache(proj)

		eg.Go(func() error {
			// Bound concurrency on reading files to avoid "too many files to open" error (issue #3)
			sema.Acquire(ctx, 1)
			src, err := os.ReadFile(w.path)
			sema.Release(1)
			if err != nil {
				return fmt.Errorf("could not read %q: %w", w.path, err)
			}

			if cwd != "" {
				if r, err := filepath.Rel(cwd, w.path); err == nil {
					w.path = r // Use relative path if possible
				}
			}
			errs, err := l.check(w.path, src, proj, proc, ac, rwc)
			if err != nil {
				return fmt.Errorf("fatal error while checking %s: %w", w.path, err)
			}
			w.src = src
			w.errs = errs
			return nil
		})
	}

	proc.wait()
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	total := 0
	for i := range ws {
		total += len(ws[i].errs)
	}

	all := make([]*Error, 0, total)
	if l.errFmt != nil {
		temp := make([]*ErrorTemplateFields, 0, total)
		for i := range ws {
			w := &ws[i]
			for _, err := range w.errs {
				temp = append(temp, err.GetTemplateFields(w.src))
			}
			all = append(all, w.errs...)
		}
		if err := l.errFmt.Print(l.out, temp); err != nil {
			return nil, err
		}
	} else {
		for i := range ws {
			w := &ws[i]
			l.printErrors(w.errs, w.src)
			all = append(all, w.errs...)
		}
	}

	l.log("Found", total, "errors in", n, "files")

	return all, nil
}

// LintFile lints one YAML file and outputs the errors to given writer. The project
// parameter can be nil. In the case, the project is detected from the given path.
func (l *Linter) LintFile(path string, project *Project) ([]*Error, error) {
	if project == nil {
		p, err := l.projects.At(path)
		if err != nil {
			return nil, err
		}
		project = p
	}

	src, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read %q: %w", path, err)
	}

	if l.cwd != "" {
		if r, err := filepath.Rel(l.cwd, path); err == nil {
			path = r
		}
	}

	proc := newConcurrentProcess(runtime.NumCPU())
	dbg := l.debugWriter()
	localActions := NewLocalActionsCache(project, dbg)
	localReusableWorkflows := NewLocalReusableWorkflowCache(project, l.cwd, dbg)
	errs, err := l.check(path, src, project, proc, localActions, localReusableWorkflows)
	proc.wait()
	if err != nil {
		return nil, err
	}

	if l.errFmt != nil {
		l.errFmt.PrintErrors(l.out, errs, src)
	} else {
		l.printErrors(errs, src)
	}
	return errs, err
}

// Lint lints YAML workflow file content given as byte slice. The path parameter is used as file
// path where the content came from. Setting "<stdin>" to path parameter indicates the output came
// from STDIN.
// When nil is passed to the project parameter, it tries to find the project from the path parameter.
func (l *Linter) Lint(path string, content []byte, project *Project) ([]*Error, error) {
	if project == nil && path != "<stdin>" {
		if _, err := os.Stat(path); !errors.Is(err, fs.ErrNotExist) {
			p, err := l.projects.At(path)
			if err != nil {
				return nil, err
			}
			project = p
		}
	}
	proc := newConcurrentProcess(runtime.NumCPU())
	dbg := l.debugWriter()
	localActions := NewLocalActionsCache(project, dbg)
	localReusableWorkflows := NewLocalReusableWorkflowCache(project, l.cwd, dbg)
	errs, err := l.check(path, content, project, proc, localActions, localReusableWorkflows)
	proc.wait()
	if err != nil {
		return nil, err
	}
	if l.errFmt != nil {
		l.errFmt.PrintErrors(l.out, errs, content)
	} else {
		l.printErrors(errs, content)
	}
	return errs, nil
}

func (l *Linter) check(
	path string,
	content []byte,
	project *Project,
	proc *concurrentProcess,
	localActions *LocalActionsCache,
	localReusableWorkflows *LocalReusableWorkflowCache,
) ([]*Error, error) {
	// Note: This method is called to check multiple files in parallel.
	// It must be thread safe assuming fields of Linter are not modified while running.

	var start time.Time
	if l.logLevel >= LogLevelVerbose {
		start = time.Now()
	}

	l.log("Linting", inputFormatString(l.inputFormat), path)
	if project != nil {
		l.log("Using project at", project.RootDir())
	}

	var cfg *Config
	if l.defaultConfig != nil {
		// `-config-file` option has higher prioritiy than repository config file
		cfg = l.defaultConfig
	} else if project != nil {
		cfg = project.Config()
	}
	if cfg != nil {
		l.debug("Config: %#v", cfg)
	} else {
		l.debug("No config was found")
	}

	w, a, sf, all := ParseFile(path, content, l.inputFormat)

	if l.logLevel >= LogLevelVerbose {
		elapsed := time.Since(start)
		l.log("Found", len(all), "parse errors in", elapsed.Milliseconds(), "ms for", inputFormatString(sf), path)
	}

	if w != nil || a != nil {
		dbg := l.debugWriter()

		var rules []Rule
		if w != nil {
			rules = []Rule{
				NewRuleMatrix(),
				NewRuleCredentials(),
				NewRuleShellName(),
				NewRuleRunnerLabel(),
				NewRuleEvents(),
				NewRuleJobNeeds(),
				NewRuleAction(localActions),
				NewRuleEnvVar(),
				NewRuleID(),
				NewRuleGlob(),
				NewRulePermissions(),
				NewRuleWorkflowCall(path, localReusableWorkflows),
				NewRuleExpression(localActions, localReusableWorkflows),
				NewRuleDeprecatedCommands(),
				NewRuleIfCond(),
			}
		} else {
			rules = []Rule{
				NewRuleBranding(),
				NewRuleShellName(),
				NewRuleAction(localActions),
				NewRuleEnvVar(),
				NewRuleID(),
				NewRuleWorkflowCall(path, localReusableWorkflows),
				NewRuleExpression(localActions, localReusableWorkflows),
				NewRuleDeprecatedCommands(),
				NewRuleIfCond(),
			}
		}
		if l.shellcheck != "" {
			r, err := NewRuleShellcheck(l.shellcheck, proc)
			if err == nil {
				rules = append(rules, r)
			} else {
				l.log("Rule \"shellcheck\" was disabled:", err)
			}
		} else {
			l.log("Rule \"shellcheck\" was disabled since shellcheck command name was empty")
		}
		if l.pyflakes != "" {
			r, err := NewRulePyflakes(l.pyflakes, proc)
			if err == nil {
				rules = append(rules, r)
			} else {
				l.log("Rule \"pyflakes\" was disabled:", err)
			}
		} else {
			l.log("Rule \"pyflakes\" was disabled since pyflakes command name was empty")
		}
		if l.onRulesCreated != nil {
			rules = l.onRulesCreated(rules)
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
		if cfg != nil {
			for _, r := range rules {
				r.SetConfig(cfg)
			}
		}

		if w != nil {
			if err := v.VisitWorkflow(w); err != nil {
				l.debug("error occurred while visiting workflow syntax tree: %v", err)
				return nil, err
			}
		}
		if a != nil {
			if err := v.VisitAction(a); err != nil {
				l.debug("error occurred while visiting workflow syntax tree: %v", err)
				return nil, err
			}
		}

		for _, rule := range rules {
			errs := rule.Errs()
			l.debug("%s found %d errors", rule.Name(), len(errs))
			all = append(all, errs...)
		}

		if l.errFmt != nil {
			for _, rule := range rules {
				l.errFmt.RegisterRule(rule)
			}
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

	for _, err := range all {
		err.Filepath = path // Populate filename in the error
	}

	sort.Stable(ByErrorPosition(all))

	if l.logLevel >= LogLevelVerbose {
		elapsed := time.Since(start)
		l.log("Found total", len(all), "errors in", elapsed.Milliseconds(), "ms for", inputFormatString(sf), path)
	}

	return all, nil
}

func (l *Linter) printErrors(errs []*Error, src []byte) {
	if l.oneline {
		src = nil
	}
	for _, err := range errs {
		err.PrettyPrint(l.out, src)
	}
}
