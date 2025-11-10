package actionlint

import (
	"testing"
)

func TestTimeoutChecks(t *testing.T) {
	enforceField := NewRuleTimeoutCheck()
	enforceField.SetConfig(&Config{
		TimeoutMinutes: TimeoutMinutesConfig{
			Required: true,
		},
	})

	enforceFieldValue := NewRuleTimeoutCheck()
	enforceFieldValue.SetConfig(&Config{
		TimeoutMinutes: TimeoutMinutesConfig{
			Required:   true,
			MaxMinutes: 30,
		},
	})

	noTimeout := &Job{
		Pos:  &Pos{Line: 1, Col: 1},
		Name: &String{Value: "no-timeout"},
		Steps: []*Step{
			{
				Name: &String{Value: "do-something"},
				Exec: &ExecAction{Uses: &String{Value: "actions/checkout@v4"}},
			},
		},
	}

	highTimeout := &Job{
		Pos:  &Pos{Line: 1, Col: 1},
		Name: &String{Value: "high-timeout"},
		Steps: []*Step{
			{
				Name: &String{Value: "do-something"},
				Exec: &ExecAction{Uses: &String{Value: "actions/checkout@v4"}},
			},
		},
		TimeoutMinutes: &Float{Value: 45},
	}
	goodTimeout := &Job{
		Pos:  &Pos{Line: 1, Col: 1},
		Name: &String{Value: "good-timeout"},
		Steps: []*Step{
			{
				Name: &String{Value: "do-something"},
				Exec: &ExecAction{Uses: &String{Value: "actions/checkout@v4"}},
			},
		},
		TimeoutMinutes: &Float{Value: 15},
	}

	reusableWorkflow := &Job{
		Pos:  &Pos{Line: 1, Col: 1},
		Name: &String{Value: "reusable-workflow"},
	}

	tests := []struct {
		name   string
		config *RuleTimeoutCheck
		job    *Job
		expect []*Error
	}{
		{"no timeout fail", enforceField, noTimeout, []*Error{{Message: "You must have a timeout-minutes set to avoid overspend.", Line: 1, Column: 1}}},
		{"no timeout pass", enforceField, highTimeout, []*Error{}},
		{"high timeout fail", enforceFieldValue, highTimeout, []*Error{{Message: "Your timeout-minutes is greater than 30 minutes.", Line: 1, Column: 1}}},
		{"good timeout value pass", enforceFieldValue, goodTimeout, []*Error{}},
		{"reusable workflow pass", enforceField, reusableWorkflow, []*Error{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.config.errs = nil
			tt.config.VisitJobPre(tt.job)
			errs := tt.config.Errs()

			if len(errs) != len(tt.expect) {
				t.Fatalf("expected %d errors but got %d: %+v", len(tt.expect), len(errs), errs)
			}
			for i, e := range errs {
				if e.Message != tt.expect[i].Message {
					t.Errorf("expected error message %q but got %q", tt.expect[i].Message, e.Message)
				}
				if e.Line != tt.expect[i].Line || e.Column != tt.expect[i].Column {
					t.Errorf("expected error position %d:%d but got %d:%d", tt.expect[i].Line, tt.expect[i].Column, e.Line, e.Column)
				}
			}
		})
	}
}
