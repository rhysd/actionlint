package actionlint

import (
	"bytes"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v3"
)

func testGetWantedActionMetadata() *ActionMetadata {
	want := &ActionMetadata{
		Name: "My action",
		Inputs: ActionMetadataInputs{
			"name":     {"name", false},
			"message":  {"message", true},
			"addition": {"addition", false},
		},
		Outputs: ActionMetadataOutputs{
			"user_id": {"user_id"},
		},
		Runs: ActionMetadataRuns{
			Using: "node20",
		},
	}
	return want
}

// Normal cases

func testCachedFlag(t *testing.T, want, have bool) {
	if want != have {
		msg := "metadata should be cached but actually it is not cached"
		if !want {
			msg = "metadata should not be cached but actually it is cached"
		}
		t.Error(msg)
	}
}

func TestLocalActionsFindMetadata(t *testing.T) {
	proj := &Project{filepath.Join("testdata", "action_metadata"), nil}
	c := NewLocalActionsCache(proj, nil)

	want := testGetWantedActionMetadata()
	for _, spec := range []string{"./action-yml", "./action-yaml"} {
		t.Run(spec, func(t *testing.T) {
			// read metadata repeatedly (should be cached)
			for i := 0; i < 3; i++ {
				have, cached, err := c.FindMetadata(spec)
				if err != nil {
					t.Fatal(i, err)
				}
				if have == nil {
					t.Fatal(i, "metadata is nil")
				}
				testCachedFlag(t, cached, i > 0)
				if !cmp.Equal(want, have) {
					t.Fatal(i, cmp.Diff(want, have))
				}
			}
		})
	}

	t.Run("./empty", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			m, cached, err := c.FindMetadata("./empty")
			if err != nil {
				t.Fatal(i, err)
			}
			if m == nil {
				t.Fatal(i, "metadata is nil")
			}
			if len(m.Inputs) != 0 {
				t.Fatal("inputs are not empty", m.Inputs)
			}
			if len(m.Outputs) != 0 {
				t.Fatal("outputs are not empty", m.Outputs)
			}
			testCachedFlag(t, cached, i > 0)
		}
	})

	t.Run("./uppercase", func(t *testing.T) {
		want := testGetWantedActionMetadata()
		for _, i := range want.Inputs {
			i.Name = strings.ToUpper(i.Name)
		}
		for _, o := range want.Outputs {
			o.Name = strings.ToUpper(o.Name)
		}

		have, cached, err := c.FindMetadata("./uppercase")
		if err != nil {
			t.Fatal(err)
		}
		testCachedFlag(t, cached, false)
		if !cmp.Equal(want, have) {
			t.Fatal(cmp.Diff(want, have))
		}
	})

	for _, using := range []string{"docker", "composite"} {
		a, u := "./"+using, using
		t.Run(a, func(t *testing.T) {
			want := &ActionMetadata{
				Name: "My action",
				Runs: ActionMetadataRuns{
					Using: u,
				},
			}
			have, cached, err := c.FindMetadata(a)
			if err != nil {
				t.Fatal(err)
			}
			testCachedFlag(t, cached, false)
			if !cmp.Equal(want, have) {
				t.Fatal(cmp.Diff(want, have))
			}
		})
	}
}

func TestLocalActionsFindConcurrently(t *testing.T) {
	n := 10
	proj := &Project{filepath.Join("testdata", "action_metadata"), nil}
	c := NewLocalActionsCache(proj, nil)
	ret := make(chan *ActionMetadata)
	err := make(chan error)

	for i := 0; i < n; i++ {
		go func() {
			m, _, e := c.FindMetadata("./action-yml")
			if e != nil {
				err <- e
				return
			}
			ret <- m
		}()
	}

	ms := []*ActionMetadata{}
	errs := []error{}
	for i := 0; i < n; i++ {
		select {
		case m := <-ret:
			ms = append(ms, m)
		case e := <-err:
			errs = append(errs, e)
		}
	}

	if len(errs) != 0 {
		t.Fatal("some error occurred:", errs)
	}

	want := testGetWantedActionMetadata()
	for _, have := range ms {
		if !cmp.Equal(want, have) {
			t.Fatal(cmp.Diff(want, have))
		}
	}

	_, cached, _ := c.FindMetadata("./action-yml")
	testCachedFlag(t, true, cached)
}

func TestLocalActionsParsingSkipped(t *testing.T) {
	tests := []struct {
		what string
		proj *Project
		spec string
	}{
		{
			what: "project is nil",
			proj: nil,
			spec: "./action",
		},
		{
			what: "not a local action",
			proj: &Project{"", nil},
			spec: "actions/checkout@v3",
		},
		{
			what: "action does not exist (#25, #40)",
			proj: &Project{filepath.Join("testdata", "action_metadata"), nil},
			spec: "./this-action-does-not-exist",
		},
	}

	for _, tc := range tests {
		t.Run(tc.what, func(t *testing.T) {
			c := NewLocalActionsCache(tc.proj, nil)
			m, cached, err := c.FindMetadata(tc.spec)
			if err != nil {
				t.Fatal(tc.spec, "error occurred:", err)
			}
			if m != nil {
				t.Fatal(tc.spec, "metadata was parsed", m)
			}
			testCachedFlag(t, false, cached)
		})
	}
}

