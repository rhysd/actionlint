package actionlint

import (
	"testing"
)

func TestRuleAction_ReusableWorkflow(t *testing.T) {
	tests := []struct {
		name    string
		uses    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Local reusable workflow",
			uses:    "./path/to/workflow.yml",
			wantErr: true,
			errMsg:  "The step uses \"./path/to/workflow.yml\" which appears to be a reusable workflow. Reusable workflows cannot be used as steps. Use 'jobs.<job_id>.uses' to call reusable workflows at the job level instead.",
		},
		{
			name:    "Remote reusable workflow",
			uses:    "octo-org/this-repo/.github/workflows/workflow.yml@main",
			wantErr: true,
			errMsg:  "The step uses \"octo-org/this-repo/.github/workflows/workflow.yml@main\" which appears to be a reusable workflow. Reusable workflows cannot be used as steps. Use 'jobs.<job_id>.uses' to call reusable workflows at the job level instead.",
		},
		{
			name:    "Workflow file with yaml extension",
			uses:    "octo-org/this-repo/.github/workflows/workflow.yaml@main",
			wantErr: true,
			errMsg:  "The step uses \"octo-org/this-repo/.github/workflows/workflow.yaml@main\" which appears to be a reusable workflow. Reusable workflows cannot be used as steps. Use 'jobs.<job_id>.uses' to call reusable workflows at the job level instead.",
		},
		{
			name:    "Valid action",
			uses:    "actions/checkout@v4",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := NewRuleAction(nil)
			step := &Step{
				Exec: &ExecAction{
					Uses: &String{
						Value: tt.uses,
						Pos:   &Pos{},
					},
				},
			}

			err := rule.VisitStep(step)
			if err != nil {
				t.Errorf("Unexpected error while visiting step: %s", err)
				return
			}

			errs := rule.Errs()
			if tt.wantErr {
				if len(errs) == 0 {
					t.Errorf("Expected error, but got none")
				} else if errs[0].Message != tt.errMsg {
					t.Errorf("Expected error message %q, but got %q", tt.errMsg, errs[0].Message)
				}
			} else {
				if len(errs) > 0 {
					t.Errorf("Unexpected errors: %v", errs)
				}
			}
		})
	}
}
