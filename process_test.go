package actionlint

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/sys/execabs"
)

func TestProcessRunProcessSerial(t *testing.T) {
	p := newConcurrentProcess(1)
	ret := []string{}
	mu := sync.Mutex{}
	starts := []time.Time{}
	ends := []time.Time{}
	numProcs := 3

	for i := 0; i < numProcs; i++ {
		in := fmt.Sprintf("message %d", i)
		p.run("echo", []string{in}, "", func(b []byte, err error) error {
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

	if err := p.wait(); err != nil {
		t.Fatal(err)
	}

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
	if _, err := execabs.LookPath("sleep"); err != nil {
		t.Skipf("sleep command is necessary to run this test: %s", err)
	}

	p := newConcurrentProcess(5)

	start := time.Now()
	for i := 0; i < 5; i++ {
		p.run("sleep", []string{"0.1"}, "", func(b []byte, err error) error {
			if err != nil {
				t.Error(err)
				return err
			}
			return nil
		})
	}
	if err := p.wait(); err != nil {
		t.Fatal(err)
	}

	sec := time.Since(start).Seconds()
	if sec >= 0.5 {
		t.Fatalf("commands did not run concurrently. running five `sleep 0.1` commands took %v seconds", sec)
	}
}

func TestProcessInputStdin(t *testing.T) {
	if _, err := execabs.LookPath("cat"); err != nil {
		t.Skipf("cat command is necessary to run this test: %s", err)
	}

	p := newConcurrentProcess(1)
	out := ""

	p.run("cat", []string{}, "this is test", func(b []byte, err error) error {
		if err != nil {
			t.Error(err)
			return err
		}
		out = string(b)
		return nil
	})

	if err := p.wait(); err != nil {
		t.Fatal(err)
	}

	if out != "this is test" {
		t.Fatalf("stdin was not input to `cat` command: %q", out)
	}
}

func TestProcessErrorCommandNotFound(t *testing.T) {
	p := newConcurrentProcess(1)

	p.run("this-command-does-not-exist", []string{}, "", func(b []byte, err error) error {
		if err != nil {
			return fmt.Errorf("yay! error found! %w", err)
		}
		t.Error("command not found error did not occur")
		return nil
	})

	successorDone := false
	p.run("echo", []string{}, "", func(b []byte, err error) error {
		if err != nil {
			t.Error(err)
			return err
		}
		successorDone = true
		return nil
	})

	err := p.wait()
	if err == nil || !strings.Contains(err.Error(), "yay! error found!") {
		t.Fatalf("error was not reported by p.wait(): %v", err)
	}

	if !successorDone {
		t.Fatal("a command following the error did not run")
	}
}

func TestProcessErrorInCallback(t *testing.T) {
	p := newConcurrentProcess(1)

	p.run("echo", []string{}, "", func(b []byte, err error) error {
		if err != nil {
			t.Error(err)
			return err
		}
		return fmt.Errorf("dummy error")
	})

	successorDone := false
	p.run("echo", []string{}, "", func(b []byte, err error) error {
		if err != nil {
			t.Error(err)
			return err
		}
		successorDone = true
		return nil
	})

	err := p.wait()
	if err == nil || err.Error() != "dummy error" {
		t.Fatalf("error was not reported by p.wait(): %v", err)
	}

	if !successorDone {
		t.Fatal("a command following the error did not run")
	}
}
