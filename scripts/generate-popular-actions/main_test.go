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

// Normal cases

func TestDefaultPopularActions(t *testing.T) {
	popularActions, err := newGen(nil, nil, nil).registry()
	if err != nil {
		t.Fatal("could not load default popular actions:", err)
	}

	if len(popularActions) == 0 {
		t.Fatal("popularActions is empty")
	}

	slugs := map[string]int{}
	for i, a := range popularActions {
		if a.Slug == "" {
			t.Errorf("repository slug must not empty at popularActions[%d]", i)
		} else if j, ok := slugs[a.Slug]; ok && popularActions[i].Path == popularActions[j].Path {
			t.Errorf("duplicate registry. action %q at popularActions[%d] was already added at popularActions[%d]", a.Slug, i, j)
		} else {
			slugs[a.Slug] = i
		}

		if len(a.Tags) == 0 {
			t.Errorf("no tag is specified for %q", a.Slug)
		}

		tags := map[string]int{}
		for i, tag := range a.Tags {
			if tag == "" {
				t.Errorf("tags[%d] at action %q must not be empty string", i, a.Slug)
				continue
			}
			if tag == a.Next {
				t.Errorf("tags[%d] at action %q is equal to next version %q", i, a.Slug, a.Next)
			}
			if j, ok := tags[tag]; ok {
				t.Errorf("duplicate tag %q at action %q appears: tags[%d] v.s. tags[%d]", tag, a.Slug, i, j)
			} else {
				tags[tag] = i
			}
		}

		if a.FileExt != "yaml" && a.FileExt != "yml" && a.FileExt != "" {
			t.Errorf(`file ext of action %q is neither "yml" nor "yaml": %q`, a.Slug, a.FileExt)
		}
	}
}

func TestReadWriteJSONL(t *testing.T) {
	files := []string{
		"no_new_version.jsonl",
		"skip_inputs.jsonl",
		"skip_outputs.jsonl",
		"skip_both.jsonl",
		"skip_both.jsonl",
		"outdated.jsonl",
		"known_outdated.jsonl",
	}

	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			f := filepath.Join("testdata", "jsonl", file)
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			status := newGen(stdout, stderr, io.Discard).run([]string{"test", "-s", f, "-f", "jsonl"})
			if status != 0 {
				t.Fatalf("exit status is non-zero: %d: %s", status, stderr.Bytes())
			}

			b, err := os.ReadFile(f)
			if err != nil {
				panic(err)
			}
			want := string(b)
			have := stdout.String()

			if want != have {
				t.Fatalf("read content and output content differ\n%s", cmp.Diff(want, have))
			}
		})
	}
}

func TestWriteGoToStdout(t *testing.T) {
	testCases := []struct {
		in   string
		want string
	}{
		{
			in:   "no_new_version.jsonl",
			want: "no_new_version.go",
		},
		{
			in:   "skip_inputs.jsonl",
			want: "skip_inputs.go",
		},
		{
			in:   "skip_outputs.jsonl",
			want: "skip_outputs.go",
		},
		{
			in:   "skip_both.jsonl",
			want: "skip_both.go",
		},
		{
			in:   "outdated.jsonl",
			want: "outdated.go",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.in, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			status := newGen(stdout, stderr, io.Discard).run([]string{"test", "-s", filepath.Join("testdata", "jsonl", tc.in)})
			if status != 0 {
				t.Fatalf("exit status is non-zero: %d: %s", status, stderr.Bytes())
			}

			b, err := os.ReadFile(filepath.Join("testdata", "go", tc.want))
			if err != nil {
				panic(err)
			}
			want := string(b)
			have := stdout.String()

			if want != have {
				t.Fatalf("read content and output content differ\n%s", cmp.Diff(want, have))
			}
		})
	}
}

func TestWriteJSONLFile(t *testing.T) {
	in := filepath.Join("testdata", "jsonl", "no_new_version.jsonl")
	b, err := os.ReadFile(in)
	if err != nil {
		panic(err)
	}

	out := filepath.Join("testdata", "out.jsonl")
	defer os.Remove(out)

	stdout := io.Discard
	stderr := io.Discard
	status := newGen(stdout, stderr, io.Discard).run([]string{"test", "-s", in, "-f", "jsonl", out})
	if status != 0 {
		t.Fatal("exit status is non-zero:", status)
	}

	want := string(b)
	b, err = os.ReadFile(out)
	if err != nil {
		t.Fatalf("file was not created at %s: %s", out, err)
	}
	have := string(b)

	if want != have {
		t.Fatalf("read content and output content differ\n%s", cmp.Diff(want, have))
	}
}

