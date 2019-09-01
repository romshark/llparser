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
	Actual   *Token
}

func (err *ErrUnexpectedToken) Error() string {
	actualStr := "<nil>"
	if err.Actual != nil {
		actualStr = string(err.Actual.Src())
	}
	if err.Expected == nil {
		return fmt.Sprintf(
			"unexpected token '%s' at %s",
			actualStr,
			err.At,
		)
	}
	return fmt.Sprintf(
		"unexpected token '%s', expected {%s} at %s",
		actualStr,
		err.Expected.Desig(),
		err.At,
	)
}
