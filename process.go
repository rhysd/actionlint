package actionlint

import (
	"fmt"
	"io"
	"os/exec"
	"sync"
)

type processRunRequest struct {
	exe      string
	args     []string
	stdin    string
	callback func([]byte, error) error
}

// concurrentProcess is a manager to run process concurrently. Since running process consumes OS
// resources, running too many processes concurrently causes some issues. On macOS, making new
// process hangs (see issue #3). And also running processes which opens files causes an error
// "pipe: too many files to open". To avoid it, this class manages how many processes are run at
// the same time.
type concurrentProcess struct {
	wg         sync.WaitGroup
	ch         chan *processRunRequest
	done       chan struct{}
	numWorkers int
	maxWorkers int
	err        error
	errMu      sync.Mutex
}

func newConcurrentProcess(par int) *concurrentProcess {
	return &concurrentProcess{
		ch:         make(chan *processRunRequest),
		done:       make(chan struct{}),
		maxWorkers: par,
	}
}

func runProcessWithStdin(exe string, args []string, stdin string) ([]byte, error) {
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

func (proc *concurrentProcess) handleRequest(req *processRunRequest) {
	stdout, err := runProcessWithStdin(req.exe, req.args, req.stdin)
	if err := req.callback(stdout, err); err != nil {
		proc.errMu.Lock()
		if proc.err == nil {
			proc.err = err
		}
		proc.errMu.Unlock()
	}
}

func (proc *concurrentProcess) newWorker() {
	proc.numWorkers++
	proc.wg.Add(1)
	go func(recv <-chan *processRunRequest, done <-chan struct{}) {
		for {
			select {
			case req := <-recv:
				proc.handleRequest(req)
			case <-done:
				proc.wg.Done()
				return
			}
		}
	}(proc.ch, proc.done)
}

func (proc *concurrentProcess) hasError() bool {
	proc.errMu.Lock()
	defer proc.errMu.Unlock()
	return proc.err != nil
}

func (proc *concurrentProcess) run(exe string, args []string, stdin string, callback func([]byte, error) error) {
	// This works fine since it is single-producer-multi-consumers
	if proc.numWorkers < proc.maxWorkers {
		// Defer making new workers. This is efficient when no worker is necessary. It happens when shellcheck
		// and pyflakes are never run (e.g. they're disabled).
		proc.newWorker()
	}
	proc.ch <- &processRunRequest{exe, args, stdin, callback}
}

func (proc *concurrentProcess) wait() error {
	close(proc.done) // Request workers to shutdown
	proc.wg.Wait()   // Wait for workers completing to shutdown
	proc.numWorkers = 0
	return proc.err
}
