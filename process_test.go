package actionlint

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

func testStartEchoCommand(t *testing.T, proc *concurrentProcess, done *bool) {
	*done = false
	echo := testSkipIfNoCommand(t, proc, "echo")
	echo.run([]string{}, "", func(b []byte, err error) error {
		if err != nil {
			t.Error(err)
			return err
		}
		*done = true
		return nil
	})
	// This function does not wait the command finishes
}

func testSkipIfNoCommand(t *testing.T, p *concurrentProcess, cmd string) *externalCommand {
	c, err := p.newCommandRunner(cmd)
	if err != nil {
		t.Skipf("%s command is necessary to run this test: %s", cmd, err)
	}
	return c
}

func TestProcessRunProcessSerial(t *testing.T) {
	p := newConcurrentProcess(1)
	ret := []string{}
	mu := sync.Mutex{}
	starts := []time.Time{}
	ends := []time.Time{}
	numProcs := 3
	echo := testSkipIfNoCommand(t, p, "echo")

	for i := 0; i < numProcs; i++ {
		in := fmt.Sprintf("message %d", i)
		echo.run([]string{in}, "", func(b []byte, err error) error {
			mu.Lock()
			defer mu.Unlock()

			starts = append(starts, time.Now())
			defer func() {
				ends = append(ends, time.Now())
			}()

			if err != nil {
				t.Error(err)
				return err
			}

			ret = append(ret, string(b))
			return nil
		})
	}

	if err := echo.wait(); err != nil {
		t.Fatal(err)
	}
	p.wait()

	if len(ret) != numProcs {
		t.Fatalf("wanted %d outputs but got %#v", numProcs, ret)
	}

	// Check error messages
	for i := 0; i < numProcs; i++ {
		e := fmt.Sprintf("message %d", i)
		if !strings.HasPrefix(ret[i], e) {
			t.Fatalf("ret[%d] does not start with %q: %#v", i, e, ret)
		}
	}

	starts = starts[1:]
	ends = ends[:len(ends)-1]

	for i, s := range starts {
		e := ends[i]
		if s.Before(e) {
			t.Errorf("Command #%d started at %s before previous command #%d ends at %s", i+1, s, i, e)
		}
	}
}

func TestProcessRunConcurrently(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("this test is flaky on Windows")
	}

	p := newConcurrentProcess(5)
	sleep := testSkipIfNoCommand(t, p, "sleep")

	start := time.Now()
	for i := 0; i < 5; i++ {
		sleep.run([]string{"0.1"}, "", func(b []byte, err error) error {
			if err != nil {
				t.Error(err)
				return err
			}
			return nil
		})
	}
	if err := sleep.wait(); err != nil {
		t.Fatal(err)
	}
	p.wait()

	sec := time.Since(start).Seconds()
	if sec >= 0.5 {
		t.Fatalf("commands did not run concurrently. running five `sleep 0.1` commands took %v seconds", sec)
	}
}

func TestProcessRunMultipleCommandsConcurrently(t *testing.T) {
	p := newConcurrentProcess(3)

	done := make([]bool, 5)
	cmds := make([]*externalCommand, 0, 5)
	for i := 0; i < 5; i++ {
		idx := i
		echo := testSkipIfNoCommand(t, p, "echo")
		echo.run([]string{"hello"}, "", func(b []byte, err error) error {
			if err != nil {
				t.Error(err)
				return err
			}
			done[idx] = true
			return nil
		})
		cmds = append(cmds, echo)
	}

	for i, c := range cmds {
		if err := c.wait(); err != nil {
			t.Errorf("cmds[%d] failed: %s", i, err)
		}
	}

	for i, b := range done {
		if !b {
			t.Errorf("cmds[%d] did not finish", i)
		}
	}
}

