package parser

import "fmt"

// ErrorCode represents an error code
type ErrorCode string

// Error represents a parser error
type Error struct {
	Code ErrorCode
	Msg  string
	At   Cursor
}

func (err *Error) Error() string {
	return fmt.Sprintf("%s: %s at %s", err.Code, err.Msg, err.At.String())
}
