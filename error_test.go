package actionlint

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/google/go-cmp/cmp"
)

func init() {
	color.NoColor = true
}

type testErrorWriter struct{}

func (w testErrorWriter) Write(b []byte) (int, error) {
	return 0, errors.New("dummy write error")
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
  |
1 | this is source
  | ^~~~`,
		},
		{
			message: "error at middle of source",
			line:    1,
			column:  6,
			source:  "this is source",
			expected: `filename.txt:1:6: error at middle of source [kind]
  |
1 | this is source
  |      ^~`,
		},
		{
			message: "error at end of source",
			line:    1,
			column:  15,
			source:  "this is source",
			expected: `filename.txt:1:15: error at end of source [kind]
  |
1 | this is source
  |               ^`,
		},
		{
			message: "error at one character word in source",
			line:    1,
			column:  5,
			source:  "foo . bar",
			expected: `filename.txt:1:5: error at one character word in source [kind]
  |
1 | foo . bar
  |     ^`,
		},
		{
			message: "error at space in source",
			line:    1,
			column:  4,
			source:  "foo bar",
			expected: `filename.txt:1:4: error at space in source [kind]
  |
1 | foo bar
  |    ^`,
		},
		{
			message: "simple message with multi-line source",
			line:    3,
			column:  3,
			source:  "this\nis\nsource",
			expected: `filename.txt:3:3: simple message with multi-line source [kind]
  |
3 | source
  |   ^~~~`,
		},
		{
			message: "error at end of multi-line source",
			line:    3,
			column:  7,
			source:  "this\nis\nsource",
			expected: `filename.txt:3:7: error at end of multi-line source [kind]
  |
3 | source
  |       ^`,
		},
		{
			message: "error at newline of multi-line source",
			line:    2,
			column:  3,
			source:  "this\nis\nsource",
			expected: `filename.txt:2:3: error at newline of multi-line source [kind]
  |
2 | is
  |   ^`,
		},
		{
			message: "error at blank line of multi-line source",
			line:    2,
			column:  1,
			source:  "this\n\nsource",
			expected: `filename.txt:2:1: error at blank line of multi-line source [kind]
  |
2 | 
  | ^`,
		},
		{
			message: "error at line more than 10",
			line:    11,
			column:  2,
			source:  "\n\n\n\n\n\n\n\n\n\nfooo",
			expected: `filename.txt:11:2: error at line more than 10 [kind]
   |
11 | fooo
   |  ^~~`,
		},
		{
			message:  "error at out of source",
			line:     2,
			column:   6,
			source:   "aaa\nbbb\nccc",
			expected: "filename.txt:2:6: error at out of source [kind]",
		},
		{
			message: "error at zero column",
			line:    1,
			column:  0,
			source:  "this is source",
			expected: `filename.txt:1:0: error at zero column [kind]
  |
1 | this is source
  | `,
		},
		{
			message:  "error at zero line and zero column",
			line:     0,
			column:   0,
			source:   "this is source",
			expected: `filename.txt:0:0: error at zero line and zero column [kind]`,
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

func TestErrorSortErrorsByPosition(t *testing.T) {
	testCases := [][]struct {
		line int
		col  int
	}{
		{},
		{
			{1, 2},
		},
		{
			{1, 2},
			{4, 1},
			{3, 20},
			{1, 1},
		},
		{
			{1, 1},
			{1, 1},
			{1, 1},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			errs := make([]*Error, 0, len(tc))
			for _, p := range tc {
				errs = append(errs, &Error{Line: p.line, Column: p.col})
			}

			sort.Stable(ByErrorPosition(errs))

			for i := 0; i < len(errs)-1; i++ {
				l, r := errs[i], errs[i+1]
				sorted := l.Line <= r.Line
				if l.Line == r.Line {
					sorted = l.Column <= r.Column
				}
				if !sorted {
					t.Fatalf("errs[%d] and errs[%d] are not sorted: %s", i, i+1, errs)
				}
			}
		})
	}
}

func TestErrorSortErrorsByFile(t *testing.T) {
	errs := []*Error{
		{
			Filepath: "path/to/C.txt",
			Line:     1,
			Column:   2,
		},
		{
			Filepath: "path/to/A.txt",
			Line:     2,
			Column:   4,
		},
		{
			Filepath: "path/to/B.txt",
			Line:     3,
			Column:   6,
		},
	}

	sort.Stable(ByErrorPosition(errs))

	for i, want := range []string{
		"path/to/A.txt",
		"path/to/B.txt",
		"path/to/C.txt",
	} {
		if have := errs[i].Filepath; have != want {
			t.Errorf("Errors were not sorted correctly. expected %q for errs[%d] but got %q: %v", want, i, have, errs)
		}
	}
}

func TestErrorGetTemplateFieldsOK(t *testing.T) {
	testCases := []struct {
		message string
		column  int
		endCol  int
		source  string
		snippet string
	}{
		{
			message: "simple message with source",
			column:  1,
			endCol:  4,
			source:  "this is source",
			snippet: "this is source\n^~~~",
		},
		{
			message: "simple message",
			column:  1,
			endCol:  1,
			snippet: "",
		},
		{
			message: "error at zero column",
			column:  0,
			endCol:  0,
			source:  "this is source",
			snippet: "this is source",
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			err := errorAt(&Pos{1, tc.column}, "kind", tc.message)
			err.Filepath = "filename.txt"
			f := err.GetTemplateFields([]byte(tc.source))
			if f.Message != tc.message {
				t.Fatalf("wanted %q but have %q", tc.message, f.Message)
			}
			if f.Line != err.Line {
				t.Fatalf("wanted %d but have %d", err.Line, f.Line)
			}
			if f.Column != tc.column {
				t.Fatalf("wanted %d but have %d", tc.column, f.Column)
			}
			if f.Filepath != err.Filepath {
				t.Fatalf("wanted %q but have %q", err.Filepath, f.Filepath)
			}
			if f.Kind != err.Kind {
				t.Fatalf("wanted %q but have %q", err.Kind, f.Kind)
			}
			if f.Snippet != tc.snippet {
				t.Fatalf("wanted %q but have %q", tc.snippet, f.Snippet)
			}
			if f.EndColumn != tc.endCol {
				t.Fatalf("wanted %d but have %d", tc.endCol, f.EndColumn)
			}
		})
	}
}

// Regression for #128
func TestErrorGetTemplateFieldsColumnIsOutOfBounds(t *testing.T) {
	err := errorAt(&Pos{1, 9999}, "kind", "this is message")
	err.Filepath = "filename.yaml"
	f := err.GetTemplateFields([]byte("this is source"))
	if strings.Contains(f.Snippet, "\n") {
		t.Fatalf("snippet should contain indicator but it has: %q", f.Snippet)
	}
}

var testErrorTemplateFields = []*ErrorTemplateFields{
	{
		Message:   "message 1",
		Filepath:  "file1",
		Line:      1,
		Column:    2,
		EndColumn: 3,
		Snippet:   "snippet 1",
		Kind:      "kind1",
	},
	{
		Message:   "message 2",
		Filepath:  "file2",
		Line:      3,
		Column:    4,
		EndColumn: 5,
		Snippet:   "snippet 2",
		Kind:      "kind2",
	},
}

func TestErrorPrintFormattedWithTemplateFields(t *testing.T) {
	testCases := []struct {
		temp string
		want string
	}{
		{
			temp: "{{(index . 0).Message}} {{(index . 1).Message}}",
			want: "message 1 message 2",
		},
		{
			temp: "{{range $ = .}}({{$.Line}}, {{$.Column}}){{end}}",
			want: "(1, 2)(3, 4)",
		},
		{
			temp: "{{range $ = .}}[{{$.Column}}..{{$.EndColumn}}]{{end}}",
			want: "[2..3][4..5]",
		},
		{
			temp: "{{range $ = .}}{{json $.Snippet}}{{end}}",
			want: "\"snippet 1\"\n\"snippet 2\"\n",
		},
		{
			temp: "{{range $ = .}}{{replace $.Kind \"kind\" \"king\"}}{{end}}",
			want: "king1king2",
		},
		// Rules are not registerred so index is always 0
		{
			temp: "{{range $ = .}}{{kindIndex $.Kind}}\n{{end}}",
			want: "0\n0\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.temp, func(t *testing.T) {
			f, err := NewErrorFormatter(tc.temp)
			if err != nil {
				t.Fatal(err)
			}
			var b strings.Builder
			if err := f.Print(&b, testErrorTemplateFields); err != nil {
				t.Fatal(err)
			}
			have := b.String()
			if tc.want != have {
				t.Fatalf("wanted %q but have %q", tc.want, have)
			}
		})
	}
}

func TestErrorPrintFormattedErrors(t *testing.T) {
	errs := []*Error{
		errorAt(&Pos{1, 1}, "kind1", "error1"),
		errorAt(&Pos{1, 0}, "kind2", "error2"),
	}

	f, err := NewErrorFormatter("{{range $ = .}}({{$.Message | printf \"%q\"}},{{$.Snippet | printf \"%q\"}}){{end}}")
	if err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	if err := f.PrintErrors(&b, errs, []byte("this is source")); err != nil {
		t.Fatal(err)
	}
	have := b.String()
	want := `("error1","this is source\n^~~~")("error2","this is source")`
	if want != have {
		t.Fatalf("wanted %q but have %q", want, have)
	}
}

func TestErrorPrintSerializedIntoJSON(t *testing.T) {
	f, err := NewErrorFormatter("{{json .}}")
	if err != nil {
		t.Fatal(err)
	}
	var b bytes.Buffer
	if err := f.Print(&b, testErrorTemplateFields); err != nil {
		t.Fatal(err)
	}

	decoded := []*ErrorTemplateFields{}
	if err := json.Unmarshal(b.Bytes(), &decoded); err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(testErrorTemplateFields, decoded) {
		t.Fatal(cmp.Diff(testErrorTemplateFields, decoded))
	}
}

func TestErrorPrintKindIndex(t *testing.T) {
	errs := []*Error{
		errorAt(&Pos{}, "rule1", "error 1"),
		errorAt(&Pos{}, "rule2", "error 2"),
		errorAt(&Pos{}, "syntax-check", "error 3"),
	}

	f, err := NewErrorFormatter("{{range $ = .}}{{kindIndex $.Kind}} {{end}}")
	if err != nil {
		t.Fatal(err)
	}

	f.RegisterRule(&RuleBase{
		name: "rule1",
		desc: "description for rule1",
	})
	f.RegisterRule(&RuleBase{
		name: "rule2",
		desc: "description for rule2",
	})

	var b bytes.Buffer
	if err := f.PrintErrors(&b, errs, []byte("dummy source")); err != nil {
		t.Fatal(err)
	}

	want := "1 2 0 "
	have := b.String()
	if want != have {
		t.Fatalf("wanted %q but got %q", want, have)
	}
}

func TestErrorPrintAllKinds(t *testing.T) {
	f, err := NewErrorFormatter("{{range $k,$v := allKinds}}{{$k}}: {{$v.Index}}: {{$v.Description}}\n{{end}}")
	if err != nil {
		t.Fatal(err)
	}

	f.RegisterRule(&RuleBase{
		name: "rule1",
		desc: "description for rule1",
	})
	f.RegisterRule(&RuleBase{
		name: "rule2",
		desc: "description for rule2",
	})

	var b bytes.Buffer
	if err := f.PrintErrors(&b, []*Error{}, []byte("dummy source")); err != nil {
		t.Fatal(err)
	}
	output := b.String()

	for _, want := range []string{
		"syntax-check: 0: Checks for GitHub Actions workflow syntax\n",
		"rule1: 1: description for rule1\n",
		"rule2: 2: description for rule2\n",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("%q is not included in `allKinds` output: %q", want, output)
		}
	}
}

func TestErrorNewErrorFormatterError(t *testing.T) {
	testCases := []struct {
		temp string
		want string
	}{
		{"hello", "template to format error messages must contain at least one {{ }} placeholder"},
		{"{{xxx", "template \"{{xxx\" to format error messages could not be parsed"},
	}

	for _, tc := range testCases {
		t.Run(tc.temp, func(t *testing.T) {
			_, err := NewErrorFormatter(tc.temp)
			if err == nil {
				t.Fatal("error did not occur")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("%q is not contained in error message %q", tc.want, err.Error())
			}
		})
	}
}

func TestErrorFormatterPrintError(t *testing.T) {
	testCases := []struct {
		out  io.Writer
		temp string
		want string
	}{
		{io.Discard, "{{.Foo}}", "can't evaluate field Foo in type"},
		{testErrorWriter{}, "{{(index . 0).Message}}", "dummy write error"},
	}

	for _, tc := range testCases {
		t.Run(tc.temp, func(t *testing.T) {
			f, err := NewErrorFormatter(tc.temp)
			if err != nil {
				t.Fatal(err)
			}
			err = f.Print(tc.out, testErrorTemplateFields)
			if err == nil {
				t.Fatal("error did not occur")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("%q is not contained in error message %q", tc.want, err.Error())
			}
		})
	}
}

func TestErrorFormatterPrintJSONEncodeError(t *testing.T) {
	f, err := NewErrorFormatter("{{json .}}")
	if err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	err = f.temp.Execute(&b, math.NaN())
	if err == nil {
		t.Fatal("error did not occur", b.String())
	}
	want := "could not encode template value into JSON"
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("%q is not contained in error message %q", want, err.Error())
	}
}

func TestErrorString(t *testing.T) {
	err := &Error{
		Message: "this is message",
		Line:    1,
		Column:  2,
		Kind:    "test",
	}
	want := err.Error()
	have := err.String()
	if want != have {
		t.Fatalf("wanted %q but have %q", want, have)
	}
}
