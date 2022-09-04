package actionlint

import (
	"fmt"
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
				Inputs: map[string]*ReusableWorkflowMetadataInput{
					"i": {
						Type: StringType{},
					},
				},
				Outputs: map[string]struct{}{
					"o": {},
				},
				Secrets: map[string]ReusableWorkflowMetadataSecretRequired{
					"x": false,
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
			        requried: true
			`,
			want: &ReusableWorkflowMetadata{
				Inputs: map[string]*ReusableWorkflowMetadataInput{
					"a": {
						Type: StringType{},
					},
					"b": {
						Type: NumberType{},
					},
					"c": {
						Type: BoolType{},
					},
					"d": nil,
					"e": {
						Type: StringType{},
					},
					"f": {
						Type:     StringType{},
						Required: true,
					},
					"g": {
						Type: StringType{},
					},
					"h": {
						Type: StringType{},
					},
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
				Inputs: map[string]*ReusableWorkflowMetadataInput{
					"i": {
						Type: StringType{},
					},
				},
				Outputs: map[string]struct{}{
					"o": {},
				},
				Secrets: map[string]ReusableWorkflowMetadataSecretRequired{
					"x": false,
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
				Secrets: map[string]ReusableWorkflowMetadataSecretRequired{
					"x": false,
					"y": false,
					"z": true,
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
	}

	for _, tc := range tests {
		t.Run(tc.what, func(t *testing.T) {
			src := strings.TrimSpace(tc.src)
			src = strings.ReplaceAll(src, "\t", "")

			m, err := parseReusableWorkflowMetadata([]byte(src))
			if err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(m, tc.want) {
				t.Fatal("Parse result is unexpected. diff:\n" + cmp.Diff(m, tc.want))
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
	Inputs: map[string]*ReusableWorkflowMetadataInput{
		"input1": {Type: StringType{}},
		"input2": {Type: BoolType{}, Required: true},
	},
	Outputs: map[string]struct{}{
		"output1": {},
	},
	Secrets: map[string]ReusableWorkflowMetadataSecretRequired{
		"secret1": false,
		"secret2": true,
	},
}

func TestReusableWorkflowCacheFindMetadataOK(t *testing.T) {
	proj := &Project{filepath.Join("testdata", "reusable_workflow_metadata"), nil}
	c := NewLocalReusableWorkflowCache(proj, "", nil)

	m, err := c.FindMetadata("./ok.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(m, testReusableWorkflowWantedMetadata) {
		t.Fatal(cmp.Diff(m, testReusableWorkflowWantedMetadata))
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
			want: " is invalid: yaml:",
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
			what: "not existing workflow",
			proj: p,
			spec: "./this-workflow-does-not-exist.yaml",
		},
		{
			what: "external workflow",
			proj: p,
			spec: "repo/owner/workflow@main",
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
		want   map[string]*ReusableWorkflowMetadataInput
	}{
		{
			what: "type of inputs",
			inputs: map[string]*WorkflowCallEventInput{
				"string_input":  {Type: WorkflowCallEventInputTypeString},
				"number_input":  {Type: WorkflowCallEventInputTypeNumber},
				"bool_input":    {Type: WorkflowCallEventInputTypeBoolean},
				"unknown_input": {},
			},
			want: map[string]*ReusableWorkflowMetadataInput{
				"string_input":  {Type: StringType{}},
				"number_input":  {Type: NumberType{}},
				"bool_input":    {Type: BoolType{}},
				"unknown_input": {Type: AnyType{}},
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
			want: map[string]*ReusableWorkflowMetadataInput{
				"unspecified":          {Required: false, Type: AnyType{}},
				"not_required":         {Required: false, Type: AnyType{}},
				"required":             {Required: true, Type: AnyType{}},
				"required_but_default": {Required: false, Type: AnyType{}},
				"expression":           {Required: false, Type: AnyType{}},
			},
		},
		{
			what:   "empty",
			inputs: map[string]*WorkflowCallEventInput{},
			want:   map[string]*ReusableWorkflowMetadataInput{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.what, func(t *testing.T) {
			cwd := filepath.Join("path", "to", "project")
			proj := &Project{cwd, nil}
			c := NewLocalReusableWorkflowCache(proj, cwd, nil)
			e := &WorkflowCallEvent{Inputs: map[*String]*WorkflowCallEventInput{}}
			for n, i := range tc.inputs {
				e.Inputs[&String{Value: n, Pos: &Pos{}}] = i
			}

			c.WriteWorkflowCallEvent(filepath.Join("foo", "test.yaml"), e)

			m, ok := c.readCache("./foo/test.yaml")
			if !ok {
				t.Fatal("Event was not converted to event")
			}

			if !cmp.Equal(m.Inputs, tc.want) {
				t.Error(cmp.Diff(m.Inputs, tc.want))
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
	}
	for _, outputs := range tests {
		t.Run(fmt.Sprintf("%s", outputs), func(t *testing.T) {
			cwd := filepath.Join("path", "to", "project")
			proj := &Project{cwd, nil}
			c := NewLocalReusableWorkflowCache(proj, cwd, nil)
			e := &WorkflowCallEvent{Outputs: map[*String]*WorkflowCallEventOutput{}}
			for _, o := range outputs {
				e.Outputs[&String{Value: o, Pos: &Pos{}}] = &WorkflowCallEventOutput{}
			}

			c.WriteWorkflowCallEvent(filepath.Join("foo", "test.yaml"), e)

			m, ok := c.readCache("./foo/test.yaml")
			if !ok {
				t.Fatal("Event was not converted to event")
			}

			want := map[string]struct{}{}
			for _, o := range outputs {
				want[o] = struct{}{}
			}

			if !cmp.Equal(m.Outputs, want) {
				t.Error(cmp.Diff(m.Outputs, want))
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
	}
	for _, secrets := range tests {
		t.Run(fmt.Sprintf("%s", secrets), func(t *testing.T) {
			cwd := filepath.Join("path", "to", "project")
			proj := &Project{cwd, nil}
			c := NewLocalReusableWorkflowCache(proj, cwd, nil)
			e := &WorkflowCallEvent{Secrets: map[*String]*WorkflowCallEventSecret{}}
			for n, r := range secrets {
				e.Secrets[&String{Value: n, Pos: &Pos{}}] = &WorkflowCallEventSecret{Required: r}
			}

			c.WriteWorkflowCallEvent(filepath.Join("foo", "test.yaml"), e)

			m, ok := c.readCache("./foo/test.yaml")
			if !ok {
				t.Fatal("Event was not converted to event")
			}

			want := map[string]ReusableWorkflowMetadataSecretRequired{}
			for n, r := range secrets {
				want[n] = ReusableWorkflowMetadataSecretRequired(r != nil && r.Value)
			}

			if !cmp.Equal(m.Secrets, want) {
				t.Error(cmp.Diff(m.Secrets, want))
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
