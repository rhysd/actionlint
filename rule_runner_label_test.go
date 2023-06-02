package actionlint

import (
	"strings"
	"testing"
)

func TestRuleRunnerLabelCheckLabels(t *testing.T) {
	testCases := []struct {
		what   string
		labels []string
		matrix []string
		known  []string
		errs   []string
	}{
		// Normal cases
		{
			what:   "simple GH-hosted Linux runner label",
			labels: []string{"ubuntu-20.04"},
		},
		{
			what:   "simple GH-hosted Windows runner label",
			labels: []string{"windows-latest"},
		},
		{
			what:   "simple GH-hosted macOS runner label",
			labels: []string{"macos-11.0"},
		},
		{
			what:   "simple GH-hosted runner label in upper case",
			labels: []string{"macOS-11"},
		},
		{
			what:   "self-hosted Linux runner",
			labels: []string{"self-hosted", "linux"},
		},
		{
			what:   "self-hosted all Linux runner labels",
			labels: []string{"self-hosted", "linux", "ubuntu-22.04", "ubuntu-latest"},
		},
		{
			what:   "self-hosted all macOS runner labels",
			labels: []string{"self-hosted", "macOS", "macOS-latest", "macOS-12.0", "macOS-12"},
		},
		{
			what:   "self-hosted Linux runner in upper case",
			labels: []string{"SELF-HOSTED", "LINUX"},
		},
		{
			what:   "self-hosted macOS runner",
			labels: []string{"self-hosted", "macos", "arm64"},
		},
		{
			what:   "self-hosted runner with GH-hosted runner label",
			labels: []string{"self-hosted", "ubuntu-20.04"},
		},
		{
			what:   "multiple labels for GH-hosted runner",
			labels: []string{"ubuntu-latest", "ubuntu-22.04"},
		},
		{
			what:   "user-defined labels",
			labels: []string{"self-hosted", "foo", "bar", "linux"},
			known:  []string{"foo", "bar"},
		},
		{
			what:   "matrix",
			labels: []string{"${{matrix.os}}"},
			matrix: []string{"ubuntu-latest", "windows-latest"},
		},
		{
			what:   "matrix at second label",
			labels: []string{"self-hosted", "${{matrix.os}}"},
			matrix: []string{"linux", "windows"},
		},
		{
			what:   "matrix at first label",
			labels: []string{"${{matrix.os}}", "ubuntu-22.04"},
			matrix: []string{"ubuntu-latest"},
		},
		{
			what:   "user-defined label with matrix",
			labels: []string{"self-hosted", "${{matrix.os}}"},
			matrix: []string{"foo", "bar"},
			known:  []string{"foo", "bar"},
		},
		{
			what:   "cannot check label: prefix",
			labels: []string{"foo-${{matrix.os}}"},
			matrix: []string{"ubuntu-latest"},
		},
		{
			what:   "cannot check label: siffux",
			labels: []string{"${{matrix.os}}-bar"},
			matrix: []string{"ubuntu-latest"},
		},
		{
			what:   "cannot check label: not a matrix",
			labels: []string{"${{fromJSON(env.FOO).os}}"},
			matrix: []string{"ubuntu-latest"},
		},
		{
			what:   "cannot check label: not a matrix 2",
			labels: []string{"${{env.os}}"},
			matrix: []string{"ubuntu-latest"},
		},
		{
			what:   "cannot check label: matrix object",
			labels: []string{"${{matrix}}"},
			matrix: []string{"ubuntu-latest"},
		},
		{
			what:   "cannot parse expression",
			labels: []string{"${{}}"},
			matrix: []string{"ubuntu-latest"},
		},
		{
			what:   "give up checking matrix value containing expression",
			labels: []string{"${{matrix.os}}"},
			matrix: []string{"ubuntu-latest", "${{env.OS}}"},
		},
		{
			what:   "use matrix value but no matrix exist",
			labels: []string{"${{matrix.os}}"},
		},
		// TODO: Add tests for 'include:'
		// TODO: Check matrix with 'include:'

		// Error cases
		{
			what:   "undefined label",
			labels: []string{"linux-latest"},
			errs:   []string{`"linux-latest" is unknown`},
		},
		{
			what:   "undefined self-hosted label",
			labels: []string{"self-hosted", "foo"},
			errs:   []string{`"foo" is unknown`},
		},
		{
			what:   "GH-hosted runner labels conflict",
			labels: []string{"ubuntu-latest", "windows-latest"},
			errs:   []string{`label "windows-latest" conflicts with label "ubuntu-latest"`},
		},
		{
			what:   "self-hosted runner labels conflict",
			labels: []string{"self-hosted", "windows", "linux"},
			errs:   []string{`label "linux" conflicts with label "windows"`},
		},
		{
			what:   "self-hosted runner labels conflict with GH-hosted runner label",
			labels: []string{"self-hosted", "windows", "macOS-latest"},
			errs:   []string{`label "macOS-latest" conflicts with label "windows"`},
		},
		{
			what:   "GH-hosted labels multiple conflicts",
			labels: []string{"ubuntu-latest", "windows-latest", "macos-latest"},
			errs: []string{
				`label "windows-latest" conflicts with label "ubuntu-latest"`,
				`label "macos-latest" conflicts with label "ubuntu-latest"`,
			},
		},
		{
			what:   "GH-hosted labels conflict mixed with self-hosted runner labels",
			labels: []string{"self-hosted", "ubuntu-latest", "x64", "windows-latest", "foo"},
			known:  []string{"foo"},
			errs:   []string{`label "windows-latest" conflicts with label "ubuntu-latest"`},
		},
		{
			what:   "GH-hosted labels conflict ignore case",
			labels: []string{"macOS-latest", "Windows-latest"},
			errs:   []string{`label "Windows-latest" conflicts with label "macOS-latest"`},
		},
		{
			what:   "GH-hosted labels conflict with matrix at second label",
			labels: []string{"ubuntu-latest", "${{matrix.os}}"},
			matrix: []string{"windows-latest", "macos-latest"},
			errs: []string{
				`label "windows-latest" conflicts with label "ubuntu-latest"`,
				`label "macos-latest" conflicts with label "ubuntu-latest"`,
			},
		},
		{
			what:   "GH-hosted labels conflict with matrix at first label",
			labels: []string{"${{matrix.os}}", "ubuntu-latest"},
			matrix: []string{"windows-latest", "macos-latest"},
			errs:   []string{`label "ubuntu-latest" conflicts with label`},
		},
		{
			what:   "GH-hosted labels conflicts with multiple matrixes",
			labels: []string{"${{matrix.os}}", "${{matrix.os}}"},
			matrix: []string{"windows-latest", "macos-latest"},
			errs: []string{
				`label "windows-latest" conflicts with label "macos-latest"`,
				`label "macos-latest" conflicts with label "windows-latest"`,
			},
		},
		{
			what:   "Linux labels conflict",
			labels: []string{"ubuntu-latest", "ubuntu-20.04"},
			errs:   []string{`label "ubuntu-20.04" conflicts with label "ubuntu-latest"`},
		},
		{
			what:   "macOS labels conflict",
			labels: []string{"macos-11", "macos-12"},
			errs:   []string{`label "macos-12" conflicts with label "macos-11"`},
		},
		// TODO: Add error tests for 'include:'
	}

	for _, tc := range testCases {
		t.Run(tc.what, func(t *testing.T) {
			pos := &Pos{}
			labels := make([]*String, 0, len(tc.labels))
			for _, l := range tc.labels {
				labels = append(labels, &String{l, false, pos})
			}
			node := &Job{
				RunsOn: &Runner{
					Labels: labels,
				},
			}

			if tc.matrix != nil {
				n := &String{"os", false, pos}
				row := make([]RawYAMLValue, 0, len(tc.matrix))
				for _, m := range tc.matrix {
					row = append(row, &RawYAMLString{m, pos})
				}
				st := &Strategy{
					Matrix: &Matrix{
						Rows: map[string]*MatrixRow{
							"os": {
								Name:   n,
								Values: row,
							},
						},
					},
					Pos: pos,
				}
				node.Strategy = st
			}

			rule := NewRuleRunnerLabel()
			cfg := Config{}
			cfg.SelfHostedRunner.Labels = tc.known
			rule.SetConfig(&cfg)
			if err := rule.VisitJobPre(node); err != nil {
				t.Fatal(err)
			}

			errs := rule.Errs()
			if len(errs) != len(tc.errs) {
				t.Fatalf("%d error(s) are wanted but got %d error(s) actually: %v", len(tc.errs), len(errs), errs)
			}
			for i, want := range tc.errs {
				have := errs[i].Error()
				if !strings.Contains(have, want) {
					t.Fatalf("%q is not contained in error message of errs[%d]: %q", want, i, have)
				}
			}
		})
	}
}

func TestRuleRunnerLabelDoNothingOnNoRunsOn(t *testing.T) {
	rule := NewRuleRunnerLabel()
	if err := rule.VisitJobPre(&Job{}); err != nil {
		t.Fatal(err)
	}
	if errs := rule.Errs(); len(errs) > 0 {
		t.Fatalf("%d error(s) occurred: %v", len(errs), errs)
	}
}

func TestRuleRunnerLabelAllGitHubHostedRunnerLabels(t *testing.T) {
	all := []string{}
	all = append(all, allGitHubHostedRunnerLabels...)
	all = append(all, selfHostedRunnerPresetOSLabels...)

	if len(all) != len(defaultRunnerOSCompats) {
		t.Errorf("%d elements in allGitHubHostedRunnerLabels but %d elements in githubHostedRunnerCompats", len(all), len(defaultRunnerOSCompats))
	}
	for _, l := range all {
		if l != strings.ToLower(l) {
			t.Errorf("label %q in allGitHubHostedRunnerLabels is not in lower-case", l)
		}
		if _, ok := defaultRunnerOSCompats[l]; !ok {
			t.Errorf("%q is included in allGitHubHostedRunnerLabels but not included in githubHostedRunnerCompats", l)
		}
	}
}
