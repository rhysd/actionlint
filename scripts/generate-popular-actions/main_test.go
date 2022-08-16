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

var testDummyPopularActions = []*action{
	{
		slug: "rhysd/action-setup-vim",
		tags: []string{"v1"},
		next: "v2",
	},
}

// Normal cases

func TestYAMLExtString(t *testing.T) {
	if yamlExtYML.String() != "yml" {
		t.Errorf("expected yml but got %s", yamlExtYML.String())
	}
	if yamlExtYAML.String() != "yaml" {
		t.Errorf("expected yaml but got %s", yamlExtYAML.String())
	}
}

func TestDataSource(t *testing.T) {
	if len(popularActions) == 0 {
		t.Fatal("popularActions is empty")
	}

	slugs := map[string]int{}
	for i, a := range popularActions {
		if j, ok := slugs[a.slug]; ok && popularActions[i].path == popularActions[j].path {
			t.Errorf("slug %q at popularActions[%d] was already added at popularActions[%d]", a.slug, i, j)
		} else {
			slugs[a.slug] = i
		}

		if len(a.tags) == 0 {
			t.Errorf("no tag is specified for %q", a.slug)
		}

		tags := map[string]int{}
		for i, tag := range a.tags {
			if tag == "" {
				t.Errorf("tags[%d] at action %q must not be empty string", i, a.slug)
				continue
			}
			if tag == a.next {
				t.Errorf("tags[%d] at action %q is equal to next verion %q", i, a.slug, a.next)
			}
			if j, ok := tags[tag]; ok {
				t.Errorf("duplicate tag %q at action %q appears: tags[%d] v.s. tags[%d]", tag, a.slug, i, j)
			} else {
				tags[tag] = i
			}
		}

		if a.ext != yamlExtYML && a.ext != yamlExtYAML {
			t.Errorf("ext of action %q is neither yamlExtYML nor yamlExtYAML: %d", a.slug, a.ext)
		}
	}
}

