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
		return nil, fmt.Errorf("could not make stdin pipe for %s process while checking script: %w", exe, err)
	}
	if _, err := io.WriteString(p, stdin); err != nil {
		return nil, fmt.Errorf("could not write to stdin of %s process while checking script: %w", exe, err)
	}
	p.Close()

	return cmd.Output()
}
