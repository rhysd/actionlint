package actionlint

import (
	"testing"
)

func TestRuleShellSyntaxDetectsErrors(t *testing.T) {
	tests := []struct {
		name   string
		script string
		shell  string
		want   bool // true if error expected
	}{
		{
			name:   "missing fi in bash",
			script: "if [ -f file ]; then\n  echo ok\n",
			shell:  "bash",
			want:   true,
		},
		{
			name:   "missing done in for loop",
			script: "for i in 1 2 3; do\n  echo $i\n",
			shell:  "bash",
			want:   true,
		},
		{
			name:   "missing done in while loop",
			script: "while true; do\n  echo loop\n",
			shell:  "bash",
			want:   true,
		},
		{
			name:   "valid if statement",
			script: "if [ -f file ]; then\n  echo ok\nfi\n",
			shell:  "bash",
			want:   false,
		},
		{
			name:   "valid for loop",
			script: "for i in 1 2 3; do\n  echo $i\ndone\n",
			shell:  "bash",
			want:   false,
		},
		{
			name:   "valid sh script",
			script: "if [ -f file ]; then\n  echo ok\nfi\n",
			shell:  "sh",
			want:   false,
		},
		{
			name:   "missing fi in sh",
			script: "if [ -f file ]; then\n  echo ok\n",
			shell:  "sh",
			want:   true,
		},
		{
			name:   "pwsh is skipped",
			script: "if ($true",
			shell:  "pwsh",
			want:   false,
		},
		{
			name:   "python is skipped",
			script: "if True\n  print()",
			shell:  "python",
			want:   false,
		},
		{
			name:   "custom bash shell",
			script: "if [ -f file ]; then\n  echo ok\n",
			shell:  "bash -e {0}",
			want:   true,
		},
		{
			name:   "custom sh shell",
			script: "if [ -f file ]; then\n  echo ok\n",
			shell:  "sh -e {0}",
			want:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rule := NewRuleShellSyntax()
			pos := &Pos{Line: 1, Col: 1}
			rule.checkShellSyntax(tc.script, tc.shell, pos)

			errs := rule.Errs()
			hasError := len(errs) > 0

			if hasError != tc.want {
				if tc.want {
					t.Errorf("expected error but got none for script: %q", tc.script)
				} else {
					t.Errorf("expected no error but got: %v", errs)
				}
			}
		})
	}
}

func TestRuleShellSyntaxSanitizesExpressions(t *testing.T) {
	// Test that ${{ }} expressions are properly sanitized before parsing
	rule := NewRuleShellSyntax()
	pos := &Pos{Line: 1, Col: 1}

	// Script with expression that would cause parse error if not sanitized
	script := `if [ "${{ inputs.flag }}" = "true" ]; then
  echo "Flag is true"
fi
`
	rule.checkShellSyntax(script, "bash", pos)

	errs := rule.Errs()
	if len(errs) > 0 {
		t.Errorf("expected no error for script with expressions, but got: %v", errs)
	}
}

func TestRuleShellSyntaxShellDetection(t *testing.T) {
	tests := []struct {
		name          string
		workflowShell string
		jobShell      string
		stepShell     string
		runnerShell   string
		want          string
	}{
		{
			name: "default is bash",
			want: "bash",
		},
		{
			name:          "workflow default",
			workflowShell: "sh",
			want:          "sh",
		},
		{
			name:     "job default",
			jobShell: "sh",
			want:     "sh",
		},
		{
			name:      "step shell",
			stepShell: "sh",
			want:      "sh",
		},
		{
			name:      "step overrides job",
			jobShell:  "bash",
			stepShell: "sh",
			want:      "sh",
		},
		{
			name:          "job overrides workflow",
			workflowShell: "bash",
			jobShell:      "sh",
			want:          "sh",
		},
		{
			name:        "runner shell",
			runnerShell: "pwsh",
			want:        "pwsh",
		},
		{
			name:          "workflow overrides runner",
			workflowShell: "bash",
			runnerShell:   "pwsh",
			want:          "bash",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rule := NewRuleShellSyntax()
			rule.workflowShell = tc.workflowShell
			rule.jobShell = tc.jobShell
			rule.runnerShell = tc.runnerShell

			exec := &ExecRun{}
			if tc.stepShell != "" {
				exec.Shell = &String{Value: tc.stepShell}
			}

			got := rule.getShellName(exec)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
