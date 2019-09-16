package parser

import (
	"fmt"
	"strings"
)

// Pattern represents an abstract pattern
type Pattern interface {
	// Container returns true for container patterns
	Container() bool

	// TerminalPattern returns the terminal pattern for composite patterns
	// and nil for non-terminal patterns
	TerminalPattern() Pattern

	// Desig returns the textual designation of the pattern
	Desig() string
}

// Exact represents an exact terminal token pattern
type Exact struct {
	Kind        FragmentKind
	Expectation []rune
}

// Container implements the Pattern interface
func (Exact) Container() bool { return false }

// TerminalPattern implements the Pattern interface
func (Exact) TerminalPattern() Pattern { return nil }

// Desig implements the Pattern interface
func (tm Exact) Desig() string {
	return "'" + string(tm.Expectation) + "'"
}

// Lexed represents a lexed pattern
type Lexed struct {
	Kind        FragmentKind
	Designation string
	Fn          func(Cursor) uint
}

// Container implements the Pattern interface
func (Lexed) Container() bool { return false }

// TerminalPattern implements the Pattern interface
func (Lexed) TerminalPattern() Pattern { return nil }

// Desig implements the Pattern interface
func (ck Lexed) Desig() string { return ck.Designation }

// Sequence represents an exact sequence of arbitrary patterns
type Sequence []Pattern

// Container implements the Pattern interface
func (Sequence) Container() bool { return true }

// TerminalPattern implements the Pattern interface
func (Sequence) TerminalPattern() Pattern { return nil }

// Desig implements the Pattern interface
func (seq Sequence) Desig() string {
	str := make([]string, len(seq))
	for ix, el := range seq {
		str[ix] = el.Desig()
	}
	return "{" + strings.Join(str, ", ") + "}"
}

// Repeated represents at least one arbitrary patterns
type Repeated struct {
	Pattern Pattern
	Min     uint
	Max     uint
}

// Container implements the Pattern interface
func (Repeated) Container() bool { return true }

// TerminalPattern implements the Pattern interface
func (oom Repeated) TerminalPattern() Pattern { return oom.Pattern }

// Desig implements the Pattern interface
func (oom Repeated) Desig() string {
	switch {
	case oom.Max < 1:
		return fmt.Sprintf(
			"%d+ repetitions of %s",
			oom.Min,
			oom.Pattern.Desig(),
		)
	case oom.Max == 0 && oom.Min == 0:
		return fmt.Sprintf(
			"0+ repetitions of %s",
			oom.Pattern.Desig(),
		)
	case oom.Max == 0 && oom.Min == 1:
		return fmt.Sprintf(
			"optional %s",
			oom.Pattern.Desig(),
		)
	case oom.Max == oom.Min:
		return fmt.Sprintf(
			"exactly %d repetitions of %s",
			oom.Min,
			oom.Pattern.Desig(),
		)
	}
	return fmt.Sprintf(
		"%d-%d repetitions of %s",
		oom.Min,
		oom.Max,
		oom.Pattern.Desig(),
	)
}

// Either represents either of the arbitrary patterns
type Either []Pattern

// Container implements the Pattern interface
func (Either) Container() bool { return true }

// TerminalPattern implements the Pattern interface
func (Either) TerminalPattern() Pattern { return nil }

// Desig implements the Pattern interface
func (eth Either) Desig() string {
	str := make([]string, len(eth))
	for ix, el := range eth {
		str[ix] = el.Desig()
	}
	return "either of [" + strings.Join(str, ", ") + "]"
}
