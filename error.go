package parser

import "fmt"

// Err represents a generic parser error
type Err struct {
	Err error
	At  Cursor
}

func (err *Err) Error() string {
	return fmt.Sprintf("%s at %s", err.Err, err.At.String())
}

// ErrUnexpectedToken represents a parser error
type ErrUnexpectedToken struct {
	At       Cursor
	Expected Pattern
}

func (err *ErrUnexpectedToken) Error() string {
	if err.Expected == nil {
		return fmt.Sprintf(
			"unexpected token at %s",
			err.At,
		)
	}
	return fmt.Sprintf(
		"unexpected token, expected {%s} at %s",
		err.Expected.Desig(),
		err.At,
	)
}

type errEOF struct{}

func (err errEOF) Error() string { return "eof" }
