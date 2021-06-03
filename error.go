package actionlint

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"
)

var (
	bold   = color.New(color.Bold)
	green  = color.New(color.FgGreen)
	yellow = color.New(color.FgYellow)
	gray   = color.New(color.FgHiBlack)
)

// Error represents an error detected by actionlint rules
type Error struct {
	Message  string
	Filepath string
	Line     int
	Column   int
	Kind     string
}

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

func (e *Error) PrettyPrint(w io.Writer, source []byte) {
	yellow.Fprint(w, e.Filepath)
	gray.Fprint(w, ":")
	fmt.Fprint(w, e.Line)
	gray.Fprint(w, ":")
	fmt.Fprint(w, e.Column)
	gray.Fprint(w, ": ")
	bold.Fprint(w, e.Message)
	gray.Fprintf(w, " [%s]\n", e.Kind)

	if source == nil {
		return
	}
	line := e.getLine(source)
	if line == "" {
		return
	}

	fmt.Fprintln(w, line)
	green.Fprintln(w, e.getIndicator(line))
}

func (e *Error) getLine(source []byte) string {
	s := bufio.NewScanner(bytes.NewReader(source))
	l := 0
	for s.Scan() {
		l++
		if l == e.Line {
			return s.Text()
		}
	}
	return ""
}

func (e *Error) getIndicator(line string) string {
	start := e.Column - 1 // Column is 1-based

	// Count non-space characters after '^'
	count := 0
	r := strings.NewReader(line[start:])
	for {
		c, s, err := r.ReadRune()
		if err != nil || s == 0 {
			break
		}
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			break
		}
		count++
	}
	if count > 0 {
		count-- // Decrement for place for '^'
	}

	return fmt.Sprintf("%s^%s", strings.Repeat(" ", start), strings.Repeat("~", count))
}
