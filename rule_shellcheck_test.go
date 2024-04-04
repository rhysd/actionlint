package actionlint

import (
	"fmt"
	"testing"
)

func TestRuleShellcheckSanitizeExpressionsInScript(t *testing.T) {
	testCases := []struct {
		input string
		want  string
	}{
		{
			"",
			"",
		},
		{
			"foo",
			"foo",
		},
		{
			"${{}}",
			"_____",
		},
		{
			"${{ matrix.foo }}",
			"_________________",
		},
		{
			"aaa ${{ matrix.foo }} bbb",
			"aaa _________________ bbb",
		},
		{
			"${{}}${{}}",
			"__________",
		},
		{
			"p${{a}}q${{b}}r",
			"p______q______r",
		},
		{
			"${{",
			"${{",
		},
		{
			"}}",
			"}}",
		},
		{
			"aaa${{foo",
			"aaa${{foo",
		},
		{
			"a${{b}}${{c",
			"a______${{c",
		},
		{
			"a${{b}}c}}d",
			"a______c}}d",
		},
		{
			"a}}b${{c}}d",
			"a}}b______d",
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d_%s", i, tc.input), func(t *testing.T) {
			have := sanitizeExpressionsInScript(tc.input)
			if tc.want != have {
				t.Fatalf("sanitized result is unexpected.\nwant: %q\nhave: %q", tc.want, have)
			}
		})
	}
}

// Regression for #409
func TestRuleShellcheckDetectShell(t *testing.T) {
	tests := []struct {
		what     string
		want     string
		workflow string // Shell name set at 'defaults' in Workflow node
		job      string // Shell name set at 'defaults' in Job node
		step     string // Shell name set at 'shell' in Step node
		runner   string // Runner name at 'runs-on' in Job node
	}{
		{
			what: "no default shell",
			want: "bash",
		},
		{
			what:     "workflow default",
			want:     "pwsh",
			workflow: "pwsh",
		},
		{
			what: "job default",
			want: "pwsh",
			job:  "pwsh",
		},
		{
			what: "step config",
			want: "pwsh",
			step: "pwsh",
		},
		{
			what:     "job default is more proioritized than workflow",
			want:     "pwsh",
			workflow: "bash",
			job:      "pwsh",
		},
		{
			what:     "step config is more proioritized than job",
			want:     "pwsh",
			workflow: "sh",
			job:      "bash",
			step:     "pwsh",
		},
		{
			what:   "default shell detected from runner",
			want:   "pwsh",
			runner: "windows-latest",
		},
		{
			what:     "workflow default is more proioritized than runner",
			want:     "bash",
			workflow: "bash",
			runner:   "windows-latest",
		},
		{
			what:   "job default is more proioritized than runner",
			want:   "bash",
			job:    "bash",
			runner: "windows-latest",
		},
		{
			what:   "step config is more proioritized than runner",
			want:   "bash",
			step:   "bash",
			runner: "windows-latest",
		},
		{
			what:   "no shell is detected from Ubuntu runner",
			want:   "bash",
			runner: "ubuntu-latest",
		},
	}

	for _, tc := range tests {
		t.Run(tc.what, func(t *testing.T) {
			r := newRuleShellcheck(&externalCommand{})

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
			if tc.runner != "" {
				j.RunsOn = &Runner{
					Labels: []*String{
						{Value: tc.runner},
					},
				}
			}
			r.VisitJobPre(j)

			e := &ExecRun{}
			if tc.step != "" {
				e.Shell = &String{Value: tc.step}
			}
			if s := r.getShellName(e); s != tc.want {
				t.Fatalf("detected shell %q but wanted %q", s, tc.want)
			}
		})
	}
}
