package actionlint

import (
	"strings"
	"testing"
)

func TestInvalidGlobPatternMessage(t *testing.T) {
	want := "42: this is message"
	err := &InvalidGlobPattern{"this is message", 42}
	have := err.Error()
	if want != have {
		t.Fatalf("want %q but have %q", want, have)
	}
}

func TestValidateGlobOK(t *testing.T) {
	testCases := []string{
		"a",
		"abc/def",
		"[ab]",
		"[a-z]",
		"[a-zA-Z_]",
		"*",
		"**",
		"a?",
		"a+",
		`\+`,
		`\\`,
		`\++\\?`,
		"[a-z]+",
		"[a-z]?",
		"*.*.*-**",
		"!a",
		"a!",
		// examples in official documents
		"feature/*",
		"feature/**",
		"main",
		"releases/mona-the-octcat",
		"*",
		"**",
		"*feature",
		"v2*",
		"v[12].[0-9]+.[0-9]+",
	}

	for _, input := range testCases {
		t.Run(input, func(t *testing.T) {
			if errs := ValidateRefGlob(input); len(errs) > 0 {
				t.Errorf("ref glob %q caused errors: %#v", input, errs)
			}
			if errs := ValidatePathGlob(input); len(errs) > 0 {
				t.Errorf("path glob %q caused errors: %#v", input, errs)
			}
		})
	}
}

func TestValidateGlobPathOnlyOK(t *testing.T) {
	testCases := []string{
		".",
		"/foo",
		"/foo/",
		"/foo/bar",
		`\[\?\*\+\\`,
		"foo bar",
		"~/foo",
		"foo:bar",
		"foo^bar",
		// examples in official document
		"*.jsx?",
		"*.js",
		"**.js",
		"docs/*",
		"docs/**",
		"docs/**/*.md",
		"**/docs/**",
		"**/README.md",
		"**/*src/**",
		"*/*-post.md",
		"**/migrate-*.sql",
		"!README.md",
	}

	for _, input := range testCases {
		t.Run(input, func(t *testing.T) {
			if errs := ValidatePathGlob(input); len(errs) > 0 {
				t.Fatalf("path glob %q caused errors: %#v", input, errs)
			}
		})
	}
}

func TestValidateGlobSyntaxError(t *testing.T) {
	testCases := []struct {
		what        string
		input       string
		expected    string
		expectedAll []string
	}{
		{
			what:     "empty",
			input:    "",
			expected: "glob pattern cannot be empty",
		},
		{
			what:     "nothing after negate",
			input:    "!",
			expected: "at least one character must follow !",
		},
		{
			what:     "? as first character",
			input:    "?",
			expected: "the preceding character must not be special character",
		},
		{
			what:     "? following special character",
			input:    "*?",
			expected: "the preceding character must not be special character",
		},
		{
			what:     "? following ?",
			input:    "a??",
			expected: "the preceding character must not be special character",
		},
		{
			what:     "+ as first character",
			input:    "+",
			expected: "the preceding character must not be special character",
		},
		{
			what:     "+ following special character",
			input:    "*+",
			expected: "the preceding character must not be special character",
		},
		{
			what:     "+ following +",
			input:    "a++",
			expected: "the preceding character must not be special character",
		},
		{
			what:     "newline in pattern",
			input:    "\n",
			expected: "newline cannot be contained",
		},
		{
			what:     "empty match",
			input:    "[]",
			expected: "character match must not be empty",
		},
		{
			what:     "missing [ with empty match",
			input:    "[",
			expected: "missing ]",
		},
		{
			what:     "missing [",
			input:    "[a",
			expected: "missing ]",
		},
		{
			what:     "missing [ with range",
			input:    "[a-c",
			expected: "missing ]",
		},
		{
			what:     "missing end of range",
			input:    "[a-]",
			expected: "end of range is missing",
		},
		{
			what:     "EOF inside range",
			input:    "[a-",
			expected: "missing ]",
		},
		{
			what:     "invalid range",
			input:    "[b-a]",
			expected: "start of range 'b' (98) is larger than end of range 'a' (97)",
		},
		{
			what:  "multiple errors",
			input: "+?[][a-]*+\n[b",
			expectedAll: []string{
				"the preceding character must not be special character",
				"the preceding character must not be special character",
				"character match must not be empty",
				"end of range is missing",
				"the preceding character must not be special character",
				"newline cannot be contained",
				"missing ]",
			},
		},
	}

	for _, kind := range []string{"ref", "path"} {
		for _, tc := range testCases {
			t.Run(kind+"/"+tc.what, func(t *testing.T) {
				var errs []InvalidGlobPattern
				if kind == "ref" {
					errs = ValidateRefGlob(tc.input)
				} else {
					errs = ValidatePathGlob(tc.input)
				}

				expected := tc.expectedAll
				if len(expected) == 0 {
					expected = []string{tc.expected}
				}

				if len(errs) != len(expected) {
					t.Fatalf("wanted %d errors from %s glob %q but got %d errors", len(expected), kind, tc.input, len(errs))
				}

				for i := range errs {
					err := errs[i]
					want, have := expected[i], err.Message
					if !strings.Contains(have, want) {
						t.Errorf("%dth error message at col:%d from %s glob %q does not contain expected string:\n  want: %s\n  have: %s", i+1, err.Column, kind, tc.input, want, have)
					}
				}
			})
		}
	}
}

func TestValidateGitRefNameInvalidCharacter(t *testing.T) {
	testCases := []struct {
		what        string
		input       string
		expected    string
		expectedAll []string
	}{
		{
			what:     "start with /",
			input:    "/foo",
			expected: "ref name must not start with /",
		},
		{
			what:     "end with /",
			input:    "foo/",
			expected: "ref name must not end with / and .",
		},
		{
			what:     "end with .",
			input:    "foo.",
			expected: "ref name must not end with / and .",
		},
		{
			what:  "escaped special chars",
			input: `\[\?\*`,
			expectedAll: []string{
				"ref name cannot contain spaces, ~, ^, :, [, ?, *",
				"ref name cannot contain spaces, ~, ^, :, [, ?, *",
				"ref name cannot contain spaces, ~, ^, :, [, ?, *",
			},
		},
		{
			what:     "escaped non-special character",
			input:    `\d`,
			expected: "only special characters [, ?, +, *, \\ can be escaped with \\",
		},
		{
			what: "prohibited characters for ref names",
			input: " 	~^:",
			expectedAll: []string{
				"ref name cannot contain spaces, ~, ^, :, [, ?, *",
				"ref name cannot contain spaces, ~, ^, :, [, ?, *",
				"ref name cannot contain spaces, ~, ^, :, [, ?, *",
				"ref name cannot contain spaces, ~, ^, :, [, ?, *",
				"ref name cannot contain spaces, ~, ^, :, [, ?, *",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.what, func(t *testing.T) {
			errs := ValidateRefGlob(tc.input)

			expected := tc.expectedAll
			if len(expected) == 0 {
				expected = []string{tc.expected}
			}

			if len(errs) != len(expected) {
				t.Fatalf("wanted %d errors from %q but got %d errors", len(expected), tc.input, len(errs))
			}

			for i := range errs {
				err := errs[i]
				want, have := expected[i], err.Message
				if !strings.Contains(have, want) {
					t.Errorf("%dth error message at col:%d from %q does not contain expected string:\n  want: %s\n  have: %s", i+1, err.Column, tc.input, want, have)
				}
			}
		})
	}
}

// TODO: Test column
