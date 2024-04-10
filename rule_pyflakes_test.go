package actionlint

import (
	"testing"
)

func TestRulePyflakesDetectPythonShell(t *testing.T) {
	tests := []struct {
		what     string
		isPython bool
		workflow string // Shell name set at 'defaults' in Workflow node
		job      string // Shell name set at 'defaults' in Job node
		step     string // Shell name set at 'shell' in Step node
	}{
		{
			what: "no default shell",
			isPython: false,
		},
		{
			what: "workflow default",
			isPython: true,
			workflow: "python",
		},
		{
			what: "job default",
			isPython: true,
			job: "python",
		},
		{
			what: "step shell",
			isPython: true,
			step: "python",
		},
		{
			what: "custom shell",
			isPython: true,
			workflow: "python {0}",
		},
		{
			what: "other shell",
			isPython: false,
			workflow: "pwsh",
		},
		{
			what: "other custom shell",
			isPython: false,
			workflow: "bash -e {0}",
		},
	}

	for _, tc := range tests {
		t.Run(tc.what, func(t *testing.T) {
			r := newRulePyflakes(&externalCommand{})

			w := &Workflow{}
			if tc.workflow != "" {
				w.Defaults = &Defaults{
					Run: &DefaultsRun{
						Shell: &String{Value: tc.workflow},
					},
				}
			}
			r.VisitWorkflowPre(w)

			j := &Job{}
			if tc.job != "" {
				j.Defaults = &Defaults{
					Run: &DefaultsRun{
						Shell: &String{Value: tc.job},
					},
				}
			}
			r.VisitJobPre(j)

			e := &ExecRun{}
			if tc.step != "" {
				e.Shell = &String{Value: tc.step}
			}
			if have := r.isPythonShell(e); have != tc.isPython {
				t.Fatalf("Actual isPython=%v but wanted isPython=%v", have, tc.isPython)
			}
		})
	}
}
