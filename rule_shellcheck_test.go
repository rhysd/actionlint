package actionlint

import (
	"strconv"
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
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			have := sanitizeExpressionsInScript(tc.input)
			if tc.want != have {
				t.Fatalf("sanitized result is unexpected.\nwant: %q\nhave: %q", tc.want, have)
			}
		})
	}
}
