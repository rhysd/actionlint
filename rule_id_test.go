package actionlint

import (
	"strings"
	"testing"
)

func TestCheckInvalidJobNames(t *testing.T) {
	// Note: Empty string is already reported as error by parser
	inputs := []string{
		"-foo",
		"v1.2.3",
		"hello!",
		"じょぶ",
	}

	for _, input := range inputs {
		tests := []struct {
			where string
			job   *Job
		}{
			{
				"jobs.<job_id>",
				&Job{
					ID: &String{Value: input, Pos: &Pos{}},
				},
			},
			{
				"jobs.<job_id>.needs",
				&Job{
					ID: &String{Value: "test", Pos: &Pos{}},
					Needs: []*String{
						{Value: input, Pos: &Pos{}},
					},
				},
			},
		}

		for _, tc := range tests {
			t.Run(input+"/"+tc.where, func(t *testing.T) {
				r := NewRuleID()
				r.VisitJobPre(tc.job)
				errs := r.Errs()
				if len(errs) != 1 {
					t.Fatalf("Wanted exactly one error but got %d errors: %#v", len(errs), errs)
				}
				msg := errs[0].Error()
				if !strings.Contains(msg, "invalid job ID") {
					t.Errorf("Unexpected error message %q", msg)
				}
			})
		}
	}
}

func TestCheckValidJobNames(t *testing.T) {
	inputs := []string{
		"foo-bar",
		"foo_bar",
		"_FOO123-",
		"1_2_3-foo",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			job := &Job{
				ID: &String{Value: input, Pos: &Pos{}},
			}
			r := NewRuleID()
			r.VisitJobPre(job)
			errs := r.Errs()
			if len(errs) > 0 {
				t.Fatalf("Unexpected error(s): %#v", errs)
			}
		})
	}
}
