package misc

import (
	"errors"
	"fmt"

	parser "github.com/romshark/llparser"
)

const (
	_ parser.FragmentKind = iota

	// FrSpace represents a space fragment kind
	FrSpace

	// FrWord represents a word fragment kind
	FrWord

	// FrSign represents a special character fragment kind
	FrSign
)

// Lexer represents a basic lexer tokenizing source code
// into 3 basic categories: spaces (whitespaces, tabs, line-breaks),
// signs (any ASCII special character) and
// words (any other character)
type Lexer struct{ cr parser.Cursor }

// NewLexer creates a new basic-latin lexer instance
func NewLexer(src *parser.SourceFile) *Lexer {
	if src == nil {
		panic("missing source file during lexer initialization")
	}
	return &Lexer{
		cr: parser.NewCursor(src),
	}
}

// Position returns the current position of the lexer
func (lx *Lexer) Position() parser.Cursor { return lx.cr }

// Set sets a new cursor
func (lx *Lexer) Set(cursor parser.Cursor) { lx.cr = cursor }

func finalizedToken(
	tk *parser.Token,
	end parser.Cursor,
) *parser.Token {
	if end.Index == tk.VBegin.Index {
		return nil
	}
	tk.VEnd = end
	return tk
}

func (lx *Lexer) reachedEOF() bool {
	return lx.cr.Index >= uint(len(lx.cr.File.Src))
}

func (lx *Lexer) readSpace() (*parser.Token, error) {
	tk := &parser.Token{
		VKind:  FrSpace,
		VBegin: lx.cr,
	}
	for {
		if lx.reachedEOF() {
			// EOF
			return finalizedToken(tk, lx.cr), nil
		}

		switch lx.cr.File.Src[lx.cr.Index] {
		case ' ':
			fallthrough
		case '\t':
			lx.cr.Index++
			lx.cr.Column++
			continue
		case '\n':
			lx.cr.Index++
			lx.cr.Column = 1
			lx.cr.Line++
			continue
		case '\r':
			sz := isLineBreak(lx.cr.File.Src, lx.cr.Index)
			if sz < 0 {
				// Dangling carriage-return character
				return nil, fmt.Errorf(
					"unexpected character %d at %s",
					lx.cr.File.Src[lx.cr.Index],
					lx.cr,
				)
			}
			lx.cr.Index += uint(sz)
			lx.cr.Column = 1
			lx.cr.Line++
			continue
		}
		break
	}
	// End of sequence
	return finalizedToken(tk, lx.cr), nil
}

func (lx *Lexer) readWord() *parser.Token {
	tk := &parser.Token{
		VKind:  FrWord,
		VBegin: lx.cr,
	}
	for {
		if lx.reachedEOF() {
			break
		}
		bt := lx.cr.File.Src[lx.cr.Index]
		if isSpace(bt) ||
			isLineBreak(lx.cr.File.Src, lx.cr.Index) != -1 ||
			isSpecialChar(bt) {
			// EOF
			break
		}
		lx.cr.Index++
		lx.cr.Column++
	}
	// End of sequence
	return finalizedToken(tk, lx.cr)
}

func (lx *Lexer) readSingleChar(
	kind parser.FragmentKind,
) *parser.Token {
	tk := &parser.Token{
		VKind:  kind,
		VBegin: lx.cr,
	}
	lx.cr.Index++
	lx.cr.Column++
	return finalizedToken(tk, lx.cr)
}

// Read returns either the next token or nil if end of file is reached
func (lx *Lexer) Read() (*parser.Token, error) {
	if lx.reachedEOF() {
		// EOF
		return nil, nil
	}
	bt := lx.cr.File.Src[lx.cr.Index]
	switch {
	case bt == ' ':
		fallthrough
	case bt == '\t':
		fallthrough
	case isLineBreak(lx.cr.File.Src, lx.cr.Index) != -1:
		return lx.readSpace()
	case isSpecialChar(bt):
		return lx.readSingleChar(FrSign), nil
	default:
		return lx.readWord(), nil
	}
}

// ReadExact tries to read an exact string and returns false if
// str couldn't have been matched
func (lx *Lexer) ReadExact(
	expectation []rune,
	kind parser.FragmentKind,
) (
	token *parser.Token,
	matched bool,
	err error,
) {
	if len(expectation) < 1 {
		return nil, false, errors.New("empty string expected")
	}

	token = &parser.Token{
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
func (lx *Lexer) ReadUntil(
	fn func(parser.Cursor) uint,
	kind parser.FragmentKind,
) (
	token *parser.Token,
	err error,
) {
	token = &parser.Token{
		VKind:  kind,
		VBegin: lx.cr,
	}

	for {
		if lx.reachedEOF() {
			return finalizedToken(token, lx.cr), nil
		}

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
	}

	return finalizedToken(token, lx.cr), nil
}
