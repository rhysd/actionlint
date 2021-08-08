package actionlint

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/fatih/color"
	"github.com/mattn/go-runewidth"
)

var (
	bold   = color.New(color.Bold)
	green  = color.New(color.FgGreen)
	yellow = color.New(color.FgYellow)
	gray   = color.New(color.FgHiBlack)
)

// Error represents an error detected by actionlint rules
type Error struct {
	// Message is an error message.
	Message string
	// Filepath is a file path where the error occurred.
	Filepath string
	// Line is a line number where the error occurred. This value is 1-based.
	Line int
	// Column is a column number where the error occurred. This value is 1-based.
	Column int
	// Kind is a string to represent kind of the error. Usually rule name which found the error.
	Kind string
}

type errorTemplateFields struct {
	Message  string
	Filepath string
	Line     int
	Column   int
	Kind     string
	Snippet  string
}

// Error returns summary of the error as string.
func (e *Error) Error() string {
	return fmt.Sprintf("%s:%d:%d: %s [%s]", e.Filepath, e.Line, e.Column, e.Message, e.Kind)
}

func errorAt(pos *Pos, kind string, msg string) *Error {
	return &Error{
		Message: msg,
		Line:    pos.Line,
		Column:  pos.Col,
		Kind:    kind,
	}
}

func errorfAt(pos *Pos, kind string, format string, args ...interface{}) *Error {
	return &Error{
		Message: fmt.Sprintf(format, args...),
		Line:    pos.Line,
		Column:  pos.Col,
		Kind:    kind,
	}
}

// PrintTemplate prints this error with given Go template. Available fields are:
// - Message: Error message body.
// - Snippet: Code snippet and indicator to indicate where the error occurred.
// - Filepath: Canonical relative file path. This is empty when input was read from stdin.
// - Line: Line number of error position.
// - Column: Column number of error position.
// - Kind: Rule name the error belongs to.
func (e *Error) PrintTemplate(t *template.Template, w io.Writer, source []byte) error {
	var snippet string
	if len(source) > 0 && e.Line > 0 {
		if l, ok := e.getLine(source); ok {
			snippet = l
			if i := e.getIndicator(l); i != "" {
				snippet += "\n" + i
			}
		}
	}

	fields := errorTemplateFields{
		Message:  e.Message,
		Filepath: e.Filepath,
		Line:     e.Line,
		Column:   e.Column,
		Kind:     e.Kind,
		Snippet:  snippet,
	}

	return t.Execute(w, &fields)
}

// PrettyPrint prints the error with user-friendly way. It prints file name, source position, error
// message with colorful output and source snippet with indicator. When nil is set to source, no
// source snippet is not printed. To disable colorful output, set true to fatih/color.NoColor.
func (e *Error) PrettyPrint(w io.Writer, source []byte) {
	yellow.Fprint(w, e.Filepath)
	gray.Fprint(w, ":")
	fmt.Fprint(w, e.Line)
	gray.Fprint(w, ":")
	fmt.Fprint(w, e.Column)
	gray.Fprint(w, ": ")
	bold.Fprint(w, e.Message)
	gray.Fprintf(w, " [%s]\n", e.Kind)

	if len(source) == 0 || e.Line <= 0 {
		return
	}
	line, ok := e.getLine(source)
	if !ok || len(line) < e.Column-1 {
		return
	}

	lnum := fmt.Sprintf("%d | ", e.Line)
	indent := strings.Repeat(" ", len(lnum)-2)
	gray.Fprintf(w, "%s|\n", indent)
	gray.Fprint(w, lnum)
	fmt.Fprintln(w, line)
	gray.Fprintf(w, "%s| ", indent)
	green.Fprintln(w, e.getIndicator(line))
}

func (e *Error) getLine(source []byte) (string, bool) {
	s := bufio.NewScanner(bytes.NewReader(source))
	l := 0
	for s.Scan() {
		l++
		if l == e.Line {
			return s.Text(), true
		}
	}
	return "", false
}

func (e *Error) getIndicator(line string) string {
	if e.Column <= 0 {
		return ""
	}

	start := e.Column - 1 // Column is 1-based

	// Count width of non-space characters after '^' for underline
	uw := 0
	r := strings.NewReader(line[start:])
	for {
		c, s, err := r.ReadRune()
		if err != nil || s == 0 || c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			break
		}
		uw += runewidth.RuneWidth(c)
	}
	if uw > 0 {
		uw-- // Decrement for place for '^'
	}

	// Count width of spaces before '^'
	sw := runewidth.StringWidth(line[:start])
	return fmt.Sprintf("%s^%s", strings.Repeat(" ", sw), strings.Repeat("~", uw))
}

// ByErrorPosition is type for sort.Interface. It sorts errors slice in line and column order.
type ByErrorPosition []*Error

func (by ByErrorPosition) Len() int {
	return len(by)
}

func (by ByErrorPosition) Less(i, j int) bool {
	if by[i].Line == by[j].Line {
		return by[i].Column < by[j].Column
	}
	return by[i].Line < by[j].Line
}

func (by ByErrorPosition) Swap(i, j int) {
	by[i], by[j] = by[j], by[i]
}
