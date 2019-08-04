package parser

// Pattern represents an abstract pattern
type Pattern interface {
	// TerminalPattern returns the terminal pattern for composite patterns
	// and nil for non-terminal patterns
	TerminalPattern() Pattern
}

// Term represents a concrete terminal token pattern
type Term FragmentKind

// TerminalPattern implements the Pattern interface
func (Term) TerminalPattern() Pattern { return nil }

// TermExact represents an exact terminal token pattern
type TermExact string

// TerminalPattern implements the Pattern interface
func (TermExact) TerminalPattern() Pattern { return nil }

// Checked represents an arbitrary terminal token pattern matched by a function
type Checked func(string) bool

// TerminalPattern implements the Pattern interface
func (Checked) TerminalPattern() Pattern { return nil }

// Optional represents an arbitrary optional pattern
type Optional struct{ Pattern }

// TerminalPattern implements the Pattern interface
func (opt Optional) TerminalPattern() Pattern { return Pattern(opt) }

// Sequence represents an exact sequence arbitrary patterns
type Sequence []Pattern

// TerminalPattern implements the Pattern interface
func (Sequence) TerminalPattern() Pattern { return nil }

// ZeroOrMore represents zero or more arbitrary patterns
type ZeroOrMore struct{ Pattern }

// TerminalPattern implements the Pattern interface
func (zom ZeroOrMore) TerminalPattern() Pattern { return Pattern(zom) }

// OneOrMore represents at least one arbitrary patterns
type OneOrMore struct{ Pattern }

// TerminalPattern implements the Pattern interface
func (oom OneOrMore) TerminalPattern() Pattern { return Pattern(oom) }

// Either represents either of the arbitrary patterns
type Either []Pattern

// TerminalPattern implements the Pattern interface
func (Either) TerminalPattern() Pattern { return nil }
