package parser

// FragmentKind represents a fragment kind identifier
type FragmentKind int

// Fragment represents an abstract fragment
type Fragment interface {
	Kind() FragmentKind
	Begin() Cursor
	End() Cursor
	Src() []rune
	Elements() []Fragment
}
