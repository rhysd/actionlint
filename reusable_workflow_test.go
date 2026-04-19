package actionlint

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestReusableWorkflowUnmarshalOK(t *testing.T) {
	tests := []struct {
		what string
		src  string
		want *ReusableWorkflowMetadata
	}{
		{
			what: "minimal",
			src: `
			on:
			  workflow_call:
			    inputs:
			      i:
			        type: string
			    outputs:
			      o:
			        value: foo
			    secrets:
			      x:
			`,
			want: &ReusableWorkflowMetadata{
				Inputs: ReusableWorkflowMetadataInputs{
					"i": {"i", false, StringType{}},
				},
				Outputs: ReusableWorkflowMetadataOutputs{
					"o": {"o"},
				},
				Secrets: ReusableWorkflowMetadataSecrets{
					"x": {"x", false},
				},
			},
		},
		{
			what: "empty",
			src: `
			on:
			  workflow_call:
			`,
			want: &ReusableWorkflowMetadata{
				Inputs:  nil,
				Outputs: nil,
				Secrets: nil,
			},
		},
		{
			what: "empty values",
			src: `
			on:
			  workflow_call:
			    inputs:
			    outputs:
			    secrets:
			`,
			want: &ReusableWorkflowMetadata{
				Inputs:  nil,
				Outputs: nil,
				Secrets: nil,
			},
		},
		{
			what: "inputs",
			src: `
			on:
			  workflow_call:
			    inputs:
			      a:
			        type: string
			      b:
			        type: number
			      c:
			        type: boolean
			      d:
			      e:
			        type: string
			        required: false
			      f:
			        type: string
			        required: true
			      g:
			        type: string
			        default: abc
			      h:
			        type: string
			        default: abc
			        required: true
			      i:
			        required: true
			`,
			want: &ReusableWorkflowMetadata{
				Inputs: ReusableWorkflowMetadataInputs{
					"a": {"a", false, StringType{}},
					"b": {"b", false, NumberType{}},
					"c": {"c", false, BoolType{}},
					"d": {"d", false, AnyType{}},
					"e": {"e", false, StringType{}},
					"f": {"f", true, StringType{}},
					"g": {"g", false, StringType{}},
					"h": {"h", false, StringType{}},
					"i": {"i", true, AnyType{}},
				},
				Outputs: nil,
				Secrets: nil,
			},
		},
		{
			what: "outputs",
			src: `
			on:
			  workflow_call:
			    inputs:
			      i:
			        type: string
			    outputs:
			      o:
			        value: foo
			    secrets:
			      x:
			`,
			want: &ReusableWorkflowMetadata{
				Inputs: ReusableWorkflowMetadataInputs{
					"i": {"i", false, StringType{}},
				},
				Outputs: ReusableWorkflowMetadataOutputs{
					"o": {"o"},
				},
				Secrets: ReusableWorkflowMetadataSecrets{
					"x": {"x", false},
				},
			},
		},
		{
			what: "secrets",
			src: `
			on:
			  workflow_call:
			    secrets:
			      x:
			      y:
			        required: false
			      z:
			        required: true
			`,
			want: &ReusableWorkflowMetadata{
				Inputs:  nil,
				Outputs: nil,
				Secrets: ReusableWorkflowMetadataSecrets{
					"x": {"x", false},
					"y": {"y", false},
					"z": {"z", true},
				},
			},
		},
		{
			what: "empty event in scalar node",
			src: `
			on: workflow_call
			`,
			want: &ReusableWorkflowMetadata{
				Inputs:  nil,
				Outputs: nil,
				Secrets: nil,
			},
		},
		{
			what: "empty event in sequence node",
			src: `
			on: [workflow_call]
			`,
			want: &ReusableWorkflowMetadata{
				Inputs:  nil,
				Outputs: nil,
				Secrets: nil,
			},
		},
		{
			what: "empty event in sequence node with other events",
			src: `
			on: [pull_request, workflow_call, push]
			`,
			want: &ReusableWorkflowMetadata{
				Inputs:  nil,
				Outputs: nil,
				Secrets: nil,
			},
		},
		{
			what: "upper case",
			src: `
			on:
			  workflow_call:
			    inputs:
			      MY_INPUT1:
			        type: string
			      MY_INPUT2:
			        type: number
			    outputs:
			      MY_OUTPUT1:
			        value: foo
			      MY_OUTPUT2:
			        value: foo
			    secrets:
			      MY_SECRET1:
			      MY_SECRET2:
			        required: true
			`,
			want: &ReusableWorkflowMetadata{
				Inputs: ReusableWorkflowMetadataInputs{
					"my_input1": {"MY_INPUT1", false, StringType{}},
					"my_input2": {"MY_INPUT2", false, NumberType{}},
				},
				Outputs: ReusableWorkflowMetadataOutputs{
					"my_output1": {"MY_OUTPUT1"},
					"my_output2": {"MY_OUTPUT2"},
				},
				Secrets: ReusableWorkflowMetadataSecrets{
					"my_secret1": {"MY_SECRET1", false},
					"my_secret2": {"MY_SECRET2", true},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.what, func(t *testing.T) {
			src := strings.TrimSpace(tc.src)
			src = strings.ReplaceAll(src, "\t", "")

			m, err := parseReusableWorkflowMetadata([]byte(src))
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(m, tc.want); diff != "" {
				t.Fatal("Parse result is unexpected. diff:\n" + diff)
			}
		})
	}
}

func TestReusableWorkflowUnmarshalOnNodeNotFound(t *testing.T) {
	src := "hello: world"
	_, err := parseReusableWorkflowMetadata([]byte(src))
	if err == nil {
		t.Fatal("Error did not happen")
	}
	if msg := err.Error(); msg != "\"on:\" is not found" {
		t.Fatal("Unexpected error:", msg)
	}
}

func TestReusableWorkflowUnmarshalEventNotFound(t *testing.T) {
	tests := []struct {
		what string
		src  string
		line int
		col  int
	}{
		{
			what: "scalar",
			src: `
			on: push
			`,
			line: 1,
			col:  5,
		},
		{
			what: "sequence",
			src: `
			on: [push, pull_request]
			`,
			line: 1,
			col:  5,
		},
		{
			what: "empty sequence",
			src: `
			on: []
			`,
			line: 1,
			col:  5,
		},
		{
			what: "mapping",
			src: `
			on:
			  push:
			  pull_request:
			`,
			line: 2,
			col:  3,
		},
		{
			what: "empty mapping",
			src: `
			on: {}
			`,
			line: 1,
			col:  5,
		},
		{
			what: "null",
			src: `
			on:
			`,
			line: 1,
			col:  4,
		},
	}

	for _, tc := range tests {
		t.Run(tc.what, func(t *testing.T) {
			src := strings.TrimSpace(tc.src)
			src = strings.ReplaceAll(src, "\t", "")

			_, err := parseReusableWorkflowMetadata([]byte(src))
			if err == nil {
				t.Fatal("Error did not happen")
			}

			msg := err.Error()
			if !strings.Contains(msg, "\"workflow_call\" event trigger is not found in \"on:\"") {
				t.Fatal("Unexpected error:", msg)
			}
			loc := fmt.Sprintf("line:%d, column:%d", tc.line, tc.col)
			if !strings.Contains(msg, loc) {
				t.Fatalf("location is not %q: %s", loc, msg)
			}
		})
	}
}

var testReusableWorkflowWantedMetadata *ReusableWorkflowMetadata = &ReusableWorkflowMetadata{
	Inputs: ReusableWorkflowMetadataInputs{
		"input1": {"input1", false, StringType{}},
		"input2": {"input2", true, BoolType{}},
	},
	Outputs: ReusableWorkflowMetadataOutputs{
		"output1": {"output1"},
	},
	Secrets: ReusableWorkflowMetadataSecrets{
		"secret1": {"secret1", false},
		"secret2": {"secret2", true},
	},
}

func TestReusableWorkflowCacheFindMetadataOK(t *testing.T) {
	proj := &Project{filepath.Join("testdata", "reusable_workflow_metadata"), nil}
	c := NewLocalReusableWorkflowCache(proj, "", nil)

	m, err := c.FindMetadata("./ok.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(m, testReusableWorkflowWantedMetadata); diff != "" {
		t.Fatal(diff)
	}

	m2, err := c.FindMetadata("./ok.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if m != m2 {
		t.Error("metadata is not cached")
	}
}

func TestReusableWorkflowCacheFindMetadataError(t *testing.T) {
	tests := []struct {
		what string
		spec string
		want string
	}{
		{
			what: "broken workflow",
			spec: "./broken.yaml",
			want: "error while parsing reusable workflow \"./broken.yaml\"",
		},
		{
			what: "no hook",
			spec: "./no_hook.yaml",
			want: "\"workflow_call\" event trigger is not found in \"on:\"",
		},
		{
			what: "no on",
			spec: "./no_on.yaml",
			want: "\"on:\" is not found",
		},
		{
			what: "not existing workflow",
			spec: "./this-workflow-does-not-exist.yaml",
			want: "could not read reusable workflow file for \"./this-workflow-does-not-exist.yaml\":",
		},
		{
			what: "broken inputs",
			spec: "./broken_inputs.yaml",
			want: "error while parsing reusable workflow \"./broken_inputs.yaml\"",
		},
		{
			what: "broken secrets",
			spec: "./broken_secrets.yaml",
			want: "error while parsing reusable workflow \"./broken_secrets.yaml\"",
		},
	}

	for _, tc := range tests {
		t.Run(tc.what, func(t *testing.T) {
			proj := &Project{filepath.Join("testdata", "reusable_workflow_metadata"), nil}
			c := NewLocalReusableWorkflowCache(proj, "", nil)
			_, err := c.FindMetadata(tc.spec)
			if err == nil {
				t.Fatal("no error happened")
			}
			msg := err.Error()
			if !strings.Contains(msg, tc.want) {
				t.Fatalf("unexpected error. wanted %q but got %q", tc.want, msg)
			}
			// Trying to find metadata with the same spec later returns nil to avoid duplicate errors
			m, err := c.FindMetadata(tc.spec)
			if err != nil {
				t.Fatal("error happens when finding metadata again:", err)
			}
			if m != nil {
				t.Fatal("nil is not cached:", m)
			}
		})
	}
}

func TestReusableWorkflowCacheFindMetadataSkipParsing(t *testing.T) {
	p := &Project{filepath.Join("testdata", "reusable_workflow_metadata"), nil}
	tests := []struct {
		what string
		proj *Project
		spec string
	}{
		{
			what: "no project",
			proj: nil,
			spec: "./ok.yaml",
		},
		{
			what: "external workflow",
			proj: p,
			spec: "repo/owner/workflow@main",
		},
		{
			what: "template placeholder",
			proj: p,
			spec: "./${{ some_expression }}.yaml",
		},
	}

	for _, tc := range tests {
		t.Run(tc.what, func(t *testing.T) {
			c := NewLocalReusableWorkflowCache(tc.proj, "", nil)
			m, err := c.FindMetadata(tc.spec)
			if err != nil {
				t.Fatal(err)
			}
			if m != nil {
				t.Fatal("metadata should be nil:", m)
			}
			m, err = c.FindMetadata(tc.spec)
			if err != nil {
				t.Fatal(err)
			}
			if m != nil {
				t.Fatal("nil is not cached:", m)
			}
		})
	}
}

func TestReusableWorkflowConvertWorkflowPathToSpec(t *testing.T) {
	p := &Project{filepath.Join("path", "to", "project"), nil}
	cwd := filepath.Join("path", "to", "project", "cwd")
	tests := []struct {
		what string
		proj *Project
		path string
		want string
		ok   bool
	}{
		{
			what: "current dir",
			proj: p,
			path: filepath.Join("workflow.yaml"),
			want: "./cwd/workflow.yaml",
			ok:   true,
		},
		{
			what: "child dir",
			proj: p,
			path: filepath.Join("dir", "workflow.yaml"),
			want: "./cwd/dir/workflow.yaml",
			ok:   true,
		},
		{
			what: "parent dir",
			proj: p,
			path: filepath.Join("..", "dir", "workflow.yaml"),
			want: "./dir/workflow.yaml",
			ok:   true,
		},
		{
			what: "no project",
			proj: nil,
			ok:   false,
		},
		{
			what: "other project",
			proj: &Project{filepath.Join("path", "to", "other-project"), nil},
			ok:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.what, func(t *testing.T) {
			c := NewLocalReusableWorkflowCache(tc.proj, cwd, nil)
			s, ok := c.convWorkflowPathToSpec(tc.path)
			if ok != tc.ok {
				t.Fatalf("should return %v but got %v (spec=%q)", tc.ok, ok, s)
			}
			if ok && s != tc.want {
				t.Fatalf("wanted spec %q but got %q", tc.want, s)
			}
		})
	}
}

func TestReusableWorkflowMetadataFromASTNodeInputs(t *testing.T) {
	tests := []struct {
		what   string
		inputs map[string]*WorkflowCallEventInput
		want   ReusableWorkflowMetadataInputs
	}{
		{
			what: "type of inputs",
			inputs: map[string]*WorkflowCallEventInput{
				"string_input":  {Type: WorkflowCallEventInputTypeString},
				"number_input":  {Type: WorkflowCallEventInputTypeNumber},
				"bool_input":    {Type: WorkflowCallEventInputTypeBoolean},
				"unknown_input": {},
			},
			want: ReusableWorkflowMetadataInputs{
				"string_input":  {"string_input", false, StringType{}},
				"number_input":  {"number_input", false, NumberType{}},
				"bool_input":    {"bool_input", false, BoolType{}},
				"unknown_input": {"unknown_input", false, AnyType{}},
			},
		},
		{
			what: "required or optional",
			inputs: map[string]*WorkflowCallEventInput{
				"unspecified":  {},
				"not_required": {Required: &Bool{Value: false, Pos: &Pos{}}},
				"required":     {Required: &Bool{Value: true, Pos: &Pos{}}},
				"required_but_default": {
					Required: &Bool{Value: true, Pos: &Pos{}},
					Default:  &String{Pos: &Pos{}},
				},
				"expression": {
					Required: &Bool{
						Expression: &String{Pos: &Pos{}},
						Pos:        &Pos{},
					},
				},
			},
			want: ReusableWorkflowMetadataInputs{
				"unspecified":          {"unspecified", false, AnyType{}},
				"not_required":         {"not_required", false, AnyType{}},
				"required":             {"required", true, AnyType{}},
				"required_but_default": {"required_but_default", false, AnyType{}},
				"expression":           {"expression", false, AnyType{}},
			},
		},
		{
			what:   "empty",
			inputs: map[string]*WorkflowCallEventInput{},
			want:   ReusableWorkflowMetadataInputs{},
		},
		{
			what: "upper case input",
			inputs: map[string]*WorkflowCallEventInput{
				"MY_INPUT": {Type: WorkflowCallEventInputTypeString},
			},
			want: ReusableWorkflowMetadataInputs{
				"my_input": {"MY_INPUT", false, StringType{}},
			},
		},
		{
			what: "upper case inputs",
			inputs: map[string]*WorkflowCallEventInput{
				"MY_INPUT": {},
			},
			want: ReusableWorkflowMetadataInputs{
				"my_input": {"MY_INPUT", false, AnyType{}},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.what, func(t *testing.T) {
			cwd := filepath.Join("path", "to", "project")
			proj := &Project{cwd, nil}
			c := NewLocalReusableWorkflowCache(proj, cwd, nil)
			e := &WorkflowCallEvent{Inputs: []*WorkflowCallEventInput{}}
			for n, i := range tc.inputs {
				i.Name = &String{Value: n, Pos: &Pos{}}
				i.ID = strings.ToLower(n)
				e.Inputs = append(e.Inputs, i)
			}

			c.WriteWorkflowCallEvent(filepath.Join("foo", "test.yaml"), e)

			m, ok := c.readCache("./foo/test.yaml")
			if !ok {
				t.Fatal("Event was not converted to event")
			}

			if diff := cmp.Diff(m.Inputs, tc.want); diff != "" {
				t.Fatal(diff)
			}
			if len(m.Outputs) != 0 {
				t.Error("Outputs are not empty", m.Outputs)
			}
			if len(m.Secrets) != 0 {
				t.Error("Secrets are not empty", m.Secrets)
			}
		})
	}
}

func TestReusableWorkflowMetadataFromASTNodeOutputs(t *testing.T) {
	tests := [][]string{
		{},
		{"foo"},
		{"a", "b", "c"},
		{"A", "B", "C"},
	}
	for _, outputs := range tests {
		t.Run(fmt.Sprintf("%s", outputs), func(t *testing.T) {
			cwd := filepath.Join("path", "to", "project")
			proj := &Project{cwd, nil}
			c := NewLocalReusableWorkflowCache(proj, cwd, nil)
			e := &WorkflowCallEvent{Outputs: map[string]*WorkflowCallEventOutput{}}
			for _, o := range outputs {
				e.Outputs[strings.ToLower(o)] = &WorkflowCallEventOutput{
					Name: &String{Value: o, Pos: &Pos{}},
				}
			}

			c.WriteWorkflowCallEvent(filepath.Join("foo", "test.yaml"), e)

			m, ok := c.readCache("./foo/test.yaml")
			if !ok {
				t.Fatal("Event was not converted to event")
			}

			want := ReusableWorkflowMetadataOutputs{}
			for _, o := range outputs {
				want[strings.ToLower(o)] = &ReusableWorkflowMetadataOutput{o}
			}

			if diff := cmp.Diff(m.Outputs, want); diff != "" {
				t.Fatal(diff)
			}
			if len(m.Inputs) != 0 {
				t.Error("Inputs are not empty", m.Inputs)
			}
			if len(m.Secrets) != 0 {
				t.Error("Secrets are not empty", m.Secrets)
			}
		})
	}
}

func TestReusableWorkflowMetadataFromASTNodeSecrets(t *testing.T) {
	tests := []map[string]*Bool{
		{},
		{"a": nil},
		{"a": &Bool{Value: false, Pos: &Pos{}}},
		{"a": &Bool{Value: true, Pos: &Pos{}}},
		{"a": &Bool{Expression: &String{Pos: &Pos{}}, Pos: &Pos{}}},
		{
			"a": &Bool{Value: false, Pos: &Pos{}},
			"b": &Bool{Value: true, Pos: &Pos{}},
			"c": nil,
		},
		{
			"A": &Bool{Value: false, Pos: &Pos{}},
			"B": &Bool{Value: true, Pos: &Pos{}},
			"C": nil,
		},
	}
	for _, secrets := range tests {
		t.Run(fmt.Sprintf("%s", secrets), func(t *testing.T) {
			cwd := filepath.Join("path", "to", "project")
			proj := &Project{cwd, nil}
			c := NewLocalReusableWorkflowCache(proj, cwd, nil)
			e := &WorkflowCallEvent{Secrets: map[string]*WorkflowCallEventSecret{}}
			for n, r := range secrets {
				e.Secrets[strings.ToLower(n)] = &WorkflowCallEventSecret{
					Name:     &String{Value: n, Pos: &Pos{}},
					Required: r,
				}
			}

			c.WriteWorkflowCallEvent(filepath.Join("foo", "test.yaml"), e)

			m, ok := c.readCache("./foo/test.yaml")
			if !ok {
				t.Fatal("Event was not converted to event")
			}

			want := ReusableWorkflowMetadataSecrets{}
			for n, r := range secrets {
				want[strings.ToLower(n)] = &ReusableWorkflowMetadataSecret{
					Name:     n,
					Required: r != nil && r.Value,
				}
			}

			if diff := cmp.Diff(m.Secrets, want); diff != "" {
				t.Fatal(diff)
			}
			if len(m.Inputs) != 0 {
				t.Error("Inputs are not empty", m.Inputs)
			}
			if len(m.Outputs) != 0 {
				t.Error("Outputs are not empty", m.Outputs)
			}
		})
	}
}

func TestReusableWorkflowMetadataFromASTNodeDoNothing(t *testing.T) {
	cwd := filepath.Join("path", "to", "project")
	c := NewLocalReusableWorkflowCache(nil, cwd, nil)
	c.WriteWorkflowCallEvent("workflow.yaml", &WorkflowCallEvent{})
	m, ok := c.readCache("./workflow.yaml")
	if ok {
		t.Fatal("Metadata created:", m)
	}

	proj := &Project{cwd, nil}
	c = NewLocalReusableWorkflowCache(proj, filepath.Join("path", "to", "another-project"), nil)
	c.WriteWorkflowCallEvent("workflow.yaml", &WorkflowCallEvent{})
	m, ok = c.readCache("./workflow.yaml")
	if ok {
		t.Fatal("Metadata created:", m)
	}

	m1 := &ReusableWorkflowMetadata{}
	c = NewLocalReusableWorkflowCache(proj, cwd, nil)
	c.writeCache("./dir/workflow.yaml", m1)
	c.WriteWorkflowCallEvent(filepath.Join("dir", "workflow.yaml"), &WorkflowCallEvent{})
	m2, ok := c.readCache("./dir/workflow.yaml")
	if !ok {
		t.Fatal("Metadata was not created for ./dir/workflow.yaml")
	}
	if m1 != m2 {
		t.Fatalf("Metadata was not cached for ./dir/workflow.yaml %v vs %v", m1, m2)
	}
}

func TestReusableWorkflowMetadataCacheFindOneMetadataConcurrently(t *testing.T) {
	n := 10
	cwd := filepath.Join("testdata", "reusable_workflow_metadata")
	proj := &Project{cwd, nil}
	c := NewLocalReusableWorkflowCache(proj, cwd, nil)
	ret := make(chan *ReusableWorkflowMetadata)
	err := make(chan error)

	for i := 0; i < n; i++ {
		go func() {
			m, e := c.FindMetadata("./ok.yaml")
			if e != nil {
				err <- e
				return
			}
			ret <- m
		}()
	}

	ms := []*ReusableWorkflowMetadata{}
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
		t.Fatal("Error occurred:", errs)
	}

	for _, m := range ms {
		if diff := cmp.Diff(testReusableWorkflowWantedMetadata, m); diff != "" {
			t.Fatal(diff)
		}
	}

	if len(c.cache) != 1 {
		t.Errorf("Unexpected %d caches are stored: %v", len(c.cache), c.cache)
	}
	m, ok := c.readCache("./ok.yaml")
	if !ok {
		t.Fatal("Cache did not exist")
	}
	if m == nil {
		t.Fatal("nil is stored in cache")
	}
}

func TestReusableWorkflowMetadataCacheWriteFromFileAndASTNodeConcurrently(t *testing.T) {
	n := 10
	cwd := filepath.Join("testdata", "reusable_workflow_metadata")
	proj := &Project{cwd, nil}
	c := NewLocalReusableWorkflowCache(proj, cwd, nil)
	ret := make(chan struct{})
	err := make(chan error)

	fromFile := func() {
		if _, e := c.FindMetadata("./ok.yaml"); e != nil {
			err <- e
			return
		}
		ret <- struct{}{}
	}
	fromNode := func() {
		c.WriteWorkflowCallEvent("workflow.yaml", &WorkflowCallEvent{})
		if _, ok := c.readCache("./workflow.yaml"); !ok {
			err <- fmt.Errorf("Cache was not created from WorkflowCallEvent")
			return
		}
		ret <- struct{}{}
	}

	for i := 0; i < n/2; i++ {
		go fromFile()
		go fromNode()
	}

	errs := []error{}
	for i := 0; i < n; i++ {
		select {
		case <-ret:
		case e := <-err:
			errs = append(errs, e)
		}
	}

	if len(errs) != 0 {
		t.Fatal("Error occurred:", errs)
	}

	if len(c.cache) != 2 {
		t.Errorf("Size of cache should be 2 but got %d: %v", len(c.cache), c.cache)
	}
	if _, ok := c.cache["./ok.yaml"]; !ok {
		t.Error("Cache for ./ok.yaml was not created", c.cache)
	}
	if _, ok := c.cache["./workflow.yaml"]; !ok {
		t.Error("Cache for WorkflowCallEvent was not created", c.cache)
	}
}

func TestReusableWorkflowNullCache(t *testing.T) {
	c := newNullLocalReusableWorkflowCache(io.Discard)
	e := &WorkflowCallEvent{Inputs: []*WorkflowCallEventInput{}}

	// This should do nothing
	c.WriteWorkflowCallEvent(filepath.Join("foo", "test.yaml"), e)

	m, err := c.FindMetadata("./foo/test.yaml")
	if m != nil {
		t.Errorf("metadata should never be found with null cache: %v", m)
	}
	if err != nil {
		t.Errorf("error should not happen since the cache simply doesn't hit: %v", err)
	}
}

func TestReusableWorkflowCacheFactory(t *testing.T) {
	cwd := filepath.Join("path", "to", "project1")
	f := NewLocalReusableWorkflowCacheFactory(cwd, nil)

	p1 := &Project{cwd, nil}
	c1 := f.GetCache(p1)

	p2 := &Project{filepath.Join("path", "to", "project2"), nil}
	c2 := f.GetCache(p2)
	if c1 == c2 {
		t.Errorf("Different cache was not created: %v", c1)
	}

	c3 := f.GetCache(p1)
	if c1 != c3 {
		t.Errorf("Same cache was not used: %v vs %v", c1, c3)
	}

	c4 := f.GetCache(nil)
	if c4.proj != nil {
		t.Errorf("Null cache should be returned when project is nil: %v", c4)
	}
}
