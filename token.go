package parser

import "fmt"

// Token represents a terminal fragment
type Token struct {
	VBegin Cursor
	VEnd   Cursor
	VKind  FragmentKind
}

// Kind returns the kind of the token fragment
func (tk *Token) Kind() FragmentKind {
	if tk == nil {
		return 0
	}
	return tk.VKind
}

// Begin returns the beginning cursor of the token fragment
func (tk *Token) Begin() Cursor { return tk.VBegin }

// End returns the ending cursor of the token fragment
func (tk *Token) End() Cursor { return tk.VEnd }

// Src returns the source code of the token fragment
func (tk *Token) Src() []rune {
	if tk == nil {
		return nil
	}
	return tk.VBegin.File.Src[tk.VBegin.Index:tk.VEnd.Index]
}

// Elements always returns nil for terminal fragments
func (tk *Token) Elements() []Fragment { return nil }

// String stringifies the token
func (tk *Token) String() string {
	fileName := "<unknown>"
	if tk.VBegin.File != nil {
		fileName = tk.VBegin.File.Name
	}
	return fmt.Sprintf(
		"%d(%s: %d:%d-%d:%d '%s')",
		tk.VKind,
		fileName,
		tk.VBegin.Line, tk.VBegin.Column,
		tk.VEnd.Line, tk.VEnd.Column,
		string(tk.Src()),
	)
}
