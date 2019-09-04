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

// Term represents a concrete terminal token pattern
type Term FragmentKind

// Container implements the Pattern interface
func (Term) Container() bool { return false }

// TerminalPattern implements the Pattern interface
func (Term) TerminalPattern() Pattern { return nil }

// Desig implements the Pattern interface
func (tm Term) Desig() string {
	return fmt.Sprintf("terminal(%d)", tm)
}

// TermExact represents an exact terminal token pattern
type TermExact struct {
	Kind        FragmentKind
	Expectation []rune
}

// Container implements the Pattern interface
func (TermExact) Container() bool { return false }

// TerminalPattern implements the Pattern interface
func (TermExact) TerminalPattern() Pattern { return nil }

// Desig implements the Pattern interface
func (tm TermExact) Desig() string {
	return "'" + string(tm.Expectation) + "'"
}

// Checked represents an arbitrary terminal token pattern matched by a function
type Checked struct {
	Designation string
	Fn          func([]rune) bool
}

// Container implements the Pattern interface
func (Checked) Container() bool { return false }

// TerminalPattern implements the Pattern interface
func (Checked) TerminalPattern() Pattern { return nil }

// Desig implements the Pattern interface
func (ck Checked) Desig() string { return ck.Designation }

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

// Optional represents an arbitrary optional pattern
type Optional struct{ Pattern }

// Container implements the Pattern interface
func (Optional) Container() bool { return true }

// TerminalPattern implements the Pattern interface
func (opt Optional) TerminalPattern() Pattern { return opt.Pattern }

// Desig implements the Pattern interface
func (opt Optional) Desig() string { return opt.Pattern.Desig() }

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

// ZeroOrMore represents zero or more arbitrary patterns
type ZeroOrMore struct{ Pattern }

// Container implements the Pattern interface
func (ZeroOrMore) Container() bool { return true }

// TerminalPattern implements the Pattern interface
func (zom ZeroOrMore) TerminalPattern() Pattern { return zom.Pattern }

// Desig implements the Pattern interface
func (zom ZeroOrMore) Desig() string {
	return "zero or more " + zom.Pattern.Desig()
}

// OneOrMore represents at least one arbitrary patterns
type OneOrMore struct{ Pattern }

// Container implements the Pattern interface
func (OneOrMore) Container() bool { return true }

// TerminalPattern implements the Pattern interface
func (oom OneOrMore) TerminalPattern() Pattern { return oom.Pattern }

// Desig implements the Pattern interface
func (oom OneOrMore) Desig() string {
	return "zero or more " + oom.Pattern.Desig()
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
