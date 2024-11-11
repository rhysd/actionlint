package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func testRunMain(args []string) (string, string, int) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	status := run(args, stdout, stderr, io.Discard, "")
	return stdout.String(), stderr.String(), status
}

func TestOKWriteStdout(t *testing.T) {
	f := filepath.Join("testdata", "ok.md")
	stdout, stderr, status := testRunMain([]string{f, "-"})
	if status != 0 {
		t.Fatalf("status was non-zero: %d: %q", status, stderr)
	}

	b, err := os.ReadFile(filepath.Join("testdata", "ok.go"))
	if err != nil {
		t.Fatal(err)
	}
	want := string(b)

	if stdout != want {
		t.Fatal(cmp.Diff(want, stdout))
	}
}

func TestOKWriteFile(t *testing.T) {
	in := filepath.Join("testdata", "ok.md")
	out := filepath.Join("testdata", "_test_output.go")
	defer os.Remove(out)

	stdout, stderr, status := testRunMain([]string{in, out})
	if status != 0 {
		t.Fatalf("status was non-zero: %d: %q", status, stderr)
	}

	b, err := os.ReadFile(filepath.Join("testdata", "ok.go"))
	if err != nil {
		t.Fatal(err)
	}
	want := string(b)

	if stdout != "" {
		t.Fatalf("stdout is not empty: %q", stdout)
	}

	b, err = os.ReadFile(out)
	if err != nil {
		t.Fatalf("output file %q cannot be read: %v", out, err)
	}
	have := string(b)

	if want != have {
		t.Fatal(cmp.Diff(want, have))
	}
}

func TestFetchOK(t *testing.T) {
	url := "https://raw.githubusercontent.com/rhysd/actionlint/main/LICENSE.txt"
	b, err := fetch(url)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(b, []byte("the MIT License")) {
		t.Fatalf("unexpected content: %q", b)
	}
}

func TestErrorGenerate(t *testing.T) {
	testCases := []struct {
		file string
		want string
	}{
		{"no_heading.md", "heading was missing"},
		{"no_hooks.md", "no webhook table was found in given markdown source"},
		{"no_hook_name_link.md", "\"Webhook event payload\" table was found, but first cell did not contain hook name"},
	}

	for _, tc := range testCases {
		t.Run(tc.file, func(t *testing.T) {
			f := filepath.Join("testdata", tc.file)
			stdout, stderr, status := testRunMain([]string{f, "-"})
			if status == 0 {
				t.Fatalf("status was zero: %q", stdout)
			}
			if !strings.Contains(stderr, tc.want) {
				t.Fatalf("wanted %q in stderr %q", tc.want, stderr)
			}
		})
	}
}

type testErrorWriter struct{}

func (w testErrorWriter) Write(b []byte) (int, error) {
	return 0, errors.New("dummy write error")
}

func TestErrorWriteResult(t *testing.T) {
	f := filepath.Join("testdata", "ok.md")
	stderr := &bytes.Buffer{}
	status := run([]string{f, "-"}, testErrorWriter{}, stderr, io.Discard, "")
	if status == 0 {
		t.Fatal("status was zero")
	}
	msg := stderr.String()
	if !strings.Contains(msg, "dummy write error") {
		t.Fatalf("write error did not occur: %q", msg)
	}
}

func TestFetchError(t *testing.T) {
	testCases := []struct {
		what string
		url  string
		want string
	}{
		{"not found", "https://raw.githubusercontent.com/rhysd/actionlint/main/this-file-does-not-exist.txt", "request was not successful"},
		{"invalid url", "foo://bar", "could not fetch"},
	}

	for _, tc := range testCases {
		t.Run(tc.what, func(t *testing.T) {
			stderr := &bytes.Buffer{}
			status := run([]string{"-"}, io.Discard, stderr, io.Discard, tc.url)
			if status == 0 {
				t.Fatal("status was zero")
			}
			msg := stderr.String()
			if !strings.Contains(msg, tc.want) {
				t.Fatalf("unexpected error: %v: %s", msg, tc.url)
			}
		})
	}
}

func TestCmdError(t *testing.T) {
	f := filepath.Join("testdata", "ok.md")
	dirNotExist := filepath.Join("dir", "does", "not", "exist", "out.go")
	testCases := []struct {
		what string
		args []string
		want string
	}{
		{"too many args", []string{"foo", "bar", "piyo"}, "usage:"},
		{"cannot read file", []string{"oops-this-file-does-not-exist.md", "-"}, "oops-this-file-does-not-exist.md"},
		{"cannot write file", []string{f, dirNotExist}, dirNotExist},
	}

	for _, tc := range testCases {
		t.Run(tc.what, func(t *testing.T) {
			stdout, stderr, status := testRunMain(tc.args)
			if status == 0 {
				t.Fatalf("status was zero: %q", stdout)
			}
			if !strings.Contains(stderr, tc.want) {
				t.Fatalf("stderr does not contain %q: %q", tc.want, stderr)
			}
		})
	}
}
