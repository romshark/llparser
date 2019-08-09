package parser

// Action represents a callback function that's called when a certain
// fragment is matched
type Action func(Fragment) error

// Rule represents a grammatic rule
type Rule struct {
	Designation string
	Pattern     Pattern
	Kind        FragmentKind
	Action      Action
}

// Container implements the Pattern interface
func (*Rule) Container() bool { return false }

// TerminalPattern implements the Pattern interface
func (rl *Rule) TerminalPattern() Pattern { return rl.Pattern }

// Desig implements the Pattern interface
func (rl *Rule) Desig() string { return rl.Designation }