func TestWriteGoFile(t *testing.T) {
	in := filepath.Join("testdata", "jsonl", "no_new_version.jsonl")
	out := filepath.Join("testdata", "go", "out.go")
	defer os.Remove(out)

	stdout := io.Discard
	stderr := io.Discard
	status := newGen(stdout, stderr, io.Discard).run([]string{"test", "-s", in, out})
	if status != 0 {
		t.Fatal("exit status is non-zero:", status)
	}

	b, err := os.ReadFile(filepath.Join("testdata", "go", "no_new_version.go"))
	if err != nil {
		panic(err)
	}
	want := string(b)

	b, err = os.ReadFile(out)
	if err != nil {
		t.Fatalf("file was not created at %s: %s", out, err)
	}
	have := string(b)

	if want != have {
		t.Fatalf("read content and output content differ\n%s", cmp.Diff(want, have))
	}
}

func TestFetchRemoteYAML(t *testing.T) {
	tests := []struct {
		registry string
		want     string
	}{
		{"fetch.json", "fetched.go"},
		{"outdated.json", "outdated.go"},
		{"skip_both.json", "skip_both.go"},
	}

	for _, tc := range tests {
		t.Run(tc.registry, func(t *testing.T) {
			f := filepath.Join("testdata", "registry", tc.registry)
			stdout := &bytes.Buffer{}
			stderr := io.Discard
			status := newGen(stdout, stderr, io.Discard).run([]string{"test", "-r", f})
			if status != 0 {
				t.Fatal("exit status is non-zero:", status)
			}

			b, err := os.ReadFile(filepath.Join("testdata", "go", tc.want))
			if err != nil {
				panic(err)
			}
			want := string(b)
			have := stdout.String()

			if !cmp.Equal(want, have) {
				t.Fatalf("fetched JSONL data does not match: %s", cmp.Diff(want, have))
			}
		})
	}
}

func TestWriteOutdatedActionAsJSONL(t *testing.T) {
	f := filepath.Join("testdata", "registry", "outdated.json")
	stdout := &bytes.Buffer{}
	stderr := io.Discard
	status := newGen(stdout, stderr, io.Discard).run([]string{"test", "-r", f, "-f", "jsonl"})
	if status != 0 {
		t.Fatal("exit status is non-zero:", status)
	}

	b, err := os.ReadFile(filepath.Join("testdata", "jsonl", "outdated.jsonl"))
	if err != nil {
		panic(err)
	}
	want, have := string(b), stdout.String()
	if want != have {
		t.Fatal(cmp.Diff(want, have))
	}
}

func TestLogOutput(t *testing.T) {
	f := filepath.Join("testdata", "jsonl", "no_new_version.jsonl")
	stdout := &bytes.Buffer{}
	logged := &bytes.Buffer{}
	status := newGen(stdout, io.Discard, logged).run([]string{"test", "-s", f, "-f", "jsonl"})
	if status != 0 {
		t.Fatal("exit status is non-zero:", status)
	}

	so := stdout.String()
	lo := logged.String()
	if so == "" {
		t.Fatal("stdout showed nothing")
	}
	if lo == "" {
		t.Fatal("log output showed nothing")
	}

	stdout = &bytes.Buffer{}
	logged = &bytes.Buffer{}
	status = newGen(stdout, io.Discard, logged).run([]string{"test", "-s", f, "-f", "jsonl", "-q"})
	if status != 0 {
		t.Fatal("exit status is non-zero:", status)
	}

	so = stdout.String()
	lo = logged.String()
	if so == "" {
		t.Fatal("stdout showed nothing")
	}
	if lo != "" {
		t.Fatal("-q did not suppress log output")
	}
}

func TestHelpOutput(t *testing.T) {
	stdout := io.Discard
	stderr := &bytes.Buffer{}
	status := newGen(stdout, stderr, io.Discard).run([]string{"test", "-help"})
	if status != 0 {
		t.Fatal("exit status is non-zero:", status)
	}
	out := stderr.String()
	if !strings.Contains(out, "Usage:") {
		t.Fatalf("usage header is not included in -help output: %q", out)
	}
}

