package actionlint

import (
	"fmt"
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