func TestProcessWaitMultipleCommandsFinish(t *testing.T) {
	p := newConcurrentProcess(2)

	done := make([]bool, 3)
	for i := 0; i < 3; i++ {
		idx := i
		echo := testSkipIfNoCommand(t, p, "echo")
		echo.run([]string{"hello"}, "", func(b []byte, err error) error {
			if err != nil {
				t.Error(err)
				return err
			}
			done[idx] = true
			return nil
		})
	}

	p.wait()

	for i, b := range done {
		if !b {
			t.Errorf("cmds[%d] did not finish", i)
		}
	}
}

func TestProcessInputStdin(t *testing.T) {
	p := newConcurrentProcess(1)
	cat := testSkipIfNoCommand(t, p, "cat")
	out := ""

	cat.run([]string{}, "this is test", func(b []byte, err error) error {
		if err != nil {
			t.Error(err)
			return err
		}
		out = string(b)
		return nil
	})

	if err := cat.wait(); err != nil {
		t.Fatal(err)
	}
	p.wait()

	if out != "this is test" {
		t.Fatalf("stdin was not input to `cat` command: %q", out)
	}
}

func TestProcessErrorCommandNotFound(t *testing.T) {
	p := newConcurrentProcess(1)
	c := &externalCommand{
		proc: p,
		exe:  "this-command-does-not-exist",
	}

	c.run([]string{}, "", func(b []byte, err error) error {
		if err != nil {
			return fmt.Errorf("yay! error found! %w", err)
		}
		t.Error("command not found error did not occur")
		return nil
	})

	var echoDone bool
	testStartEchoCommand(t, p, &echoDone)

	err := c.wait()
	if err == nil || !strings.Contains(err.Error(), "yay! error found!") {
		t.Fatalf("error was not reported by p.Wait(): %v", err)
	}

	p.wait()

	if !echoDone {
		t.Fatal("a command following the error did not run")
	}
}

func TestProcessErrorInCallback(t *testing.T) {
	p := newConcurrentProcess(1)
	echo := testSkipIfNoCommand(t, p, "echo")

	echo.run([]string{}, "", func(b []byte, err error) error {
		if err != nil {
			t.Error(err)
			return err
		}
		return fmt.Errorf("dummy error")
	})

	var echoDone bool
	testStartEchoCommand(t, p, &echoDone)

	err := echo.wait()
	if err == nil || err.Error() != "dummy error" {
		t.Fatalf("error was not reported by p.Wait(): %v", err)
	}

	p.wait()

	if !echoDone {
		t.Fatal("a command following the error did not run")
	}
}

func TestProcessErrorLinterFailed(t *testing.T) {
	p := newConcurrentProcess(1)
	ls := testSkipIfNoCommand(t, p, "ls")

	// Running ls with directory which does not exist emulates external liter's failure.
	// For example shellcheck exits with non-zero status but it outputs nothing to stdout when it
	// fails to run.
	ls.run([]string{"oops-this-directory-does-not-exist"}, "", func(b []byte, err error) error {
		if err != nil {
			return err
		}
		t.Error("error did not occur on running the process")
		return nil
	})

	var echoDone bool
	testStartEchoCommand(t, p, &echoDone)

	err := ls.wait()
	if err == nil {
		t.Fatal("error did not occur")
	}
	msg := err.Error()
	if !strings.Contains(msg, "but stdout was empty") || !strings.Contains(msg, "oops-this-directory-does-not-exist") {
		t.Fatalf("Error message was unexpected: %q", msg)
	}

	p.wait()

	if !echoDone {
		t.Fatal("a command following the error did not run")
	}
}

func TestProcessRunConcurrentlyAndWait(t *testing.T) {
	p := newConcurrentProcess(2)
	ls := testSkipIfNoCommand(t, p, "ls")

	c := make(chan struct{})
	go func() {
		for i := 0; i < 5; i++ {
			ls.run(nil, "", func(b []byte, err error) error {
				return err
			})
		}
		c <- struct{}{}
	}()
	<-c
	p.wait()
}
