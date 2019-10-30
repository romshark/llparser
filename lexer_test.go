package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func helpEnsureEOF(t *testing.T, lex *lexer) {
	tk, err := lex.ReadUntil(func(uint, Cursor) bool { return true }, 0)
	require.Error(t, err)
	require.IsType(t, errEOF{}, err)
	require.Nil(t, tk)
}

func TestLexerReadExact(t *testing.T) {
	lex := newLexer(&SourceFile{
		Name: "test.txt",
		Src:  []rune("abc\r\n\t defg,!"),
	})

	tk1, match, err := lex.ReadExact([]rune("abc\r\n\t defg,!"), 1002)
	require.NoError(t, err)
	require.NotNil(t, tk1)
	require.True(t, match)
	require.Equal(t, FragmentKind(1002), tk1.Kind())
	require.Equal(t, "abc\r\n\t defg,!", string(tk1.Src()))
	require.Equal(t, uint(0), tk1.Begin().Index)
	require.Equal(t, uint(1), tk1.Begin().Line)
	require.Equal(t, uint(1), tk1.Begin().Column)
	require.Equal(t, uint(13), tk1.End().Index)
	require.Equal(t, uint(2), tk1.End().Line)
	require.Equal(t, uint(9), tk1.End().Column)

	helpEnsureEOF(t, lex)
}

func TestLexerReadExactNoMatch(t *testing.T) {
	lex := newLexer(&SourceFile{
		Name: "test.txt",
		Src:  []rune("abc\r\n\t defg,!"),
	})

	tk1, match1, err1 := lex.ReadExact([]rune("ac"), 1002)
	require.NoError(t, err1)
	require.NotNil(t, tk1)
	require.False(t, match1)
	require.Equal(t, FragmentKind(1002), tk1.Kind())
	require.Equal(t, "ab", string(tk1.Src()))
	require.Equal(t, uint(0), tk1.Begin().Index)
	require.Equal(t, uint(1), tk1.Begin().Line)
	require.Equal(t, uint(1), tk1.Begin().Column)
	require.Equal(t, uint(2), tk1.End().Index)
	require.Equal(t, uint(1), tk1.End().Line)
	require.Equal(t, uint(3), tk1.End().Column)

	tk2, match2, err2 := lex.ReadExact([]rune("c"), 1003)
	require.NoError(t, err2)
	require.NotNil(t, tk2)
	require.True(t, match2)
	require.Equal(t, FragmentKind(1003), tk2.Kind())
	require.Equal(t, "c", string(tk2.Src()))
	require.Equal(t, uint(2), tk2.Begin().Index)
	require.Equal(t, uint(1), tk2.Begin().Line)
	require.Equal(t, uint(3), tk2.Begin().Column)
	require.Equal(t, uint(3), tk2.End().Index)
	require.Equal(t, uint(1), tk2.End().Line)
	require.Equal(t, uint(4), tk2.End().Column)
}

// TestLexerReadUntil tests all ReadUntil cases
func TestLexerReadUntil(t *testing.T) {
	// MatchAll tests matching any input character
	t.Run("MatchAll", func(t *testing.T) {
		src := []rune("abc\r\n\t defg,!")
		lex := newLexer(&SourceFile{
			Name: "test.txt",
			Src:  src,
		})

		lastIx := uint(0)

		tk1, err := lex.ReadUntil(
			func(ix uint, _ Cursor) bool {
				lastIx = ix
				return true
			},
			1002,
		)
		require.NoError(t, err)
		require.NotNil(t, tk1)
		require.Equal(t, FragmentKind(1002), tk1.Kind())
		require.Equal(t, "abc\r\n\t defg,!", string(tk1.Src()))

		require.Equal(t, uint(0), tk1.Begin().Index)
		require.Equal(t, uint(1), tk1.Begin().Line)
		require.Equal(t, uint(1), tk1.Begin().Column)

		require.Equal(t, uint(13), tk1.End().Index)
		require.Equal(t, uint(2), tk1.End().Line)
		require.Equal(t, uint(9), tk1.End().Column)

		helpEnsureEOF(t, lex)

		require.Equal(t, uint(len(src)-1), lastIx)
	})

	// SplitCRLF tests splitting CRLF sequences
	t.Run("SplitCRLF", func(t *testing.T) {
		lex := newLexer(&SourceFile{
			Name: "test.txt",
			Src:  []rune("a\r\nbc"),
		})

		until := func(ix uint, crs Cursor) bool {
			if ix != 0 && crs.File.Src[crs.Index] == '\n' {
				return false
			}
			return true
		}

		// Read head
		tk1, err := lex.ReadUntil(until, 1002)
		require.NoError(t, err)
		require.NotNil(t, tk1)
		require.Equal(t, FragmentKind(1002), tk1.Kind())
		require.Equal(t, "a\r", string(tk1.Src()))

		require.Equal(t, uint(0), tk1.Begin().Index)
		require.Equal(t, uint(1), tk1.Begin().Line)
		require.Equal(t, uint(1), tk1.Begin().Column)

		require.Equal(t, uint(2), tk1.End().Index)
		require.Equal(t, uint(1), tk1.End().Line)
		require.Equal(t, uint(3), tk1.End().Column)

		// Read tail
		tk2, err := lex.ReadUntil(until, 1002)
		require.NoError(t, err)
		require.NotNil(t, tk2)
		require.Equal(t, FragmentKind(1002), tk2.Kind())
		require.Equal(t, "\nbc", string(tk2.Src()))

		require.Equal(t, uint(2), tk2.Begin().Index)
		require.Equal(t, uint(1), tk2.Begin().Line)
		require.Equal(t, uint(3), tk2.Begin().Column)

		require.Equal(t, uint(5), tk2.End().Index)
		require.Equal(t, uint(2), tk2.End().Line)
		require.Equal(t, uint(3), tk2.End().Column)

		helpEnsureEOF(t, lex)
	})

	// Nil returning false immediately for any cursor
	t.Run("Nil", func(t *testing.T) {
		lex := newLexer(&SourceFile{
			Name: "test.txt",
			Src:  []rune("abc"),
		})

		tk1, err := lex.ReadUntil(
			func(uint, Cursor) bool { return false },
			1002,
		)
		require.NoError(t, err)
		require.Nil(t, tk1)
	})
}