func TestLocalActionsIgnoreRemoteActions(t *testing.T) {
	proj := &Project{filepath.Join("testdata", "action_metadata"), nil}
	c := NewLocalActionsCache(proj, nil)
	for _, spec := range []string{"actions/checkout@v2", "docker://example.com/foo/bar"} {
		m, cached, err := c.FindMetadata(spec)
		if err != nil {
			t.Fatal(spec, "error occurred:", err)
		}
		if m != nil {
			t.Fatal(spec, "metadata was parsed", m)
		}
		testCachedFlag(t, false, cached)
	}
}

func TestLocalActionsLogCacheHit(t *testing.T) {
	dbg := &bytes.Buffer{}
	proj := &Project{filepath.Join("testdata", "action_metadata"), nil}
	c := NewLocalActionsCache(proj, dbg)

	want := testGetWantedActionMetadata()
	for i := 0; i < 2; i++ {
		have, _, err := c.FindMetadata("./action-yml")
		if err != nil {
			t.Fatal(err)
		}
		if !cmp.Equal(want, have) {
			t.Fatal(cmp.Diff(want, have))
		}
	}

	logs := strings.Split(strings.TrimSpace(dbg.String()), "\n")
	if len(logs) != 2 {
		t.Fatalf("2 logs were expected but got %d logs: %#v", len(logs), logs)
	}
	if !strings.Contains(logs[0], "New metadata parsed from action "+filepath.Join("testdata", "action_metadata", "action-yml")) {
		t.Fatalf("first log should be 'new metadata' but got %q", logs[0])
	}
	if !strings.Contains(logs[1], "Cache hit for ./action-yml") {
		t.Fatalf("second log should be 'cache hit' but got %q", logs[1])
	}
}

func TestLocalActionsNullCache(t *testing.T) {
	c := newNullLocalActionsCache(io.Discard)
	m, cached, err := c.FindMetadata("./path/to/action.yaml")
	if m != nil {
		t.Error("metadata should not be found:", m)
	}
	if err != nil {
		t.Error(err)
	}
	testCachedFlag(t, false, cached)
}

// Error cases

func TestLocalActionsBrokenMetadata(t *testing.T) {
	proj := &Project{filepath.Join("testdata", "action_metadata"), nil}
	c := NewLocalActionsCache(proj, nil)
	m, cached, err := c.FindMetadata("./broken")
	if err == nil {
		t.Fatal("error was not returned", m)
	}
	if !strings.Contains(err.Error(), "could not parse action metadata") {
		t.Fatal("unexpected error:", err)
	}
	testCachedFlag(t, false, cached)

	// Second try does not return error, but metadata is also nil not to show the same error from
	// multiple rules.
	m, cached, err = c.FindMetadata("./broken")
	if err != nil {
		t.Fatal("error was returned at second try", err)
	}
	if m != nil {
		t.Fatal("metadata was not nil even if it does not exist", m)
	}
	testCachedFlag(t, true, cached)
}

func TestLocalActionsDuplicateInputsOutputs(t *testing.T) {
	proj := &Project{filepath.Join("testdata", "action_metadata"), nil}
	c := NewLocalActionsCache(proj, nil)

	for _, tc := range []struct {
		spec string
		want string
	}{
		{
			spec: "./input-duplicate",
			want: "input \"FOO\" is duplicated",
		},
		{
			spec: "./output-duplicate",
			want: "output \"FOO\" is duplicated",
		},
	} {
		t.Run(tc.spec, func(t *testing.T) {
			m, cached, err := c.FindMetadata(tc.spec)
			if err == nil {
				t.Fatal("error was not returned", m)
			}
			msg := err.Error()
			if !strings.Contains(msg, tc.want) {
				t.Fatalf("error %q was expected to include %q", msg, tc.want)
			}
			testCachedFlag(t, false, cached)
		})
	}
}

func TestLocalActionsConcurrentFailures(t *testing.T) {
	n := 10
	proj := &Project{filepath.Join("testdata", "action_metadata"), nil}
	c := NewLocalActionsCache(proj, nil)
	errC := make(chan error)

	for i := 0; i < n; i++ {
		go func() {
			_, _, err := c.FindMetadata("./broken")
			errC <- err
		}()
	}

	errs := []error{}
	for i := 0; i < n; i++ {
		errs = append(errs, <-errC)
	}

	// At least once error was reported
	var err error
	for _, e := range errs {
		if e != nil {
			err = e
			break
		}
	}

	if err == nil {
		t.Fatal("error did not occur", err)
	}
	if !strings.Contains(err.Error(), "could not parse action metadata") {
		t.Fatal("unexpected error:", err)
	}
}

