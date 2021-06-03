package actionlint

import (
	"bytes"
	"testing"

	"github.com/fatih/color"
)

func init() {
	color.NoColor = true
}

func TestErrorErrorAt(t *testing.T) {
	m := "message"
	k := "kind"
	err := errorAt(&Pos{1, 2}, k, m)
	if err.Message != m {
		t.Errorf("wanted %q but got %q", m, err.Message)
	}
	if err.Filepath != "" {
		t.Errorf("wanted %q to be empty", err.Filepath)
	}
	if err.Line != 1 {
		t.Errorf("wanted line 1 but got %q", err.Line)
	}
	if err.Column != 2 {
		t.Errorf("wanted col 2 but got %q", err.Column)
	}
	if err.Kind != k {
		t.Errorf("wanted %q but got %q", k, err.Kind)
	}
}

func TestErrorErrorfAt(t *testing.T) {
	m := "this is message"
	k := "kind"
	err := errorfAt(&Pos{1, 2}, k, "%s is %s", "this", "message")
	if err.Message != m {
		t.Errorf("wanted %q but got %q", m, err.Message)
	}
	if err.Filepath != "" {
		t.Errorf("wanted %q to be empty", err.Filepath)
	}
	if err.Line != 1 {
		t.Errorf("wanted line 1 but got %q", err.Line)
	}
	if err.Column != 2 {
		t.Errorf("wanted col 2 but got %q", err.Column)
	}
	if err.Kind != k {
		t.Errorf("wanted %q but got %q", k, err.Kind)
	}
}

func TestErrorPrettyPrint(t *testing.T) {
	testCases := []struct {
		message  string
		line     int
		column   int
		kind     string
		expected string
		source   string
	}{
		{
			message:  "simple message",
			line:     1,
			column:   1,
			expected: "filename.txt:1:1: simple message [kind]",
		},
		{
			message: "simple message with source",
			line:    1,
			column:  1,
			source:  "this is source",
			expected: `filename.txt:1:1: simple message with source [kind]
1| this is source
 | ^~~~`,
		},
		{
			message: "error at middle of source",
			line:    1,
			column:  6,
			source:  "this is source",
			expected: `filename.txt:1:6: error at middle of source [kind]
1| this is source
 |      ^~`,
		},
		{
			message: "error at end of source",
			line:    1,
			column:  15,
			source:  "this is source",
			expected: `filename.txt:1:15: error at end of source [kind]
1| this is source
 |               ^`,
		},
		{
			message: "simple message with multi-line source",
			line:    3,
			column:  3,
			source:  "this\nis\nsource",
			expected: `filename.txt:3:3: simple message with multi-line source [kind]
3| source
 |   ^~~~`,
		},
		{
			message: "error at end of multi-line source",
			line:    3,
			column:  7,
			source:  "this\nis\nsource",
			expected: `filename.txt:3:7: error at end of multi-line source [kind]
3| source
 |       ^`,
		},
		{
			message: "error at newline of multi-line source",
			line:    2,
			column:  3,
			source:  "this\nis\nsource",
			expected: `filename.txt:2:3: error at newline of multi-line source [kind]
2| is
 |   ^`,
		},
		{
			message: "error at blank line of multi-line source",
			line:    2,
			column:  1,
			source:  "this\n\nsource",
			expected: `filename.txt:2:1: error at blank line of multi-line source [kind]
2| 
 | ^`,
		},
		{
			message: "error at line more than 10",
			line:    11,
			column:  2,
			source:  "\n\n\n\n\n\n\n\n\n\nfooo",
			expected: `filename.txt:11:2: error at line more than 10 [kind]
11| fooo
  |  ^~~`,
		},
		{
			message:  "error at out of source",
			line:     2,
			column:   6,
			source:   "aaa\nbbb\nccc",
			expected: "filename.txt:2:6: error at out of source [kind]",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.message, func(t *testing.T) {
			err := errorAt(&Pos{tc.line, tc.column}, "kind", tc.message)
			err.Filepath = "filename.txt"

			var buf bytes.Buffer
			err.PrettyPrint(&buf, []byte(tc.source))

			out := buf.String()
			want := tc.expected + "\n"
			if out != want {
				t.Fatalf("wanted:\n%q\n\nhave:\n%q", want, out)
			}
		})
	}
}
