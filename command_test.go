package actionlint

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
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
	status := cmd.Main([]string{"actionlint", "-shellcheck=", "-pyflakes=", "-ignore", `label .+ is unknown\.`, workflow})

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
		t.Errorf("runner-label rule should be ignored by -ignore but it is included in output: %q", out)
	}
}

// runCommandMain runs Command.Main with the given arguments and stdin content, and returns its
// exit status, stdout and stderr separately. The summary line goes to stderr while the per-error
// output goes to stdout, so they must be captured apart.
func runCommandMain(t *testing.T, stdin string, args ...string) (int, string, string) {
	t.Helper()

	var stdout, stderr bytes.Buffer
	cmd := Command{
		Stdin:  strings.NewReader(stdin),
		Stdout: &stdout,
		Stderr: &stderr,
	}
	status := cmd.Main(append([]string{"actionlint"}, args...))
	return status, stdout.String(), stderr.String()
}

func TestCommandMainSummary(t *testing.T) {
	var (
		mainWorkflow = filepath.Join("testdata", "examples", "main.yaml")
		missingOn    = filepath.Join("testdata", "err", "missing_on.yaml")
		emptyOn      = filepath.Join("testdata", "err", "empty_on.yaml")
		clean        = filepath.Join("testdata", "ok", "anchors.yaml")
	)
	stdinSrc, err := os.ReadFile(missingOn)
	if err != nil {
		t.Fatal("cannot read test data:", err)
	}

	// shellcheck and pyflakes are disabled to make the number of errors deterministic
	noExt := []string{"-shellcheck=", "-pyflakes="}

	testCases := []struct {
		name            string
		args            []string
		stdin           string
		wantStatus      int
		wantStderr      string // exact expected stderr. Empty string means stderr must be empty
		skipStderrExact bool   // when true, only stderrContains/stderrExcludes are checked
		stderrContains  []string
		stderrExcludes  []string
		stdoutContains  []string
		stdoutEmpty     bool
		wantStdoutLines int // when positive, stdout must have exactly this many lines
	}{
		{
			name:           "summary with errors",
			args:           append(append([]string{}, noExt...), "-summary", mainWorkflow),
			wantStatus:     1,
			wantStderr:     "7 errors in 1 files (1 workflows checked)\n",
			stdoutContains: []string{"main.yaml:3:5:", `unexpected key "branch"`},
		},
		{
			name:           "summary with multiple files",
			args:           append(append([]string{}, noExt...), "-summary", missingOn, emptyOn),
			wantStatus:     1,
			wantStderr:     "2 errors in 2 files (2 workflows checked)\n",
			stdoutContains: []string{`"on" section is missing`, `string should not be empty`},
		},
		{
			name:        "summary on clean workflow",
			args:        append(append([]string{}, noExt...), "-summary", clean),
			wantStatus:  0,
			wantStderr:  "0 errors in 0 files (1 workflows checked)\n",
			stdoutEmpty: true,
		},
		{
			name:            "summary composes with oneline",
			args:            append(append([]string{}, noExt...), "-oneline", "-summary", mainWorkflow),
			wantStatus:      1,
			wantStderr:      "7 errors in 1 files (1 workflows checked)\n",
			stdoutContains:  []string{"main.yaml:3:5:"},
			wantStdoutLines: 7,
		},
		{
			name:           "summary composes with format",
			args:           append(append([]string{}, noExt...), "-format", "{{json .}}", "-summary", mainWorkflow),
			wantStatus:     1,
			wantStderr:     "7 errors in 1 files (1 workflows checked)\n",
			stdoutContains: []string{`"filepath":"testdata/examples/main.yaml"`},
		},
		{
			name:       "summary with double dash option",
			args:       append(append([]string{}, noExt...), "--summary", mainWorkflow),
			wantStatus: 1,
			wantStderr: "7 errors in 1 files (1 workflows checked)\n",
		},
		{
			name:           "no summary without flag",
			args:           append(append([]string{}, noExt...), mainWorkflow),
			wantStatus:     1,
			wantStderr:     "",
			stdoutContains: []string{"main.yaml:3:5:"},
		},
		{
			name:            "summary is not printed on invalid option",
			args:            []string{"-summary", "-bogus"},
			wantStatus:      2,
			skipStderrExact: true,
			stderrContains:  []string{"flag provided but not defined: -bogus"},
			stderrExcludes:  []string{"workflows checked"},
		},
		{
			name:            "summary is not printed on fatal error",
			args:            append(append([]string{}, noExt...), "-summary", "nope.yaml"),
			wantStatus:      3,
			skipStderrExact: true,
			stderrContains:  []string{`could not read "nope.yaml"`},
			stderrExcludes:  []string{"workflows checked"},
		},
		{
			name:           "summary reads stdin",
			args:           append(append([]string{}, noExt...), "-summary", "-"),
			stdin:          string(stdinSrc),
			wantStatus:     1,
			wantStderr:     "1 errors in 1 files (1 workflows checked)\n",
			stdoutContains: []string{"<stdin>:3:1:"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			status, stdout, stderr := runCommandMain(t, tc.stdin, tc.args...)

			if status != tc.wantStatus {
				t.Errorf("exit status should be %d but got %d. stderr: %q", tc.wantStatus, status, stderr)
			}
			if !tc.skipStderrExact && stderr != tc.wantStderr {
				t.Errorf("stderr should be %q but got %q", tc.wantStderr, stderr)
			}
			for _, s := range tc.stderrContains {
				if !strings.Contains(stderr, s) {
					t.Errorf("stderr should contain %q: %q", s, stderr)
				}
			}
			for _, s := range tc.stderrExcludes {
				if strings.Contains(stderr, s) {
					t.Errorf("stderr should not contain %q: %q", s, stderr)
				}
			}
			if tc.stdoutEmpty && stdout != "" {
				t.Errorf("stdout should be empty but got %q", stdout)
			}
			for _, s := range tc.stdoutContains {
				if !strings.Contains(stdout, s) {
					t.Errorf("stdout should contain %q: %q", s, stdout)
				}
			}
			if tc.wantStdoutLines > 0 {
				lines := strings.Count(stdout, "\n")
				if lines != tc.wantStdoutLines {
					t.Errorf("stdout should have %d lines but got %d: %q", tc.wantStdoutLines, lines, stdout)
				}
			}
		})
	}
}

