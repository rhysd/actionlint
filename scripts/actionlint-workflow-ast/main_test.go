package main

import (
	"bytes"
	"errors"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

func TestViewASTOfStdin(t *testing.T) {
	stdin := strings.NewReader(`on: push

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: echo hi`)
	stdout := &bytes.Buffer{}
	stderr := ioutil.Discard
	s := run([]string{"actionlint-workflow-ast"}, stdin, stdout, stderr)
	if s != 0 {
		t.Fatal("exit status is non-zero:", s)
	}
	out := stdout.String()
	if !strings.HasPrefix(out, "&actionlint.Workflow") {
		t.Fatal("unexpected output:", out)
	}
	if !strings.Contains(out, `"echo hi"`) {
		t.Fatal("unexpected output:", out)
	}
}

func TestViewASTOfFile(t *testing.T) {
	file := filepath.Join("..", "..", "testdata", "bench", "minimal.yaml")
	stdin := strings.NewReader("")
	stdout := &bytes.Buffer{}
	stderr := ioutil.Discard
	s := run([]string{"actionlint-workflow-ast", file}, stdin, stdout, stderr)
	if s != 0 {
		t.Fatal("exit status is non-zero:", s)
	}
	out := stdout.String()
	if !strings.HasPrefix(out, "&actionlint.Workflow") {
		t.Fatal("unexpected output:", out)
	}
	if !strings.Contains(out, `"echo hi"`) {
		t.Fatal("unexpected output:", out)
	}
}

func TestViewASTShowHelp(t *testing.T) {
	stdin := strings.NewReader("")
	stdout := &bytes.Buffer{}
	stderr := ioutil.Discard
	s := run([]string{"actionlint-workflow-ast", "-h"}, stdin, stdout, stderr)
	if s != 0 {
		t.Fatal("exit status is non-zero:", s)
	}
	out := stdout.String()
	if !strings.Contains(out, "Usage:") {
		t.Fatalf("unexpected stdout: %q", out)
	}
}

func TestViewASTErrorFileDoesNotExist(t *testing.T) {
	stdin := strings.NewReader("")
	stdout := ioutil.Discard
	stderr := &bytes.Buffer{}
	s := run([]string{"actionlint-workflow-ast", "this-file-does-not-exist.yaml"}, stdin, stdout, stderr)
	if s == 0 {
		t.Fatal("exit status is zero")
	}
	out := stderr.String()
	if out == "" {
		t.Fatal("stderr is empty")
	}
}

func TestViewASTErrorParseFailed(t *testing.T) {
	stdin := strings.NewReader(`on: push

jobs:
  test:
    steps:`)
	stdout := ioutil.Discard
	stderr := &bytes.Buffer{}
	s := run([]string{"actionlint-workflow-ast"}, stdin, stdout, stderr)
	if s == 0 {
		t.Fatal("exit status is zero")
	}
	out := stderr.String()
	for _, want := range []string{
		`"steps" section must be sequence node`,
		`"steps" section is missing`,
		`"runs-on" section is missing`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("wanted %q in stderr: %s", want, out)
		}
	}
}

type alwaysFail struct{}

func (r alwaysFail) Read(p []byte) (int, error) {
	return 0, errors.New("dummy error")
}

func TestViewASTErrorFailToReadStdin(t *testing.T) {
	stdin := alwaysFail{}
	stdout := ioutil.Discard
	stderr := &bytes.Buffer{}
	s := run([]string{"actionlint-workflow-ast"}, stdin, stdout, stderr)
	if s == 0 {
		t.Fatal("exit status is zero")
	}
	out := stderr.String()
	if !strings.Contains(out, "dummy error") {
		t.Fatal("unexpected stderr:", out)
	}
}
