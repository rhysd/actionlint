package actionlint

import (
	"reflect"
	"testing"
)

func TestRuleWorkflowRunCheck(t *testing.T) {
	type expectedErr struct {
		path      string
		workflows []*Workflow
	}

	pos := &Pos{1, 1}

	expectErr := func(expected ...expectedErr) map[string][]*Error {
		errs := make(map[string][]*Error)
		for _, e := range expected {
			for _, w := range e.workflows {
				name := w.computeName()
				errs[e.path] = append(errs[e.path], errorfAt(pos, "workflow_run", "Workflow %q is not defined", name))
			}
		}
		return errs
	}

	w1 := &Workflow{Path: "repo/w1.yml"}
	w2 := &Workflow{Path: "repo/w2.yml", Name: &String{Value: "A Workflow", Pos: pos}}
	w3 := &Workflow{
		Path: "repo/w3.yml",
		Name: &String{Value: "anotherWorkflow", Pos: &Pos{1, 1}},
		On: []Event{&WebhookEvent{
			Hook: &String{Value: "workflow_run"},
			Workflows: []*String{
				{Value: "w1", Pos: pos},
				{Value: "A Workflow", Pos: pos},
			}}},
	}

	tests := []struct {
		name       string
		workflows  []*Workflow
		expectErrs map[string][]*Error
	}{
		{"only w1", []*Workflow{w1}, expectErr()},
		{"only w2", []*Workflow{w2}, expectErr()},
		{"w1 and w2", []*Workflow{w1, w2}, expectErr()},
		{"only w3", []*Workflow{w3}, expectErr(expectedErr{"repo/w3.yml", []*Workflow{w1, w2}})},
		{"w1 and w3", []*Workflow{w1, w3}, expectErr(expectedErr{"repo/w3.yml", []*Workflow{w2}})},
		{"w2 and w3", []*Workflow{w2, w3}, expectErr(expectedErr{"repo/w3.yml", []*Workflow{w1}})},
		{"all", []*Workflow{w1, w2, w3}, expectErr()},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := NewRuleWorkflowRun(NewRuleWorkflowRunSharedState())
			for _, w := range tc.workflows {
				if err := r.VisitWorkflowPre(w); err != nil {
					t.Fatal(err)
				}
			}

			errs := r.ComputeMissingReferences()

			if len(tc.expectErrs) != len(errs) {
				t.Fatalf("Wanted %d errors but have %d", len(tc.expectErrs), len(errs))
			}
			for key, expected := range tc.expectErrs {
				actual := errs[key]
				if !reflect.DeepEqual(expected, actual) {
					t.Fatalf("Errors does not match:\nWanted: %v\nBut have: %v", expected, actual)
				}
			}
		})
	}
}
