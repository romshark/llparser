package parser_test

import (
	"testing"

	parser "github.com/romshark/llparser"

	"github.com/stretchr/testify/require"
)

const (
	_ parser.FragmentKind = iota
	TestFrSpace
	TestFrSeq
	TestFrSep
	TestFrA
	TestFrB
)

// Basic terminal types
var (
	testR_A = &parser.Rule{
		Designation: "Terminal A",
		Pattern:     parser.TermExact("a"),
		Kind:        TestFrA,
	}
	testR_B = &parser.Rule{
		Designation: "Terminal B",
		Pattern:     parser.TermExact("b"),
		Kind:        TestFrB,
	}
)

// TestLexer represents a lexer implementation for testing purposes
type TestLexer struct{ cr parser.Cursor }

func NewTestLexer(src string) *TestLexer {
	return &TestLexer{
		cr: parser.NewCursor(&parser.SourceFile{
			Name: "test.txt",
			Src:  src,
		}),
	}
}

// Position returns the current position of the lexer
func (lx *TestLexer) Position() parser.Cursor { return lx.cr }

// Fork creates a new lexer branching off the original one
func (lx *TestLexer) Fork() parser.Lexer { return &TestLexer{lx.cr} }

// Set sets a new cursor
func (lx *TestLexer) Set(cursor parser.Cursor) { lx.cr = cursor }

// CheckCursor checks a cursor relative to the lexer
func (lx *TestLexer) CheckCursor(
	t *testing.T,
	cursor parser.Cursor,
	line,
	column uint,
) {
	require.Equal(t, lx.cr.File, cursor.File)
	if column > 1 || line > 1 {
		require.True(t, cursor.Index > 0)
	} else if column == 1 && line == 1 {
		require.Equal(t, uint(0), cursor.Index)
	}
	require.Equal(t, line, cursor.Line)
	require.Equal(t, column, cursor.Column)
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

func (lx *TestLexer) reachedEOF() bool {
	if lx.cr.Index >= uint(len(lx.cr.File.Src)) {
		return true
	}
	return false
}

func (lx *TestLexer) readSpace() *parser.Token {
	tk := &parser.Token{
		VKind:  TestFrSpace,
		VBegin: lx.cr,
	}
	for {
		if lx.reachedEOF() {
			// EOF
			return finalizedToken(tk, lx.cr)
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
		}
		break
	}
	// End of sequence
	return finalizedToken(tk, lx.cr)
}

func (lx *TestLexer) readSequence() *parser.Token {
	tk := &parser.Token{
		VKind:  TestFrSeq,
		VBegin: lx.cr,
	}
	for {
		if lx.reachedEOF() {
			// EOF
			return finalizedToken(tk, lx.cr)
		}
		switch lx.cr.File.Src[lx.cr.Index] {
		case ' ':
		case '\t':
		case '\n':
		case ',':
		default:
			lx.cr.Index++
			lx.cr.Column++
			continue
		}
		break
	}
	// End of sequence
	return finalizedToken(tk, lx.cr)
}

func (lx *TestLexer) readSingleChar(
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

// Next returns either the next token or nil if end of file is reached
func (lx *TestLexer) Next() (*parser.Token, error) {
	if lx.reachedEOF() {
		// EOF
		return nil, nil
	}
	bt := lx.cr.File.Src[lx.cr.Index]
	switch bt {
	case ' ':
		fallthrough
	case '\t':
		fallthrough
	case '\n':
		return lx.readSpace(), nil
	case ',':
		return lx.readSingleChar(TestFrSep), nil
	default:
		return lx.readSequence(), nil
	}
}

func TestLexerNext(t *testing.T) {
	lex := NewTestLexer("abc\n\t defg,")

	tk1, err := lex.Next()
	require.NoError(t, err)
	require.NotNil(t, tk1)
	require.Equal(t, TestFrSeq, tk1.Kind())
	require.Equal(t, "abc", tk1.Src())
	require.Equal(t, uint(0), tk1.Begin().Index)
	require.Equal(t, uint(1), tk1.Begin().Line)
	require.Equal(t, uint(1), tk1.Begin().Column)
	require.Equal(t, uint(3), tk1.End().Index)
	require.Equal(t, uint(1), tk1.End().Line)
	require.Equal(t, uint(4), tk1.End().Column)

	tk2, err := lex.Next()
	require.NoError(t, err)
	require.NotNil(t, tk2)
	require.Equal(t, TestFrSpace, tk2.Kind())
	require.Equal(t, "\n\t ", tk2.Src())
	require.Equal(t, uint(3), tk2.Begin().Index)
	require.Equal(t, uint(1), tk2.Begin().Line)
	require.Equal(t, uint(4), tk2.Begin().Column)
	require.Equal(t, uint(6), tk2.End().Index)
	require.Equal(t, uint(2), tk2.End().Line)
	require.Equal(t, uint(3), tk2.End().Column)

	tk3, err := lex.Next()
	require.NoError(t, err)
	require.NotNil(t, tk3)
	require.Equal(t, TestFrSeq, tk3.Kind())
	require.Equal(t, "defg", tk3.Src())
	require.Equal(t, uint(6), tk3.Begin().Index)
	require.Equal(t, uint(2), tk3.Begin().Line)
	require.Equal(t, uint(3), tk3.Begin().Column)
	require.Equal(t, uint(10), tk3.End().Index)
	require.Equal(t, uint(2), tk3.End().Line)
	require.Equal(t, uint(7), tk3.End().Column)

	tk4, err := lex.Next()
	require.NoError(t, err)
	require.NotNil(t, tk4)
	require.Equal(t, TestFrSep, tk4.Kind())
	require.Equal(t, ",", tk4.Src())
	require.Equal(t, uint(10), tk4.Begin().Index)
	require.Equal(t, uint(2), tk4.Begin().Line)
	require.Equal(t, uint(7), tk4.Begin().Column)
	require.Equal(t, uint(11), tk4.End().Index)
	require.Equal(t, uint(2), tk4.End().Line)
	require.Equal(t, uint(8), tk4.End().Column)

	tk5, err := lex.Next()
	require.NoError(t, err)
	require.Nil(t, tk5)
}
