package actionlint

import (
	"fmt"
	"testing"
)

func TestRuleIfCond(t *testing.T) {
	tests := []struct {
		cond  string
		valid bool
	}{
		{"", true},
		{"true", false},
		{"false", false},
		{"true || false", true},
		{"github.ref_name == 'foo'", true},
		{"${{ false }}", true},
		{"${{ false }}\n", false},
		{"${{ false }} ", false},
		{" ${{ false }}", false},
		{"${{ true }} && ${{ true }}", false},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%q at step", tc.cond), func(t *testing.T) {
			var s Step
			if len(tc.cond) > 0 {
				s.If = &String{Value: tc.cond, Pos: &Pos{}}
			}

			r := NewRuleIfCond()
			if err := r.VisitStep(&s); err != nil {
				t.Fatal(err)
			}

			errs := r.Errs()
			if tc.valid && len(errs) > 0 {
				t.Fatalf("wanted no error but have %q for condition %q", errs, tc.cond)
			}
			if !tc.valid && len(errs) != 1 {
				t.Fatalf("wanted one error but have %q for condition %q", errs, tc.cond)
			}
		})
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%q at job", tc.cond), func(t *testing.T) {
			var j Job
			if len(tc.cond) > 0 {
				j.If = &String{Value: tc.cond, Pos: &Pos{}}
				j.Snapshot = &Snapshot{
					ImageName: &String{Value: "test", Pos: &Pos{}},
					If:        &String{Value: tc.cond, Pos: &Pos{}},
				}
			}

			r := NewRuleIfCond()
			if err := r.VisitJobPre(&j); err != nil {
				t.Fatal(err)
			}

			errs := r.Errs()
			if tc.valid && len(errs) > 0 {
				t.Fatalf("wanted no error but have %q for condition %q", errs, tc.cond)
			}
			if !tc.valid && len(errs) != 2 {
				t.Fatalf("wanted one error but have %q for condition %q", errs, tc.cond)
			}
		})
	}
}