func TestLocalActionsConcurrentMultipleMetadataAndFailures(t *testing.T) {
	proj := &Project{filepath.Join("testdata", "action_metadata"), nil}
	c := NewLocalActionsCache(proj, nil)

	inputs := []string{
		"./action-yml",
		"./action-yaml",
		"./action-yml",
		"./broken",
		"./action-yaml",
		"./action-yaml",
		"./broken",
		"./action-yml",
		"./broken",
		"./action-yaml",
	}

	reqC := make(chan string)
	retC := make(chan *ActionMetadata)
	errC := make(chan error)
	done := make(chan struct{})

	for i := 0; i < 3; i++ {
		go func() {
			for {
				select {
				case spec := <-reqC:
					m, _, err := c.FindMetadata(spec)
					if m == nil {
						errC <- err
						break
					}
					retC <- m
				case <-done:
					return
				}
			}
		}()
	}

	go func() {
		for _, in := range inputs {
			select {
			case reqC <- in:
			case <-done:
				return
			}
		}
	}()

	ret := []*ActionMetadata{}
	errs := []error{}
	for i := 0; i < len(inputs); i++ {
		select {
		case m := <-retC:
			ret = append(ret, m)
		case err := <-errC:
			errs = append(errs, err)
		}
	}
	close(done)

	numErrs := 0
	for _, in := range inputs {
		if in == "./broken" {
			numErrs++
		}
	}
	numRet := len(inputs) - numErrs

	if len(errs) != numErrs {
		t.Fatalf("wanted %d errors but got %d: %v", numErrs, len(errs), errs)
	}
	if len(ret) != numRet {
		t.Fatalf("wanted %d errors but got %d: %v", numRet, len(ret), ret)
	}

	var err error
	for _, e := range errs {
		if e != nil {
			err = e
			break
		}
	}

	if err == nil {
		t.Fatal("error did not occur", err)
	}
	if !strings.Contains(err.Error(), "could not parse action metadata") {
		t.Fatal("unexpected error:", err)
	}

	want := testGetWantedActionMetadata()
	for _, have := range ret {
		if !cmp.Equal(want, have) {
			t.Fatal("unexpected metadata:", cmp.Diff(want, have))
		}
	}
}

func TestActionMetadataYAMLUnmarshalOK(t *testing.T) {
	testCases := []struct {
		what  string
		input string
		want  ActionMetadata
	}{
		{
			what:  "no input and no output",
			input: `name: Test`,
			want: ActionMetadata{
				Name: "Test",
			},
		},
		{
			what: "inputs",
			input: `name: Test
inputs:
  input1:
    description: test
  input2:
    description: test
    required: false
  input3:
    description: test
    required: true
    default: 'default'
  input4:
    description: test
    required: false
    default: 'default'
  input5:
    description: test
    required: true`,
			want: ActionMetadata{
				Name: "Test",
				Inputs: ActionMetadataInputs{
					"input1": {"input1", false},
					"input2": {"input2", false},
					"input3": {"input3", false},
					"input4": {"input4", false},
					"input5": {"input5", true},
				},
			},
		},
		{
			what: "outputs",
			input: `name: Test
outputs:
  output1:
    description: test
  output2:
    description: test
`,
			want: ActionMetadata{
				Name: "Test",
				Outputs: ActionMetadataOutputs{
					"output1": {"output1"},
					"output2": {"output2"},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.what, func(t *testing.T) {
			var have ActionMetadata
			if err := yaml.Unmarshal([]byte(tc.input), &have); err != nil {
				t.Fatal(err)
			}
			if !cmp.Equal(&tc.want, &have) {
				t.Fatal(cmp.Diff(&tc.want, &have))
			}
		})
	}
}

func TestActionMetadataYAMLUnmarshalError(t *testing.T) {
	testCases := []struct {
		what  string
		input string
		want  string
	}{
		{
			what: "invalid inputs",
			input: `name: Test
inputs: "foo"`,
			want: "inputs must be mapping node",
		},
		{
			what: "invalid inputs.*",
			input: `name: Test
inputs:
  input1: "foo"`,
			want: "into actionlint.actionInputMetadata",
		},
		{
			what: "invalid outputs",
			input: `name: Test
outputs: "foo"`,
			want: "outputs must be mapping node",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.what, func(t *testing.T) {
			var data ActionMetadata
			err := yaml.Unmarshal([]byte(tc.input), &data)
			if err == nil {
				t.Fatal("error did not occur")
			}
			msg := err.Error()
			if !strings.Contains(msg, tc.want) {
				t.Fatalf("%q is not contained in error message %q", tc.want, msg)
			}
		})
	}
}

func TestLocalActionsCacheFactory(t *testing.T) {
	f := NewLocalActionsCacheFactory(io.Discard)
	p1 := &Project{"path/to/project1", nil}
	c1 := f.GetCache(p1)

	p2 := &Project{"path/to/project2", nil}
	c2 := f.GetCache(p2)
	if c1 == c2 {
		t.Errorf("different cache was not created: %v", c1)
	}

	c3 := f.GetCache(p1)
	if c1 != c3 {
		t.Errorf("same cache is not returned for the same project: %v vs %v", c1, c3)
	}

	c4 := f.GetCache(nil)
	if c4.proj != nil {
		t.Errorf("null cache must be returned if given project is nil: %v", c4)
	}
}
