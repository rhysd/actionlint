package actionlint

import (
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestRuleWorkflowCallCheckWorkflowCallUsesFormat(t *testing.T) {
	tests := []struct {
		uses string
		ok   bool
	}{
		{"owner/repo/x.yml@ref", true},
		{"owner/repo/x.yml@@", true},
		{"owner/repo/x.yml@release/v1", true},
		{"./path/to/x.yml", true},
		{"${{ env.FOO }}", true},
		{"./path/to/x.yml@ref", false},
		{"/path/to/x.yml@ref", false},
		{"./", false},
		{".", false},
		{"owner/x.yml@ref", false},
		{"owner/repo@ref", false},
		{"owner/repo/x.yml", false},
		{"/repo/x.yml@ref", false},
		{"owner//x.yml@ref", false},
		{"owner/repo/@ref", false},
		{"owner/repo/x.yml@", false},
	}

	for _, tc := range tests {
		t.Run(tc.uses, func(t *testing.T) {
			c := NewLocalReusableWorkflowCache(nil, "", nil)
			r := NewRuleWorkflowCall("", c)
			j := &Job{
				WorkflowCall: &WorkflowCall{
					Uses: &String{
						Value: tc.uses,
						Pos:   &Pos{},
					},
				},
			}
			err := r.VisitJobPre(j)
			if err != nil {
				t.Fatal(err)
			}
			errs := r.Errs()
			if tc.ok && len(errs) > 0 {
				t.Fatalf("Error occurred: %v", errs)
			}
			if !tc.ok {
				if len(errs) > 2 || len(errs) == 0 {
					t.Fatalf("Wanted one error but have: %v", errs)
				}
			}
		})
	}
}

func TestRuleWorkflowCallNestedWorkflowCalls(t *testing.T) {
	w := &Workflow{
		On: []Event{
			&WorkflowCallEvent{
				Pos: &Pos{},
			},
		},
	}

	j := &Job{
		WorkflowCall: &WorkflowCall{
			Uses: &String{
				Value: "o/r/w.yaml@r",
				Pos:   &Pos{},
			},
		},
	}

	c := NewLocalReusableWorkflowCache(nil, "", nil)
	r := NewRuleWorkflowCall("", c)

	if err := r.VisitWorkflowPre(w); err != nil {
		t.Fatal(err)
	}

	if err := r.VisitJobPre(j); err != nil {
		t.Fatal(err)
	}
	errs := r.Errs()

	if len(errs) > 0 {
		t.Fatal("unexpected errors:", errs)
	}
}

func TestRuleWorkflowCallWriteEventNodeToMetadataCache(t *testing.T) {
	s := func(v string) *String {
		return &String{Value: v, Pos: &Pos{}}
	}
	w := &Workflow{
		On: []Event{
			&WorkflowCallEvent{
				Inputs: map[*String]*WorkflowCallEventInput{
					s("input1"): {Type: WorkflowCallEventInputTypeString},
				},
				Outputs: map[*String]*WorkflowCallEventOutput{
					s("output1"): {},
				},
				Secrets: map[*String]*WorkflowCallEventSecret{
					s("secret1"): {},
				},
				Pos: &Pos{},
			},
		},
	}

	cwd := filepath.Join("path", "to", "project")
	c := NewLocalReusableWorkflowCache(&Project{cwd, nil}, cwd, nil)
	r := NewRuleWorkflowCall("test-workflow.yaml", c)

	if err := r.VisitWorkflowPre(w); err != nil {
		t.Fatal(err)
	}

	errs := r.Errs()
	if len(errs) > 0 {
		t.Fatal(errs)
	}

	m, ok := c.readCache("./test-workflow.yaml")
	if !ok {
		t.Fatal("no metadata was created")
	}

	want := &ReusableWorkflowMetadata{
		Inputs: map[string]*ReusableWorkflowMetadataInput{
			"input1": {Type: StringType{}},
		},
		Outputs: map[string]struct{}{
			"output1": {},
		},
		Secrets: map[string]ReusableWorkflowMetadataSecretRequired{
			"secret1": false,
		},
	}

	if !cmp.Equal(want, m) {
		t.Fatal(cmp.Diff(want, m))
	}
}

func TestRuleWorkflowCallCheckReusableWorkflowCall(t *testing.T) {
	metadata := &ReusableWorkflowMetadata{
		Inputs: map[string]*ReusableWorkflowMetadataInput{
			"optional_input": {Type: StringType{}},
			"required_input": {
				Type:     StringType{},
				Required: true,
			},
		},
		Outputs: map[string]struct{}{
			"output": {},
		},
		Secrets: map[string]ReusableWorkflowMetadataSecretRequired{
			"optional_secret": false,
			"required_secret": true,
		},
	}
	cwd := filepath.Join("testdata", "reusable_workflow_metadata")
	cache := NewLocalReusableWorkflowCache(&Project{cwd, nil}, cwd, nil)
	cache.writeCache("./workflow.yaml", metadata)

	tests := []struct {
		what           string
		uses           string
		inputs         []string
		secrets        []string
		inheritSecrets bool
		errs           []string
	}{
		{
			what:    "all",
			uses:    "./workflow.yaml",
			inputs:  []string{"optional_input", "required_input"},
			secrets: []string{"optional_secret", "required_secret"},
		},
		{
			what:    "only required",
			uses:    "./workflow.yaml",
			inputs:  []string{"required_input"},
			secrets: []string{"required_secret"},
		},
		{
			what:    "unknown workflow",
			uses:    "./unknown-workflow.yaml",
			inputs:  []string{"aaa", "bbb"},
			secrets: []string{"xxx", "yyy"},
		},
		{
			what:    "missing required input and secret",
			uses:    "./workflow.yaml",
			inputs:  []string{"optional_input"},
			secrets: []string{"optional_secret"},
			errs: []string{
				"input \"required_input\" is required",
				"secret \"required_secret\" is required",
			},
		},
		{
			what:    "undefined input and secret",
			uses:    "./workflow.yaml",
			inputs:  []string{"required_input", "unknown_input"},
			secrets: []string{"required_secret", "unknown_secret"},
			errs: []string{
				"input \"unknown_input\" is not defined in \"./workflow.yaml\" reusable workflow. defined inputs are \"optional_input\", \"required_input\"",
				"secret \"unknown_secret\" is not defined in \"./workflow.yaml\" reusable workflow. defined secrets are \"optional_secret\", \"required_secret\"",
			},
		},
		{
			what:           "inherit secrets",
			uses:           "./workflow.yaml",
			inputs:         []string{"required_input"},
			secrets:        []string{"unknown_secret", "optional_secret"},
			inheritSecrets: true,
		},
		{
			what:    "read workflow",
			uses:    "./ok.yaml", // Defined in testdata/reusable_workflow_metadata/ok.yaml
			inputs:  []string{"input2"},
			secrets: []string{"secret2"},
		},
		{
			what: "read broken workflow",
			uses: "./broken.yaml", // Defined in testdata/reusable_workflow_metadata/broken.yaml
			errs: []string{
				"error while parsing reusable workflow \"./broken.yaml\"",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.what, func(t *testing.T) {
			r := NewRuleWorkflowCall("this-workflow.yaml", cache)

			w := &Workflow{
				On: []Event{
					&WorkflowCallEvent{
						Pos: &Pos{},
					},
				},
			}
			if err := r.VisitWorkflowPre(w); err != nil {
				t.Fatal(err)
			}

			c := &WorkflowCall{
				Uses:           &String{Value: tc.uses, Pos: &Pos{}},
				Inputs:         map[string]*WorkflowCallInput{},
				Secrets:        map[string]*WorkflowCallSecret{},
				InheritSecrets: tc.inheritSecrets,
			}
			for _, i := range tc.inputs {
				c.Inputs[i] = &WorkflowCallInput{
					Name:  &String{Value: i, Pos: &Pos{}},
					Value: &String{Value: "", Pos: &Pos{}},
				}
			}
			for _, s := range tc.secrets {
				c.Secrets[s] = &WorkflowCallSecret{
					Name:  &String{Value: s, Pos: &Pos{}},
					Value: &String{Value: "", Pos: &Pos{}},
				}
			}

			j := &Job{WorkflowCall: c}
			if err := r.VisitJobPre(j); err != nil {
				t.Fatal(err)
			}

			errs := []string{}
			for _, err := range r.Errs() {
				errs = append(errs, err.Error())
			}
			sort.Strings(errs)

			if len(errs) != len(tc.errs) {
				t.Fatalf(
					"Number of errors is unexpected. %d errors was expected but got %d errors. Expected errors are %v but actual errors are %v",
					len(tc.errs),
					len(errs),
					tc.errs,
					errs,
				)
			}

			for i, have := range errs {
				want := tc.errs[i]
				if !strings.Contains(have, want) {
					t.Errorf("%d-th error is unexpected. %q should be contained in error message %q", i, want, have)
				}
			}
		})
	}
}
