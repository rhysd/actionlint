package actionlint

import "testing"

func TestStringIsExpressionAssigned(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"${{...}}", true},
		{" ${{...}} ", true},
		{`${{ foo == '{"a": {"b": "c"}}' }}`, true}, // edge case
		{"", false},
		{"${}", false},
		{"{{}}", false},
		{"${{", false},
		{"}}", false},
		{"${{ ${{ }}", false},
		{"abc ${{...}}", false},
		{"${{...}} abc", false},
		{"${{...}}${{...}}", false},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			s := &String{Value: tc.input}
			have := s.IsExpressionAssigned()
			if tc.want != have {
				t.Fatalf("wanted %v but got %v for input %q", tc.want, have, tc.input)
			}
		})
	}
}

func TestStringContainsExpression(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"${{...}}", true},
		{"foo-${{...}}-bar", true},
		{"${{...}}-${{...}}", true},
		{"${{...", false},
		{"...}}", false},
		{"${{...} }", false},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			s := &String{Value: tc.input}
			if have := s.ContainsExpression(); tc.want != have {
				t.Fatalf("wanted %v but the method returned %v for input %q", tc.want, have, tc.input)
			}
			if have := ContainsExpression(tc.input); tc.want != have {
				t.Fatalf("wanted %v but the function returned %v for input %q", tc.want, have, tc.input)
			}
		})
	}
}
