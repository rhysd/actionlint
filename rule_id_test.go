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
		"12345",
	}

	for _, input := range inputs {
		tests := []struct {
			where    string
			job      *Job
			expected string
		}{
			{
				"jobs.<job_id>",
				&Job{
					ID: &String{Value: input, Pos: &Pos{}},
				},
				"invalid job ID",
			},
			{
				"jobs.<job_id>.needs",
				&Job{
					ID: &String{Value: "test", Pos: &Pos{}},
					Needs: []*String{
						{Value: input, Pos: &Pos{}},
					},
				},
				"invalid job ID",
			},
			{
				"steps.*.id",
				&Job{
					Steps: []*Step{
						{ID: &String{Value: input, Pos: &Pos{}}},
					},
				},
				"invalid step ID",
			},
		}

		for _, tc := range tests {
			t.Run(input+"/"+tc.where, func(t *testing.T) {
				r := NewRuleID()
				r.VisitJobPre(tc.job)
				for _, s := range tc.job.Steps {
					r.VisitStep(s)
				}
				errs := r.Errs()
				if len(errs) != 1 {
					t.Fatalf("Wanted exactly one error but got %d errors: %#v", len(errs), errs)
				}
				msg := errs[0].Error()
				if !strings.Contains(msg, tc.expected) {
					t.Errorf("Error message %q should contain %q", msg, tc.expected)
				}
			})
		}
	}
}

func TestCheckValidJobNames(t *testing.T) {
	inputs := []string{
		"foo-bar",
		"foo_bar",
		"foo--bar",
		"foo__bar",
		"_FOO123-",
		"abcdefhijklmnopqrstuvwxyzABCDEFHIJKLMNOPQRSTUVWXYZ",
		"_____",
		"_-_-_",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			job := &Job{
				ID: &String{Value: input, Pos: &Pos{}},
				Needs: []*String{
					{Value: input, Pos: &Pos{}},
				},
				Steps: []*Step{
					{ID: &String{Value: input, Pos: &Pos{}}},
				},
			}
			r := NewRuleID()
			r.VisitJobPre(job)
			for _, s := range job.Steps {
				r.VisitStep(s)
			}
			errs := r.Errs()
			if len(errs) > 0 {
				t.Fatalf("Unexpected error(s): %#v", errs)
			}
		})
	}
}
