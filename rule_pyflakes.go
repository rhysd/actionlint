package actionlint

import (
	"bytes"
	"fmt"
	"sync"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sys/execabs"
)

type shellIsPythonKind int

const (
	shellIsPythonKindUnspecified shellIsPythonKind = iota
	shellIsPythonKindPython
	shellIsPythonKindNotPython
)

func getShellIsPythonKind(shell *String) shellIsPythonKind {
	if shell == nil {
		return shellIsPythonKindUnspecified
	}
	if shell.Value == "python" {
		return shellIsPythonKindPython
	}
	return shellIsPythonKindNotPython
}

// RulePyflakes is a rule to check Python scripts at 'run:' using pyflakes.
// https://github.com/PyCQA/pyflakes
type RulePyflakes struct {
	RuleBase
	cmd                   string
	workflowShellIsPython shellIsPythonKind
	jobShellIsPython      shellIsPythonKind
	group                 errgroup.Group
	mu                    sync.Mutex
	proc                  *concurrentProcess
}

// NewRulePyflakes creates new RulePyflakes instance. Parameter executable can be command name
// or relative/absolute file path. When the given executable is not found in system, it returns
// an error.
func NewRulePyflakes(executable string, proc *concurrentProcess) (*RulePyflakes, error) {
	p, err := execabs.LookPath(executable)
	if err != nil {
		return nil, err
	}
	r := &RulePyflakes{
		RuleBase:              RuleBase{name: "pyflakes"},
		cmd:                   p,
		workflowShellIsPython: shellIsPythonKindUnspecified,
		jobShellIsPython:      shellIsPythonKindUnspecified,
		proc:                  proc,
	}
	return r, nil
}

// VisitJobPre is callback when visiting Job node before visiting its children.
func (rule *RulePyflakes) VisitJobPre(n *Job) error {
	if n.Defaults != nil && n.Defaults.Run != nil {
		rule.jobShellIsPython = getShellIsPythonKind(n.Defaults.Run.Shell)
	}
	return nil
}

// VisitJobPost is callback when visiting Job node after visiting its children.
func (rule *RulePyflakes) VisitJobPost(n *Job) error {
	rule.jobShellIsPython = shellIsPythonKindUnspecified // reset
	return nil
}

// VisitWorkflowPre is callback when visiting Workflow node before visiting its children.
func (rule *RulePyflakes) VisitWorkflowPre(n *Workflow) error {
	if n.Defaults != nil && n.Defaults.Run != nil {
		rule.workflowShellIsPython = getShellIsPythonKind(n.Defaults.Run.Shell)
	}
	return nil
}

// VisitWorkflowPost is callback when visiting Workflow node after visiting its children.
func (rule *RulePyflakes) VisitWorkflowPost(n *Workflow) error {
	// TODO: Check errors caused in goroutines to run pyflakes and returns it

	// Wait all pyflakes processes finish
	if err := rule.group.Wait(); err != nil {
		return err
	}

	rule.workflowShellIsPython = shellIsPythonKindUnspecified // reset

	return nil
}

// VisitStep is callback when visiting Step node.
func (rule *RulePyflakes) VisitStep(n *Step) error {
	if rule.cmd == "" {
		return nil
	}

	run, ok := n.Exec.(*ExecRun)
	if !ok || run.Run == nil {
		return nil
	}

	if !rule.isPythonShell(run) {
		return nil
	}

	rule.group.Go(func() error {
		return rule.runPyflakes(rule.cmd, run.Run.Value, run.RunPos)
	})
	return nil
}

// Cleanup is callback when visiting finished. This callback is called even if the visiting failed since some callback returned an error
func (rule *RulePyflakes) Cleanup() {
	rule.group.Wait() // Ensure all processes ended
}

func (rule *RulePyflakes) isPythonShell(r *ExecRun) bool {
	if r.Shell != nil {
		return r.Shell.Value == "python"
	}

	if rule.jobShellIsPython != shellIsPythonKindUnspecified {
		return rule.jobShellIsPython == shellIsPythonKindPython
	}

	return rule.workflowShellIsPython == shellIsPythonKindPython
}

func (rule *RulePyflakes) runPyflakes(executable, src string, pos *Pos) error {
	src = sanitizeExpressionsInScript(src) // Defiend at rule_shellcheck.go
	rule.debug("%s: Running %s for Python script:\n%s", pos, executable, src)

	stdout, err := rule.proc.run(executable, []string{}, src)
	if err != nil {
		rule.debug("Command %s failed: %v", executable, err)
		return fmt.Errorf("`%s` did not run successfully while checking script at %s: %w", executable, pos, err)
	}
	if len(stdout) == 0 {
		return nil
	}

	rule.mu.Lock()
	defer rule.mu.Unlock()
	for len(stdout) > 0 {
		if stdout, err = rule.parseNextError(stdout, pos); err != nil {
			return err
		}
	}

	return nil
}

func (rule *RulePyflakes) parseNextError(stdout []byte, pos *Pos) ([]byte, error) {
	b := stdout

	// Eat "<stdin>:"
	idx := bytes.Index(b, []byte("<stdin>:"))
	if idx == -1 {
		return nil, fmt.Errorf("error message from pyflakes does not start with \"<stdin>:\" while checking script at %s. stdout:\n%s", pos, stdout)
	}
	b = b[idx+len("<stdin>:"):]

	var msg []byte
	if idx := bytes.Index(b, []byte("\r\n")); idx >= 0 {
		msg = b[:idx]
		b = b[idx+2:]
	} else if idx := bytes.IndexByte(b, '\n'); idx >= 0 {
		msg = b[:idx]
		b = b[idx+1:]
	} else {
		return nil, fmt.Errorf("error message from pyflakes does not end with \\n nor \\r\\n while checking script at %s. output: %q", pos, stdout)
	}
	rule.errorf(pos, "pyflakes reported issue in this script: %s", msg)

	return b, nil
}
