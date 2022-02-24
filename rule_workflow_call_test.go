package actionlint

import (
	"strings"
	"testing"
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
			r := NewRuleWorkflowCall()
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

func TestRuleWorkflowCallCheckCannotCallNestedly(t *testing.T) {
	w := &Workflow{
		On: []Event{
			&WorkflowCallEvent{
				Pos: &Pos{
					Line: 12,
					Col:  34,
				},
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

	r := NewRuleWorkflowCall()

	if err := r.VisitWorkflowPre(w); err != nil {
		t.Fatal(err)
	}

	if err := r.VisitJobPre(j); err != nil {
		t.Fatal(err)
	}

	err := r.Errs()[0]
	msg := err.Error()

	want := "reusable workflow cannot be nested"
	if !strings.Contains(msg, want) {
		t.Errorf("error message %q does not contain expected message %q", msg, want)
	}

	want = "line:12,col:34"
	if !strings.Contains(msg, want) {
		t.Errorf("error message %q does not contain proper position %q", msg, want)
	}
}
