package actionlint

import (
	"fmt"
	"io"
	"os/exec"
)

// concurrentProcess is a manager to run process concurrently. Since running process consumes OS
// resources, running too many processes concurrently causes some issues. On macOS, making new
// process hangs (see issue #3). And also running processes which opens files causes an error
// "pipe: too many files to open". To avoid it, this class manages how many processes are run at
// the same time.
//
// TODO: Reduce number of goroutines by preparing workers in this struct.
type concurrentProcess struct {
	guard chan struct{}
}

func newConcurrentProcess(par int) *concurrentProcess {
	return &concurrentProcess{make(chan struct{}, par)}
}

func (proc *concurrentProcess) lock() {
	proc.guard <- struct{}{}
}

func (proc *concurrentProcess) unlock() {
	<-proc.guard
}

func (proc *concurrentProcess) run(exe string, args []string, stdin string) ([]byte, error) {
	proc.lock()
	defer proc.unlock()

	cmd := exec.Command(exe, args...)
	cmd.Stderr = nil

	p, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("could not make stdin pipe for %s process: %w", exe, err)
	}
	if _, err := io.WriteString(p, stdin); err != nil {
		p.Close()
		return nil, fmt.Errorf("could not write to stdin of %s process: %w", exe, err)
	}
	p.Close()

	stdout, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code := exitErr.ExitCode()
			if code < 0 {
				return nil, fmt.Errorf("%s was terminated. stderr: %q", exe, exitErr.Stderr)
			}
			if len(stdout) == 0 {
				return nil, fmt.Errorf("%s exited with status %d but stdout was empty. stderr: %q", exe, code, exitErr.Stderr)
			}
			// Reaches here when exit status is non-zero and stdout is not empty, shellcheck successfully found some errors
		} else {
			return nil, err
		}
	}

	return stdout, nil
}
