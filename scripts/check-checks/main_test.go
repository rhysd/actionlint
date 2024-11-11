package main

import (
	"bytes"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// Do not mess up the test output. Temporarily commenting out this function may be helpful when debugging some test cases.
func init() {
	log.SetOutput(io.Discard)
	stderr = io.Discard
}

func testErr(t *testing.T, err error, want ...string) {
	t.Helper()
	if err == nil {
		t.Fatal("error did not occur")
	}
	msg := err.Error()
	for _, w := range want {
		if !strings.Contains(msg, w) {
			t.Errorf("error message %q does not contain expected text %q", msg, w)
		}
	}
}

func TestMainGenerateOK(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("check-checks doesn't support Windows")
	}
	root := t.TempDir()

	in, err := os.Open(filepath.FromSlash("testdata/ok/minimal.in"))
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.FromSlash(root + "/minimal.in")
	tmp, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = io.Copy(tmp, in)
	if err != nil {
		t.Fatal(err)
	}
	in.Close()
	tmp.Close()

	if err := Main([]string{"exe", "-fix", path}); err != nil {
		t.Fatal(err)
	}

	want, err := os.ReadFile(filepath.FromSlash("testdata/ok/minimal.out"))
	if err != nil {
		t.Fatal(err)
	}
	have, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(want, have) {
		t.Fatal(cmp.Diff(want, have))
	}
}

func TestMainCheckOK(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("check-checks doesn't support Windows")
	}
	path := filepath.FromSlash("testdata/ok/minimal.out")
	if err := Main([]string{"exe", path}); err != nil {
		t.Fatal(err)
	}
}

func TestMainCheckQuietOK(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("check-checks doesn't support Windows")
	}
	path := filepath.FromSlash("testdata/ok/minimal.out")
	if err := Main([]string{"exe", "-quiet", path}); err != nil {
		t.Fatal(err)
	}
}

func TestMainPrintHelp(t *testing.T) {
	if err := Main([]string{"exe", "-help"}); err != nil {
		t.Fatal(err)
	}
}

func TestMainCheckError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("check-checks doesn't support Windows")
	}
	path := filepath.FromSlash("testdata/ok/minimal.in")
	testErr(t, Main([]string{"exe", path}), "checks document has some update")
}

func TestMainFileNotFound(t *testing.T) {
	testErr(t, Main([]string{"exe", "this-file-does-not-exist.md"}), "could not read the document file")
}

func TestMainTooManyArgs(t *testing.T) {
	testErr(t, Main([]string{"exe", "a", "b", "c"}), "this command should take exact one file path but got")
}

func TestMainInvalidCheckFlag(t *testing.T) {
	testErr(t, Main([]string{"exe", "-c", "foo.md"}), "flag provided but not defined")
}

func TestMainUpdateError(t *testing.T) {
	path := filepath.FromSlash("testdata/err/no_playground_link.md")
	if err := Main([]string{"exe", path}); err == nil {
		t.Fatal("no error occurred")
	}
}

func TestUpdateOK(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("check-checks doesn't support Windows")
	}

	dir := filepath.FromSlash("testdata/ok")

	tests := []string{}
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range dirEntries {
		n := e.Name()
		if !strings.HasSuffix(n, ".in") {
			continue
		}
		tests = append(tests, strings.TrimSuffix(n, filepath.Ext(n)))
	}

	for _, tc := range tests {
		in := filepath.Join(dir, tc+".in")
		out := filepath.Join(dir, tc+".out")
		t.Run(tc, func(t *testing.T) {
			inBytes, err := os.ReadFile(in)
			if err != nil {
				t.Fatal(err)
			}
			have, err := Update(inBytes)
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile(out)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(want, have) {
				t.Fatal(cmp.Diff(want, have))
			}
		})
	}
}

func TestUpdateError(t *testing.T) {
	dir := filepath.FromSlash("testdata/err")

	tests := []string{}
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range dirEntries {
		n := e.Name()
		if !strings.HasSuffix(n, ".md") {
			continue
		}
		tests = append(tests, strings.TrimSuffix(n, filepath.Ext(n)))
	}

	for _, tc := range tests {
		path := filepath.Join(dir, tc+".md")
		t.Run(tc, func(t *testing.T) {
			inBytes, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := Update(inBytes); err == nil {
				t.Fatal("no error occurred")
			}
		})
	}
}

func TestStateString(t *testing.T) {
	tests := []struct {
		state state
		want  string
	}{
		{stateInit, "init"},
		{stateAnchor, "anchor"},
		{stateHeading, "heading"},
		{stateInputHeader, "input header"},
		{stateInputBlock, "input block"},
		{stateAfterInput, "after input"},
		{stateOutputHeader, "output header"},
		{stateAfterOutput, "after output"},
		{stateOutputBlock, "output block"},
		{stateEnd, "end"},
	}

	for _, tc := range tests {
		have := tc.state.String()
		if have != tc.want {
			t.Errorf("wanted %q for state %d but have %q", tc.want, tc.state, have)
		}
	}
}
