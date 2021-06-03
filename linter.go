package actionlint

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/kr/pretty"
)

func findNearestWorkflowsDir(from string) (string, error) {
	d := from
	for {
		p := filepath.Join(d, ".github", "workflows")
		if s, err := os.Stat(p); err == nil && s.IsDir() {
			return p, nil
		}

		n := filepath.Dir(d)
		if n == d {
			return "", fmt.Errorf("No .github/workflows directory was found in any parent directories of %q", from)
		}
		d = n
	}
}

type LogLevel int

const (
	LogLevelNone    LogLevel = 0
	LogLevelVerbose          = 1
	LogLevelDebug            = 2
)

type LinterOptions struct {
	Verbose   bool
	Debug     bool
	LogWriter io.Writer
	// More options will come here
}

type Linter struct {
	out      io.Writer
	logOut   io.Writer
	logLevel LogLevel
}

func NewLinter(out io.Writer, opts *LinterOptions) *Linter {
	l := LogLevelNone
	if opts.Verbose {
		l = LogLevelVerbose
	} else if opts.Debug {
		l = LogLevelDebug
	}

	var lout io.Writer = os.Stderr
	if opts.LogWriter != nil {
		lout = opts.LogWriter
	}

	return &Linter{out, lout, l}
}

func (l *Linter) Log(args ...interface{}) {
	if l.logLevel >= LogLevelVerbose {
		fmt.Fprintln(l.logOut, args...)
	}
}

func (l *Linter) DebugLog(args ...interface{}) {
	if l.logLevel >= LogLevelDebug {
		fmt.Fprintln(l.logOut, args...)
	}
}

// LintRepoDir lints YAML workflow files and outputs the errors to given writer. It finds the nearest
// `.github/workflow` directory based on `dir` and applies lint rules to all YAML worflow files
// under the directory.
func (l *Linter) LintRepoDir(dir string) ([]*Error, error) {
	l.Log("Linting all workflow files in repository:", dir)

	d, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("Could not get absolute path of %q: %w", dir, err)
	}

	wd, err := findNearestWorkflowsDir(d)
	if err != nil {
		return nil, err
	}
	l.Log("Detected workflows directory:", wd)

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
		return nil, fmt.Errorf("Could not read files in %q: %w", wd, err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("No YAML file was found in %q", wd)
	}
	l.Log("Collected", len(files), "YAML files")

	return l.LintFiles(files)
}

// LintFiles lints YAML workflow files and outputs the errors to given writer.
// It applies lint rules to all given files.
func (l *Linter) LintFiles(filepaths []string) ([]*Error, error) {
	n := len(filepaths)
	if n > 1 {
		l.Log("Linting", n, "files")
	}

	all := []*Error{}

	// TODO: Use multiple threads (per file)
	for _, p := range filepaths {
		errs, err := l.LintFile(p)
		if err != nil {
			return all, err
		}
		all = append(all, errs...)
	}
	if n > 1 {
		l.Log("Found", len(all), "errors in", n, "files")
	}

	return all, nil
}

func (l *Linter) LintFile(path string) ([]*Error, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Could not read %q: %w", path, err)
	}

	// Use relative path if possible
	if wd, err := os.Getwd(); err == nil {
		if r, err := filepath.Rel(wd, path); err == nil {
			path = r
		}
	}

	return l.Lint(path, b)
}

func (l *Linter) Lint(path string, content []byte) ([]*Error, error) {
	l.Log("Linting", path)

	w, all := Parse(content)

	if l.logLevel >= LogLevelDebug {
		fmt.Fprintln(l.logOut, "========== WORKFLOW TREE START ==========")
		pretty.Fprintf(l.logOut, "%# v\n", w)
		fmt.Fprintln(l.logOut, "=========== WORKFLOW TREE END ===========")
	}

	rules := []Rule{
		NewRuleMatrix(),
		NewRuleCredentials(),
		NewRuleShellName(),
		NewRuleEvents(),
		NewRuleJobNeeds(),
		NewRuleAction(path),
		NewRuleExpression(),
	}

	v := NewVisitor()
	for _, rule := range rules {
		v.AddPass(rule)
	}

	v.Visit(w)

	for _, rule := range rules {
		all = append(all, rule.Errs()...)
	}

	for _, err := range all {
		err.Filepath = path // Populate filename in the error
		fmt.Fprintln(l.out, err)
	}

	l.Log("Found", len(all), "errors in", path)

	return all, nil
}

func (l *Linter) printErrs(errs []*Error) {
	for _, err := range errs {
		fmt.Fprintln(l.out, err)
	}
}
