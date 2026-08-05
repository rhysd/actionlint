package actionlint

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCommandMain(t *testing.T) {
	var output bytes.Buffer

	// Create command instance populating stdin/stdout/stderr
	cmd := Command{
		Stdin:  os.Stdin,
		Stdout: &output,
		Stderr: &output,
	}

	// Run the command end-to-end. Note that given args should contain program name
	workflow := filepath.Join("testdata", "examples", "main.yaml")
	status := cmd.Main([]string{"actionlint", "--shellcheck=", "--pyflakes=", "--ignore", `label .+ is unknown\.`, workflow})

	if status != 1 {
		t.Fatal("exit status should be 1 but got", status)
	}

	out := output.String()

	for _, s := range []string{
		"main.yaml:3:5:",
		"unexpected key \"branch\" for \"push\" section",
		"^~~~~~~~~~~~~~~",
	} {
		if !strings.Contains(out, s) {
			t.Errorf("output should contain %q: %q", s, out)
		}
	}

	if strings.Contains(out, "[runner-label]") {
		t.Errorf("runner-label rule should be ignored by --ignore but it is included in output: %q", out)
	}
}

func TestCommandCompletion(t *testing.T) {
	shells := []struct {
		name    string
		contain string
	}{
		{"bash", "_actionlint"},
		{"zsh", "#compdef actionlint"},
		{"fish", "complete -c actionlint"},
		{"powershell", "Register-ArgumentCompleter"},
	}

	for _, tc := range shells {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			cmd := Command{
				Stdin:  os.Stdin,
				Stdout: &stdout,
				Stderr: &stderr,
			}
			status := cmd.Main([]string{"actionlint", "completion", tc.name})
			if status != ExitStatusSuccessNoProblem {
				t.Fatalf("expected exit status 0 but got %d: stderr=%q", status, stderr.String())
			}
			out := stdout.String()
			if !strings.Contains(out, tc.contain) {
				t.Errorf("completion output for %s should contain %q but got: %q", tc.name, tc.contain, out)
			}
		})
	}
}

func TestCommandCompletionInvalidShell(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := Command{
		Stdin:  os.Stdin,
		Stdout: &stdout,
		Stderr: &stderr,
	}
	status := cmd.Main([]string{"actionlint", "completion", "invalid"})
	if status != ExitStatusInvalidCommandOption {
		t.Fatalf("expected exit status %d but got %d", ExitStatusInvalidCommandOption, status)
	}
}

func TestCommandCompletionMissingShell(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := Command{
		Stdin:  os.Stdin,
		Stdout: &stdout,
		Stderr: &stderr,
	}
	status := cmd.Main([]string{"actionlint", "completion"})
	if status != ExitStatusInvalidCommandOption {
		t.Fatalf("expected exit status %d but got %d", ExitStatusInvalidCommandOption, status)
	}
}
