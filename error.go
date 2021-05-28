package actionlint

import "fmt"

// Error represents an error detected by actionlint rules
type Error struct {
	Message  string
	Filepath string
	Line     int
	Column   int
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s:%d:%d: %s", e.Filepath, e.Line, e.Column, e.Message)
}