func TestDetectNewRelease(t *testing.T) {
	f := filepath.Join("testdata", "registry", "new_release.json")
	stdout := &bytes.Buffer{}
	stderr := io.Discard
	status := newGen(stdout, stderr, io.Discard).run([]string{"test", "-d", "-r", f})
	if status != 2 {
		t.Fatal("exit status is not 2:", status)
	}
	out := stdout.String()
	want := "https://github.com/rhysd/action-setup-vim/tree/v1"
	if !strings.Contains(out, want) {
		t.Fatalf("expected URL %q is not included in stdout: %q", want, out)
	}
}

func TestDetectNoRelease(t *testing.T) {
	files := []string{
		"no_new_version.json",
		"empty_next_version.json",
	}

	for _, f := range files {
		t.Run(f, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := io.Discard
			p := filepath.Join("testdata", "registry", f)
			status := newGen(stdout, stderr, io.Discard).run([]string{"test", "-d", "-r", p})
			if status != 0 {
				t.Fatal("exit status is non-zero:", status)
			}
			out := stdout.String()
			if out != "No new release was found\n" {
				t.Fatalf("stdout is unexpected: %q", out)
			}
		})
	}
}

// Error cases

func TestCouldNotReadJSONLFile(t *testing.T) {
	testCases := []struct {
		file string
		want string
	}{
		{"this-file-does-not-exist.jsonl", "could not read file"},
		{"test.txt", `JSONL file name must end with ".jsonl"`},
		{"broken.jsonl", "could not parse line as JSON for action metadata"},
	}
	for _, tc := range testCases {
		t.Run(tc.file, func(t *testing.T) {
			f := filepath.Join("testdata", "jsonl", tc.file)
			stdout := io.Discard
			stderr := &bytes.Buffer{}

			status := newGen(stdout, stderr, io.Discard).run([]string{"test", "-s", f})
			if status == 0 {
				t.Fatal("exit status is unexpectedly zero")
			}

			msg := stderr.String()
			if !strings.Contains(msg, tc.want) {
				t.Fatalf("unexpected stderr: %q", msg)
			}
		})
	}
}

func TestCouldNotCreateOutputFile(t *testing.T) {
	f := filepath.Join("testdata", "jsonl", "no_new_version.jsonl")
	out := filepath.Join("testdata", "this-dir-does-not-exit", "foo.jsonl")
	stdout := io.Discard
	stderr := &bytes.Buffer{}

	status := newGen(stdout, stderr, io.Discard).run([]string{"test", "-s", f, "-f", "jsonl", out})
	if status == 0 {
		t.Fatal("exit status is unexpectedly zero")
	}

	msg := stderr.String()
	if !strings.Contains(msg, "could not open file to output") {
		t.Fatalf("unexpected stderr: %q", msg)
	}
}

type testErrorWriter struct{}

func (w testErrorWriter) Write(b []byte) (int, error) {
	return 0, errors.New("dummy error")
}

func TestWriteError(t *testing.T) {
	for _, format := range []string{"go", "jsonl"} {
		t.Run(format, func(t *testing.T) {
			f := filepath.Join("testdata", "jsonl", "no_new_version.jsonl")
			stdout := testErrorWriter{}
			stderr := &bytes.Buffer{}

			status := newGen(stdout, stderr, io.Discard).run([]string{"test", "-s", f, "-f", format})
			if status == 0 {
				t.Fatal("exit status is unexpectedly zero")
			}

			msg := stderr.String()
			if !strings.Contains(msg, "dummy error") {
				t.Fatalf("unexpected stderr: %q", msg)
			}
		})
	}
}

func TestCouldNotFetch(t *testing.T) {
	stderr := &bytes.Buffer{}
	f := filepath.Join("testdata", "registry", "repo_not_found.json")

	status := newGen(io.Discard, stderr, io.Discard).run([]string{"test", "-r", f})
	if status == 0 {
		t.Fatal("exit status is unexpectedly zero")
	}

	msg := stderr.String()
	if !strings.Contains(msg, "request was not successful") {
		t.Fatalf("unexpected stderr: %q", msg)
	}
}