func TestReadWriteJSONL(t *testing.T) {
	testCases := []struct {
		file        string
		skipInputs  slugSet
		skipOutputs slugSet
	}{
		{
			file: "test.jsonl",
		},
		{
			file:       "skip_inputs.jsonl",
			skipInputs: slugSet{"rhysd/action-setup-vim": {}},
		},
		{
			file:        "skip_outputs.jsonl",
			skipOutputs: slugSet{"rhysd/action-setup-vim": {}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.file, func(t *testing.T) {
			f := filepath.Join("testdata", tc.file)
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			status := newApp(stdout, stderr, io.Discard, testDummyPopularActions, tc.skipInputs, tc.skipOutputs).run([]string{"test", "-s", f, "-f", "jsonl"})
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
		in          string
		want        string
		skipInputs  slugSet
		skipOutputs slugSet
	}{
		{
			in:   "test.jsonl",
			want: "want.go",
		},
		{
			in:         "skip_inputs.jsonl",
			want:       "skip_inputs_want.go",
			skipInputs: slugSet{"rhysd/action-setup-vim": {}},
		},
		{
			in:          "skip_outputs.jsonl",
			want:        "skip_outputs_want.go",
			skipOutputs: slugSet{"rhysd/action-setup-vim": {}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.in, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			a := newApp(stdout, stderr, io.Discard, testDummyPopularActions, tc.skipInputs, tc.skipOutputs)
			status := a.run([]string{"test", "-s", filepath.Join("testdata", tc.in)})
			if status != 0 {
				t.Fatalf("exit status is non-zero: %d: %s", status, stderr.Bytes())
			}

			b, err := os.ReadFile(filepath.Join("testdata", tc.want))
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
	in := filepath.Join("testdata", "test.jsonl")
	b, err := os.ReadFile(in)
	if err != nil {
		panic(err)
	}

	out := filepath.Join("testdata", "out.jsonl")
	defer os.Remove(out)

	stdout := io.Discard
	stderr := io.Discard
	status := newApp(stdout, stderr, io.Discard, testDummyPopularActions, nil, nil).run([]string{"test", "-s", in, "-f", "jsonl", out})
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
	in := filepath.Join("testdata", "test.jsonl")
	out := filepath.Join("testdata", "out.go")
	defer os.Remove(out)

	stdout := io.Discard
	stderr := io.Discard
	status := newApp(stdout, stderr, io.Discard, testDummyPopularActions, nil, nil).run([]string{"test", "-s", in, out})
	if status != 0 {
		t.Fatal("exit status is non-zero:", status)
	}

	b, err := os.ReadFile(filepath.Join("testdata", "want.go"))
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
	data := []*action{
		{
			slug: "rhysd/action-setup-vim",
			tags: []string{"v1.2.7"},
		},
		{
			slug: "rhysd/changelog-from-release",
			path: "/action",
			tags: []string{"v2.2.2"},
		},
	}
	stdout := &bytes.Buffer{}
	stderr := io.Discard
	status := newApp(stdout, stderr, io.Discard, data, nil, nil).run([]string{"test"})
	if status != 0 {
		t.Fatal("exit status is non-zero:", status)
	}

	b, err := os.ReadFile(filepath.Join("testdata", "fetched.go"))
	if err != nil {
		panic(err)
	}
	want := string(b)
	have := stdout.String()

	if !cmp.Equal(want, have) {
		t.Fatalf("fetched JSONL data does not match: %s", cmp.Diff(want, have))
	}
}

func TestLogOutput(t *testing.T) {
	f := filepath.Join("testdata", "test.jsonl")
	stdout := &bytes.Buffer{}
	logged := &bytes.Buffer{}
	status := newApp(stdout, io.Discard, logged, testDummyPopularActions, nil, nil).run([]string{"test", "-s", f, "-f", "jsonl"})
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
	status = newApp(stdout, io.Discard, logged, testDummyPopularActions, nil, nil).run([]string{"test", "-s", f, "-f", "jsonl", "-q"})
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
	status := newApp(stdout, stderr, io.Discard, testDummyPopularActions, nil, nil).run([]string{"test", "-help"})
	if status != 0 {
		t.Fatal("exit status is non-zero:", status)
	}
	out := stderr.String()
	if !strings.Contains(out, "Usage:") {
		t.Fatalf("usage header is not included in -help output: %q", out)
	}
}

func TestDetectNewRelease(t *testing.T) {
	data := []*action{
		{
			slug: "rhysd/action-setup-vim",
			tags: []string{"v0"},
			next: "v1",
		},
	}
	stdout := &bytes.Buffer{}
	stderr := io.Discard
	status := newApp(stdout, stderr, io.Discard, data, nil, nil).run([]string{"test", "-d"})
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
	testCases := []struct {
		what string
		next string
	}{
		{"no new version yet", "this-is-awesome-new-version"},
		{"no next version", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.what, func(t *testing.T) {
			data := []*action{
				{
					slug: "rhysd/action-setup-vim",
					tags: []string{"v1"},
					next: tc.next,
				},
			}
			stdout := &bytes.Buffer{}
			stderr := io.Discard
			status := newApp(stdout, stderr, io.Discard, data, nil, nil).run([]string{"test", "-d"})
			if status != 0 {
				t.Fatal("exit status is non-zero:", status)
			}
			out := stdout.String()
			if out != "" {
				t.Fatalf("stdout is not empty: %q", out)
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
			f := filepath.Join("testdata", tc.file)
			stdout := io.Discard
			stderr := &bytes.Buffer{}

			status := newApp(stdout, stderr, io.Discard, testDummyPopularActions, nil, nil).run([]string{"test", "-s", f})
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
	f := filepath.Join("testdata", "test.jsonl")
	out := filepath.Join("testdata", "this-dir-does-not-exit", "foo.jsonl")
	stdout := io.Discard
	stderr := &bytes.Buffer{}

	status := newApp(stdout, stderr, io.Discard, testDummyPopularActions, nil, nil).run([]string{"test", "-s", f, "-f", "jsonl", out})
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
			f := filepath.Join("testdata", "test.jsonl")
			stdout := testErrorWriter{}
			stderr := &bytes.Buffer{}

			status := newApp(stdout, stderr, io.Discard, testDummyPopularActions, nil, nil).run([]string{"test", "-s", f, "-f", format})
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
	data := []*action{
		{
			slug: "rhysd/this-action-does-not-exist",
			tags: []string{"v1"},
			next: "v2",
		},
	}

	stdout := testErrorWriter{}
	stderr := &bytes.Buffer{}

	status := newApp(stdout, stderr, io.Discard, data, nil, nil).run([]string{"test"})
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
			stdout := testErrorWriter{}
			stderr := &bytes.Buffer{}

			status := newApp(stdout, stderr, io.Discard, testDummyPopularActions, nil, nil).run(tc.args)
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
	data := []*action{
		// This expects to cause 400 Bad Request
		{
			slug: "",
			tags: []string{"v1"},
			next: "v2",
		},
	}
	stdout := io.Discard
	stderr := &bytes.Buffer{}
	status := newApp(stdout, stderr, io.Discard, data, nil, nil).run([]string{"test", "-d"})
	if status != 1 {
		t.Fatal("exit status is not 1:", status)
	}
	out := stderr.String()
	if !strings.Contains(out, "head request for https://raw.githubusercontent.com//v2/action.yml was not successful") {
		t.Fatalf("stderr was unexpected: %q", out)
	}
}

func TestActionBuildRawURL(t *testing.T) {
	a := &action{slug: "foo/bar"}
	have := a.rawURL("v1")
	want := "https://raw.githubusercontent.com/foo/bar/v1/action.yml"
	if have != want {
		t.Errorf("Wanted %q but have %q", want, have)
	}

	a = &action{slug: "foo/bar", path: "/a/b"}
	have = a.rawURL("v1")
	want = "https://raw.githubusercontent.com/foo/bar/v1/a/b/action.yml"
	if have != want {
		t.Errorf("Wanted %q but have %q", want, have)
	}

	a = &action{slug: "foo/bar", ext: yamlExtYAML}
	have = a.rawURL("v1")
	want = "https://raw.githubusercontent.com/foo/bar/v1/action.yaml"
	if have != want {
		t.Errorf("Wanted %q but have %q", want, have)
	}
}

func TestActionBuildGitHubURL(t *testing.T) {
	a := &action{slug: "foo/bar"}
	have := a.githubURL("v1")
	want := "https://github.com/foo/bar/tree/v1"
	if have != want {
		t.Errorf("Wanted %q but have %q", want, have)
	}

	a = &action{slug: "foo/bar", path: "/a/b"}
	have = a.githubURL("v1")
	want = "https://github.com/foo/bar/tree/v1/a/b"
	if have != want {
		t.Errorf("Wanted %q but have %q", want, have)
	}
}

func TestActionBuildSpec(t *testing.T) {
	a := &action{slug: "foo/bar"}
	have := a.spec("v1")
	want := "foo/bar@v1"
	if have != want {
		t.Errorf("Wanted %q but have %q", want, have)
	}

	a = &action{slug: "foo/bar", path: "/a/b"}
	have = a.spec("v1")
	want = "foo/bar/a/b@v1"
	if have != want {
		t.Errorf("Wanted %q but have %q", want, have)
	}
}
