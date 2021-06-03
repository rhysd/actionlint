package actionlint

import (
	"fmt"
	"io"

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

func (e *Error) PrettyPrint(w io.Writer, location bool) {
	green.Fprint(w, e.Filepath)
	fmt.Fprint(w, ":")
	yellow.Fprint(w, e.Line)
	fmt.Fprint(w, ":")
	yellow.Fprint(w, e.Column)
	fmt.Fprint(w, ": ")
	bold.Fprint(w, e.Message)
	gray.Fprintf(w, " [%s]\n", e.Kind)
}
