package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func helpEnsureEOF(t *testing.T, lex *lexer) {
	tk, err := lex.ReadUntil(func(Cursor) uint { return 1 }, 0)
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
		lex := newLexer(&SourceFile{
			Name: "test.txt",
			Src:  []rune("abc\r\n\t defg,!"),
		})

		tk1, err := lex.ReadUntil(func(Cursor) uint { return 1 }, 1002)
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
	})

	// SplitCRLF tests whether CRLF sequences are splitted. The lexer is
	// expected to skip CRLF sequences as a whole
	t.Run("SplitCRLF", func(t *testing.T) {
		lex := newLexer(&SourceFile{
			Name: "test.txt",
			Src:  []rune("a\r\nbc"),
		})

		until := func(crs Cursor) uint {
			if crs.File.Src[crs.Index] == '\n' {
				// This should only be matched in the second case
				// where there's no carriage-return character in front
				// of the line-feed
				return 0
			}
			return 1
		}
		tk1, err := lex.ReadUntil(until, 1002)

		// Read head
		require.NoError(t, err)
		require.NotNil(t, tk1)
		require.Equal(t, FragmentKind(1002), tk1.Kind())
		require.Equal(t, "a\r\nbc", string(tk1.Src()))

		require.Equal(t, uint(0), tk1.Begin().Index)
		require.Equal(t, uint(1), tk1.Begin().Line)
		require.Equal(t, uint(1), tk1.Begin().Column)

		require.Equal(t, uint(5), tk1.End().Index)
		require.Equal(t, uint(2), tk1.End().Line)
		require.Equal(t, uint(3), tk1.End().Column)

		helpEnsureEOF(t, lex)
	})

	// SkipMultiple tests returning >1 offset returns
	t.Run("SkipMultiple", func(t *testing.T) {
		lex := newLexer(&SourceFile{
			Name: "test.txt",
			Src:  []rune("abc\ndef"),
		})

		tk1, err := lex.ReadUntil(
			func(crs Cursor) uint {
				if crs.File.Src[crs.Index] == 'c' {
					// This condition should never be met because the second
					// will be matched first which will cause the lexer
					// to skip 'c'
					return 0
				}
				if crs.File.Src[crs.Index] == 'b' {
					return 2
				}
				return 1
			},
			1002,
		)
		require.NoError(t, err)
		require.NotNil(t, tk1)
		require.Equal(t, FragmentKind(1002), tk1.Kind())
		require.Equal(t, "abc\ndef", string(tk1.Src()))

		require.Equal(t, uint(0), tk1.Begin().Index)
		require.Equal(t, uint(1), tk1.Begin().Line)
		require.Equal(t, uint(1), tk1.Begin().Column)

		require.Equal(t, uint(7), tk1.End().Index)
		require.Equal(t, uint(2), tk1.End().Line)
		require.Equal(t, uint(4), tk1.End().Column)

		helpEnsureEOF(t, lex)
	})

	// SkipExceed tests returning >1 offsets exceeding the source file size.
	// The lexer is expected not to crash, it should just read until EOF
	t.Run("SkipExceed", func(t *testing.T) {
		lex := newLexer(&SourceFile{
			Name: "test.txt",
			Src:  []rune("abc"),
		})

		tk1, err := lex.ReadUntil(
			func(crs Cursor) uint {
				if crs.File.Src[crs.Index] == 'c' {
					return 2
				}
				return 1
			},
			1002,
		)
		require.NoError(t, err)
		require.NotNil(t, tk1)
		require.Equal(t, FragmentKind(1002), tk1.Kind())
		require.Equal(t, "abc", string(tk1.Src()))

		require.Equal(t, uint(0), tk1.Begin().Index)
		require.Equal(t, uint(1), tk1.Begin().Line)
		require.Equal(t, uint(1), tk1.Begin().Column)

		require.Equal(t, uint(3), tk1.End().Index)
		require.Equal(t, uint(1), tk1.End().Line)
		require.Equal(t, uint(4), tk1.End().Column)

		helpEnsureEOF(t, lex)
	})

	// Nil returning 0 immediately for any cursor
	t.Run("Nil", func(t *testing.T) {
		lex := newLexer(&SourceFile{
			Name: "test.txt",
			Src:  []rune("abc"),
		})

		tk1, err := lex.ReadUntil(
			func(crs Cursor) uint { return 0 },
			1002,
		)
		require.NoError(t, err)
		require.Nil(t, tk1)
	})
}
