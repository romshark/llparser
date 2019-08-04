package main

import (
	"fmt"
	"parser/parser"
)

// Lexer represents the lexer
type Lexer struct {
	cr parser.Cursor
}

// NewLexer creates a new lexer instance positioned at the beginning
// of the provided source file
func NewLexer(file *parser.SourceFile) *Lexer {
	if file == nil {
		panic("lexer missing source file")
	}
	return &Lexer{parser.NewCursor(file)}
}

// Position returns the current position of the lexer
func (lx *Lexer) Position() parser.Cursor { return lx.cr }

// Fork creates a new lexer branching off the original one
func (lx *Lexer) Fork() parser.Lexer { return &Lexer{lx.cr} }

// Set sets a new cursor
func (lx *Lexer) Set(cursor parser.Cursor) { lx.cr = cursor }

// Peek returns true if str is at the current lexer position,
// otherwise returns false
func (lx *Lexer) Peek(str string) bool {
	ei := 0
	if uint(len(str)) > uint(len(lx.cr.File.Src))-lx.cr.Index {
		return false
	}
	for i := lx.cr.Index; ei < len(str) && i < uint(len(lx.cr.File.Src)); i++ {
		if lx.cr.File.Src[i] != str[ei] {
			return false
		}
		ei++
	}
	return true
}

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

func (lx *Lexer) readSpace() *parser.Token {
	tk := &parser.Token{
		VKind:  FrTkSpace,
		VBegin: lx.cr,
	}
	for {
		if isEOF(lx.cr.File.Src, lx.cr.Index) {
			// EOF
			return finalizedToken(tk, lx.cr)
		}
		if isSpace(lx.cr.File.Src[lx.cr.Index]) {
			lx.cr.Index++
			lx.cr.Column++
			continue
		}
		if end := isLineBreak(lx.cr.File.Src, lx.cr.Index); end != -1 {
			lx.cr.Index = uint(end)
			lx.cr.Column = 1
			lx.cr.Line++
			continue
		}
		break
	}
	// End of sequence
	return finalizedToken(tk, lx.cr)
}

func (lx *Lexer) readLatinAlphanum() *parser.Token {
	tk := &parser.Token{
		VKind:  FrTkLatinAlphanum,
		VBegin: lx.cr,
	}
	for {
		if isEOF(lx.cr.File.Src, lx.cr.Index) {
			// EOF
			return finalizedToken(tk, lx.cr)
		}
		if isLatinAlphanum(lx.cr.File.Src[lx.cr.Index]) {
			lx.cr.Index++
			lx.cr.Column++
			continue
		}
		break
	}
	// End of sequence
	return finalizedToken(tk, lx.cr)
}

func (lx *Lexer) readSingleChar(kind FragmentKind) *parser.Token {
	tk := &parser.Token{
		VKind:  kind,
		VBegin: lx.cr,
	}
	lx.cr.Index++
	lx.cr.Column++
	return finalizedToken(tk, lx.cr)
}

// Next returns either the next token or nil if end of file is reached
func (lx *Lexer) Next() (*parser.Token, error) {
	if lx.cr.Index >= uint(len(lx.cr.File.Src)) {
		// EOF
		return nil, nil
	}
	bt := lx.cr.File.Src[lx.cr.Index]
	switch {
	case isLatinAlphanum(bt):
		return lx.readLatinAlphanum(), nil
	case isSpace(bt):
		return lx.readSpace(), nil
	case isLineBreak(lx.cr.File.Src, lx.cr.Index) != -1:
		return lx.readSpace(), nil
	case bt == '(':
		return lx.readSingleChar(FrTkSymLeftParenthesis), nil
	case bt == ')':
		return lx.readSingleChar(FrTkSymRightParenthesis), nil
	case bt == '{':
		return lx.readSingleChar(FrTkSymLeftCurlyBracket), nil
	case bt == '}':
		return lx.readSingleChar(FrTkSymRightCurlyBracket), nil
	default:
		return nil, fmt.Errorf("unexpected character %d", bt)
	}
}
