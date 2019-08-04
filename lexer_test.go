package main

import (
	"parser/parser"
	"testing"

	"github.com/stretchr/testify/require"
)

func testSrc(s string) *parser.SourceFile {
	return &parser.SourceFile{
		Name: "test.txt",
		Src:  s,
	}
}

func TestLexerNext(t *testing.T) {
	t.Run("", func(t *testing.T) {
		lex := NewLexer(testSrc("abc\r\n defg"))

		tk1, err := lex.Next()
		require.NoError(t, err)
		require.NotNil(t, tk1)
		require.Equal(t, FrTkLatinAlphanum, tk1.Kind())
		require.Equal(t, "abc", tk1.Src())
		require.Equal(t, uint(1), tk1.Begin().Line)
		require.Equal(t, uint(1), tk1.Begin().Column)

		tk2, err := lex.Next()
		require.NoError(t, err)
		require.NotNil(t, tk2)
		require.Equal(t, FrTkSpace, tk2.Kind())
		require.Equal(t, "\r\n ", tk2.Src())
		require.Equal(t, uint(1), tk2.Begin().Line)
		require.Equal(t, uint(4), tk2.Begin().Column)

		tk3, err := lex.Next()
		require.NoError(t, err)
		require.NotNil(t, tk3)
		require.Equal(t, FrTkLatinAlphanum, tk3.Kind())
		require.Equal(t, "defg", tk3.Src())
		require.Equal(t, uint(2), tk3.Begin().Line)
		require.Equal(t, uint(2), tk3.Begin().Column)

		tk4, err := lex.Next()
		require.NoError(t, err)
		require.Nil(t, tk4)
	})
}

func TestLexerPeek(t *testing.T) {
	t.Run("", func(t *testing.T) {
		lex := NewLexer(testSrc("abc"))
		require.True(t, lex.Peek("abc"))
	})
	t.Run("ShorterThanSrc", func(t *testing.T) {
		lex := NewLexer(testSrc("abc"))
		require.True(t, lex.Peek("ab"))
	})
	t.Run("NoMatch", func(t *testing.T) {
		lex := NewLexer(testSrc("abc"))
		require.False(t, lex.Peek("def"))
	})
	t.Run("OutOfBound", func(t *testing.T) {
		lex := NewLexer(testSrc("abc"))
		require.False(t, lex.Peek("abcd"))
	})
	t.Run("OutOfBound", func(t *testing.T) {
		lex := NewLexer(testSrc(""))
		require.False(t, lex.Peek("ab"))
	})
}