func TestCommandMainSummaryRepository(t *testing.T) {
	// The working directory of tests is the project root, so actionlint finds .github/workflows
	// itself. The number of workflow files in the repository may change, so this test checks the
	// shape of the summary line and the consistency between the counts and the exit status.
	status, stdout, stderr := runCommandMain(t, "", "-shellcheck=", "-pyflakes=", "-summary")

	m := regexp.MustCompile(`^(\d+) errors in (\d+) files \((\d+) workflows checked\)\n$`).FindStringSubmatch(stderr)
	if m == nil {
		t.Fatalf("stderr should be a single summary line but got %q", stderr)
	}
	found, files, checked := m[1], m[2], m[3]
	if n, _ := strconv.Atoi(checked); n < 1 {
		t.Errorf("workflows checked should be positive but got %q", checked)
	}
	if n, _ := strconv.Atoi(files); n > 0 {
		if e, _ := strconv.Atoi(found); n > e {
			t.Errorf("files (%q) should not be more than errors (%q)", files, found)
		}
	}
	wantStatus := 0
	if n, _ := strconv.Atoi(found); n > 0 {
		wantStatus = 1
	}
	if status != wantStatus {
		t.Errorf("exit status should be %d but got %d. stderr: %q", wantStatus, status, stderr)
	}
	if wantStatus == 1 && stdout == "" {
		t.Error("stdout should have per-error output because some errors were found")
	}
}
