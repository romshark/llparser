package parser

// Action represents a callback function that's called when a certain
// fragment is matched
type Action func(Fragment)

// Rule represents a grammatic rule
type Rule struct {
	Designation string
	Pattern     Pattern
	Kind        FragmentKind
	Action      Action
}

// TerminalPattern implements the Pattern interface
func (rl *Rule) TerminalPattern() Pattern { return rl.Pattern }
