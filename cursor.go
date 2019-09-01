package parser

import "fmt"

// SourceFile represents a source file
type SourceFile struct {
	Name string
	Src  []rune
}

// Cursor represents a source-code location
type Cursor struct {
	Index  uint
	Column uint
	Line   uint
	File   *SourceFile
}

// NewCursor creates a new cursor location based on the given source file
func NewCursor(file *SourceFile) Cursor {
	return Cursor{
		Index:  0,
		Column: 1,
		Line:   1,
		File:   file,
	}
}

// String stringifies the cursor
func (c Cursor) String() string {
	if c.File == nil {
		return fmt.Sprintf("<unknown>:%d:%d", c.Line, c.Column)
	}
	return fmt.Sprintf("%s:%d:%d", c.File.Name, c.Line, c.Column)
}
