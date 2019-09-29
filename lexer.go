package parser

import (
	"errors"
)

// lexer represents a basic lexer tokenizing source code
// into 3 basic categories: spaces (whitespaces, tabs, line-breaks),
// signs (any ASCII special character) and
// words (any other character)
type lexer struct{ cr Cursor }

// newLexer creates a new basic-latin lexer instance
func newLexer(src *SourceFile) *lexer {
	if src == nil {
		panic("missing source file during lexer initialization")
	}
	return &lexer{
		cr: NewCursor(src),
	}
}

func finalizedToken(
	tk *Token,
	end Cursor,
) *Token {
	if end.Index == tk.VBegin.Index {
		return nil
	}
	tk.VEnd = end
	return tk
}

func (lx *lexer) reachedEOF() bool {
	return lx.cr.Index >= uint(len(lx.cr.File.Src))
}

// ReadExact tries to read an exact string and returns false if
// str couldn't have been matched
func (lx *lexer) ReadExact(
	expectation []rune,
	kind FragmentKind,
) (
	token *Token,
	matched bool,
	err error,
) {
	if len(expectation) < 1 {
		panic(errors.New("empty string expected"))
	}
	if lx.reachedEOF() {
		return nil, false, errEOF{}
	}

	token = &Token{
		VKind:  kind,
		VBegin: lx.cr,
	}

	for ix := 0; ix < len(expectation); ix++ {
		if lx.reachedEOF() {
			return finalizedToken(token, lx.cr), false, nil
		}

		// Check against the expectation
		rn := lx.cr.File.Src[lx.cr.Index]

		// Advance the cursor
		switch rn {
		case '\n':
			lx.cr.Index++
			lx.cr.Column = 1
			lx.cr.Line++
		default:
			lx.cr.Index++
			lx.cr.Column++
		}

		if rn != expectation[ix] {
			// No match
			return finalizedToken(token, lx.cr), false, nil
		}
	}

	return finalizedToken(token, lx.cr), true, nil
}

// ReadUntil reads until fn returns zero skipping as many runes as fn returns
func (lx *lexer) ReadUntil(
	fn func(Cursor) uint,
	kind FragmentKind,
) (*Token, error) {
	if lx.reachedEOF() {
		return nil, errEOF{}
	}

	token := &Token{
		VKind:  kind,
		VBegin: lx.cr,
	}

	for {
		skip := fn(lx.cr)
		if skip < 1 {
			break
		}
		for ix2 := uint(0); ix2 < skip; {
			if lx.reachedEOF() {
				return finalizedToken(token, lx.cr), nil
			}

			// Check against the expectation
			skipSpace := isLineBreak(lx.cr.File.Src, lx.cr.Index)

			// Advance the cursor
			if skipSpace != -1 {
				// Space character
				lx.cr.Index += uint(skipSpace)
				lx.cr.Column = 1
				lx.cr.Line++
				ix2 += uint(skipSpace)
			} else {
				// Non-space character
				lx.cr.Index++
				lx.cr.Column++
				ix2++
			}
		}

		if lx.reachedEOF() {
			// Premature EOF
			return finalizedToken(token, lx.cr), nil
		}
	}

	return finalizedToken(token, lx.cr), nil
}
