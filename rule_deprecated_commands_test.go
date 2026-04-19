package actionlint

import (
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestRuleDeprecatedCommandsDetectTargetCommands(t *testing.T) {
	tests := []struct {
		what string
		run  string
		want []string
	}{
		{
			what: "save-state",
			run:  "::save-state name=foo::42",
			want: []string{"save-state"},
		},
		{
			what: "set-output",
			run:  "::set-output name=foo::42",
			want: []string{"set-output"},
		},
		{
			what: "set-env",
			run:  "::set-env name=foo::42",
			want: []string{"set-env"},
		},
		{
			what: "add-path",
			run:  "::add-path::/path/to/foo",
			want: []string{"add-path"},
		},
		{
			what: "submatch",
			run:  "hello::set-output name=foo::42 world",
			want: []string{"set-output"},
		},
		{
			what: "multiple same commands",
			run:  "::set-output name=foo::42 ::set-output name=bar::xxx",
			want: []string{"set-output", "set-output"},
		},
		{
			what: "multiple different commands",
			run:  "::set-output name=foo::42 ::add-path::/path/to/foo ::save-state name=foo::42",
			want: []string{"set-output", "add-path", "save-state"},
		},
		{
			what: "multiple submatches",
			run:  "hello::set-output name=foo::42 how ::add-path::/path/to/foo are ::save-state name=foo::42 you",
			want: []string{"set-output", "add-path", "save-state"},
		},
		{
			what: "something between command and arguments",
			run:  "::set-output hello name=foo::42",
			want: []string{},
		},
		{
			what: "multiple spaces",
			run:  "::set-output    		    name=foo::42",
			want: []string{"set-output"},
		},
		{
			what: "empty string",
			run:  "",
			want: []string{},
		},
		{
			what: "no command",
			run:  "echo 'do not use set-output!'",
			want: []string{},
		},
		{
			what: "hyphen and underscore in name",
			run:  "::save-state name=foo_bar-woo::42",
			want: []string{"save-state"},
		},
		{
			what: "invalid name",
			run:  "::save-state name=-foo::42",
			want: []string{},
		},
		{
			what: "different argument",
			run:  "::save-state myname=foo::42",
			want: []string{},
		},
	}
	re := regexp.MustCompile(`\s+workflow command "([a-z-]+)" was deprecated\.`)

	for _, tc := range tests {
		t.Run(tc.what, func(t *testing.T) {
			s := &Step{
				Exec: &ExecRun{
					Run: &String{
						Value: tc.run,
						Pos:   &Pos{},
					},
				},
			}
			r := NewRuleDeprecatedCommands()
			if err := r.VisitStep(s); err != nil {
				t.Fatal(err)
			}

			errs := r.Errs()
			have := []string{}
			for i, err := range errs {
				m := err.Error()
				ss := re.FindStringSubmatch(m)
				if len(ss) == 0 {
					t.Fatalf("%dth error was unexpected: %q", i, m)
				}
				have = append(have, ss[1])
			}

			if diff := cmp.Diff(have, tc.want); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
