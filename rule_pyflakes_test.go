package actionlint

import (
	"strings"
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
			what:     "no default shell",
			isPython: false,
		},
		{
			what:     "workflow default",
			isPython: true,
			workflow: "python",
		},
		{
			what:     "job default",
			isPython: true,
			job:      "python",
		},
		{
			what:     "step shell",
			isPython: true,
			step:     "python",
		},
		{
			what:     "custom shell",
			isPython: true,
			workflow: "python {0}",
		},
		{
			what:     "other shell",
			isPython: false,
			workflow: "pwsh",
		},
		{
			what:     "other custom shell",
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

func TestRulePyflakesParsePyflakesOutputOK(t *testing.T) {
	tests := []struct {
		what  string
		input string
		want  []string
	}{
		{
			what:  "no error",
			input: "",
		},
		{
			what:  "ignore unrelated lines",
			input: "this line\nshould be\nignored\n",
		},
		{
			what:  "single error",
			input: "<stdin>:1:7: undefined name 'foo'\n",
			want: []string{
				":1:2: pyflakes reported issue in this script: 1:7: undefined name 'foo' [pyflakes]",
			},
		},
		{
			what: "multiple errors",
			input: "<stdin>:1:7: undefined name 'foo'\n" +
				"<stdin>:1:7: undefined name 'foo'\n" +
				"<stdin>:1:7: undefined name 'foo'\n",
			want: []string{
				":1:2: pyflakes reported issue in this script: 1:7: undefined name 'foo' [pyflakes]",
				":1:2: pyflakes reported issue in this script: 1:7: undefined name 'foo' [pyflakes]",
				":1:2: pyflakes reported issue in this script: 1:7: undefined name 'foo' [pyflakes]",
			},
		},
		{
			what: "syntax error",
			input: "<stdin>:1:7: unexpected EOF while parsing\n" +
				"print(\n" +
				"      ^\n",
			want: []string{
				":1:2: pyflakes reported issue in this script: 1:7: unexpected EOF while parsing [pyflakes]",
			},
		},
		{
			what: "ignore unrelated lines between multiple errors",
			input: "this line should be ignored\n" +
				"this line should be ignored\n" +
				"<stdin>:1:7: undefined name 'foo'\n" +
				"this line should be ignored\n" +
				"this line should be ignored\n" +
				"<stdin>:1:7: undefined name 'foo'\n" +
				"this line should be ignored\n" +
				"this line should be ignored\n",
			want: []string{
				":1:2: pyflakes reported issue in this script: 1:7: undefined name 'foo' [pyflakes]",
				":1:2: pyflakes reported issue in this script: 1:7: undefined name 'foo' [pyflakes]",
			},
		},
		{
			what: "CRLF",
			input: "<stdin>:1:7: undefined name 'foo'\r\n" +
				"<stdin>:1:7: undefined name 'foo'\r\n",
			want: []string{
				":1:2: pyflakes reported issue in this script: 1:7: undefined name 'foo' [pyflakes]",
				":1:2: pyflakes reported issue in this script: 1:7: undefined name 'foo' [pyflakes]",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.what, func(t *testing.T) {
			r := newRulePyflakes(&externalCommand{})
			stdout := []byte(tc.input)
			pos := &Pos{Line: 1, Col: 2}
			for len(stdout) > 0 {
				o, err := r.parseNextError(stdout, pos)
				if err != nil {
					t.Fatalf("Parse error %q while reading input %q", err, stdout)
				}
				stdout = o
			}
			have := r.Errs()
			if len(have) != len(tc.want) {
				msgs := []string{}
				for _, e := range have {
					msgs = append(msgs, e.Error())
				}
				t.Fatalf("%d errors were expected but got %d errors. got errors are:\n%#v", len(tc.want), len(have), msgs)
			}

			for i, want := range tc.want {
				have := have[i]
				msg := have.Error()
				if !strings.Contains(msg, want) {
					t.Errorf("Error %q does not contain expected message %q", msg, want)
				}
			}
		})
	}
}

func TestRulePyflakesParsePyflakesOutputError(t *testing.T) {
	r := newRulePyflakes(&externalCommand{})
	_, err := r.parseNextError([]byte("<stdin>:1:7: undefined name 'foo'"), &Pos{})
	if err == nil {
		t.Fatal("Error did not happen")
	}
	have := err.Error()
	want := `error message from pyflakes does not end with \n nor \r\n`
	if !strings.Contains(have, want) {
		t.Fatalf("Error %q does not contain expected message %q", have, want)
	}
}
