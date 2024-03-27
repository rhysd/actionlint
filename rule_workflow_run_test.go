package actionlint

import (
	"reflect"
	"testing"
)

func TestRuleWorkflowRunCheck(t *testing.T) {
	type testWorkflow struct {
		path string
		node *Workflow
	}

	type expectedErr struct {
		path      string
		workflows []*testWorkflow
	}

	pos := &Pos{1, 1}

	expectErr := func(expected ...expectedErr) map[string][]*Error {
		errs := make(map[string][]*Error)
		for _, e := range expected {
			for _, w := range e.workflows {
				name := w.node.computeName(w.path)
				errs[e.path] = append(errs[e.path], errorfAt(pos, "workflow_run", "Workflow %q is not defined", name))
			}
		}
		return errs
	}

	w1 := &testWorkflow{path: "repo/w1.yml", node: &Workflow{}}
	w2 := &testWorkflow{path: "repo/w2.yml", node: &Workflow{Name: &String{Value: "A Workflow", Pos: pos}}}
	w3 := &testWorkflow{path: "repo/w3.yml", node: &Workflow{
		Name: &String{Value: "anotherWorkflow", Pos: &Pos{1, 1}},
		On: []Event{&WebhookEvent{
			Hook: &String{Value: "workflow_run"},
			Workflows: []*String{
				{Value: "w1", Pos: pos},
				{Value: "A Workflow", Pos: pos},
			}}},
	}}

	tests := []struct {
		name       string
		workflows  []*testWorkflow
		expectErrs map[string][]*Error
	}{
		{"only w1", []*testWorkflow{w1}, expectErr()},
		{"only w2", []*testWorkflow{w2}, expectErr()},
		{"w1 and w2", []*testWorkflow{w1, w2}, expectErr()},
		{"only w3", []*testWorkflow{w3}, expectErr(expectedErr{"repo/w3.yml", []*testWorkflow{w1, w2}})},
		{"w1 and w3", []*testWorkflow{w1, w3}, expectErr(expectedErr{"repo/w3.yml", []*testWorkflow{w2}})},
		{"w2 and w3", []*testWorkflow{w2, w3}, expectErr(expectedErr{"repo/w3.yml", []*testWorkflow{w1}})},
		{"all", []*testWorkflow{w1, w2, w3}, expectErr()},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := NewRuleWorkflowRun()
			for _, w := range tc.workflows {
				if err := r.VisitWorkflowPre(w.path, w.node); err != nil {
					t.Fatal(err)
				}
			}

			errs := r.ComputeMissingReferences()

			if !reflect.DeepEqual(tc.expectErrs, errs) {
				t.Fatalf("Wanted %d errors but have: %v", len(tc.expectErrs), errs)
			}
		})
	}
}