func TestInvalidCommandArgs(t *testing.T) {
	testCases := []struct {
		args []string
		want string
	}{
		{[]string{"x", "-f", "foo"}, "invalid value for -f option: foo"},
		{[]string{"x", "aaa.go", "bbb.go"}, "this command takes one or zero argument but given: [aaa.go bbb.go]"},
		{[]string{"x", "-unknown-flag"}, "flag provided but not defined: -unknown-flag"},
	}

	for _, tc := range testCases {
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			stderr := &bytes.Buffer{}

			status := newGen(io.Discard, stderr, io.Discard).run(tc.args)
			if status == 0 {
				t.Fatal("exit status is unexpectedly zero")
			}

			msg := stderr.String()
			if !strings.Contains(msg, tc.want) {
				t.Fatalf("wanted %q in stderr: %q", tc.want, msg)
			}
		})
	}
}

func TestDetectErrorBadRequest(t *testing.T) {
	stdout := io.Discard
	stderr := &bytes.Buffer{}
	f := filepath.Join("testdata", "registry", "empty_slug.json")
	status := newGen(stdout, stderr, io.Discard).run([]string{"test", "-d", "-r", f})
	if status != 1 {
		t.Fatal("exit status is not 1:", status)
	}
	out := stderr.String()
	if !strings.Contains(out, "head request for https://raw.githubusercontent.com//v2/action.yml was not successful") {
		t.Fatalf("stderr was unexpected: %q", out)
	}
}

func TestReadActionRegistryError(t *testing.T) {
	tests := []struct {
		file string
		want string
	}{
		{
			file: "broken.json",
			want: "could not parse the local action registry file as JSON:",
		},
		{
			file: "oops-this-file-doesnt-exist.json",
			want: "could not read the file for actions registry:",
		},
	}

	for _, tc := range tests {
		t.Run(tc.file, func(t *testing.T) {
			stdout := io.Discard
			stderr := &bytes.Buffer{}
			f := filepath.Join("testdata", "registry", tc.file)
			status := newGen(stdout, stderr, io.Discard).run([]string{"test", "-d", "-r", f})
			if status != 1 {
				t.Fatal("exit status is not 1:", status)
			}
			out := stderr.String()
			if !strings.Contains(out, tc.want) {
				t.Fatalf("wanted %q in stderr: %q", tc.want, out)
			}
		})
	}
}

func TestActionBuildRawURL(t *testing.T) {
	a := &registry{Slug: "foo/bar"}
	have := a.rawURL("v1")
	want := "https://raw.githubusercontent.com/foo/bar/v1/action.yml"
	if have != want {
		t.Errorf("Wanted %q but have %q", want, have)
	}

	a = &registry{Slug: "foo/bar", Path: "/a/b"}
	have = a.rawURL("v1")
	want = "https://raw.githubusercontent.com/foo/bar/v1/a/b/action.yml"
	if have != want {
		t.Errorf("Wanted %q but have %q", want, have)
	}

	a = &registry{Slug: "foo/bar", FileExt: "yaml"}
	have = a.rawURL("v1")
	want = "https://raw.githubusercontent.com/foo/bar/v1/action.yaml"
	if have != want {
		t.Errorf("Wanted %q but have %q", want, have)
	}
}

func TestActionBuildGitHubURL(t *testing.T) {
	a := &registry{Slug: "foo/bar"}
	have := a.githubURL("v1")
	want := "https://github.com/foo/bar/tree/v1"
	if have != want {
		t.Errorf("Wanted %q but have %q", want, have)
	}

	a = &registry{Slug: "foo/bar", Path: "/a/b"}
	have = a.githubURL("v1")
	want = "https://github.com/foo/bar/tree/v1/a/b"
	if have != want {
		t.Errorf("Wanted %q but have %q", want, have)
	}
}

func TestActionBuildSpec(t *testing.T) {
	a := &registry{Slug: "foo/bar"}
	have := a.spec("v1")
	want := "foo/bar@v1"
	if have != want {
		t.Errorf("Wanted %q but have %q", want, have)
	}

	a = &registry{Slug: "foo/bar", Path: "/a/b"}
	have = a.spec("v1")
	want = "foo/bar/a/b@v1"
	if have != want {
		t.Errorf("Wanted %q but have %q", want, have)
	}
}
