package misc_test

import (
	"testing"

	parser "github.com/romshark/llparser"
	"github.com/romshark/llparser/misc"
	"github.com/stretchr/testify/require"
)

func TestLexerRead(t *testing.T) {
	lex := misc.NewLexer(&parser.SourceFile{
		Name: "test.txt",
		Src:  "abc\r\n\t defg,!",
	})

	tk1, err := lex.Read()
	require.NoError(t, err)
	require.NotNil(t, tk1)
	require.Equal(t, misc.FrWord, tk1.Kind())
	require.Equal(t, "abc", tk1.Src())
	require.Equal(t, uint(0), tk1.Begin().Index)
	require.Equal(t, uint(1), tk1.Begin().Line)
	require.Equal(t, uint(1), tk1.Begin().Column)
	require.Equal(t, uint(3), tk1.End().Index)
	require.Equal(t, uint(1), tk1.End().Line)
	require.Equal(t, uint(4), tk1.End().Column)

	tk2, err := lex.Read()
	require.NoError(t, err)
	require.NotNil(t, tk2)
	require.Equal(t, misc.FrSpace, tk2.Kind())
	require.Equal(t, "\r\n\t ", tk2.Src())
	require.Equal(t, uint(3), tk2.Begin().Index)
	require.Equal(t, uint(1), tk2.Begin().Line)
	require.Equal(t, uint(4), tk2.Begin().Column)
	require.Equal(t, uint(7), tk2.End().Index)
	require.Equal(t, uint(2), tk2.End().Line)
	require.Equal(t, uint(3), tk2.End().Column)

	tk3, err := lex.Read()
	require.NoError(t, err)
	require.NotNil(t, tk3)
	require.Equal(t, misc.FrWord, tk3.Kind())
	require.Equal(t, "defg", tk3.Src())
	require.Equal(t, uint(7), tk3.Begin().Index)
	require.Equal(t, uint(2), tk3.Begin().Line)
	require.Equal(t, uint(3), tk3.Begin().Column)
	require.Equal(t, uint(11), tk3.End().Index)
	require.Equal(t, uint(2), tk3.End().Line)
	require.Equal(t, uint(7), tk3.End().Column)

	tk4, err := lex.Read()
	require.NoError(t, err)
	require.NotNil(t, tk4)
	require.Equal(t, misc.FrSign, tk4.Kind())
	require.Equal(t, ",", tk4.Src())
	require.Equal(t, uint(11), tk4.Begin().Index)
	require.Equal(t, uint(2), tk4.Begin().Line)
	require.Equal(t, uint(7), tk4.Begin().Column)
	require.Equal(t, uint(12), tk4.End().Index)
	require.Equal(t, uint(2), tk4.End().Line)
	require.Equal(t, uint(8), tk4.End().Column)

	tk5, err := lex.Read()
	require.NoError(t, err)
	require.NotNil(t, tk5)
	require.Equal(t, misc.FrSign, tk5.Kind())
	require.Equal(t, "!", tk5.Src())
	require.Equal(t, uint(12), tk5.Begin().Index)
	require.Equal(t, uint(2), tk5.Begin().Line)
	require.Equal(t, uint(8), tk5.Begin().Column)
	require.Equal(t, uint(13), tk5.End().Index)
	require.Equal(t, uint(2), tk5.End().Line)
	require.Equal(t, uint(9), tk5.End().Column)

	tk6, err := lex.Read()
	require.NoError(t, err)
	require.Nil(t, tk6)
}

func TestLexerReadExact(t *testing.T) {
	lex := misc.NewLexer(&parser.SourceFile{
		Name: "test.txt",
		Src:  "abc\r\n\t defg,!",
	})

	tk1, match, err := lex.ReadExact("abc\r\n\t defg,!", 1002)
	require.NoError(t, err)
	require.NotNil(t, tk1)
	require.True(t, match)
	require.Equal(t, parser.FragmentKind(1002), tk1.Kind())
	require.Equal(t, "abc\r\n\t defg,!", tk1.Src())
	require.Equal(t, uint(0), tk1.Begin().Index)
	require.Equal(t, uint(1), tk1.Begin().Line)
	require.Equal(t, uint(1), tk1.Begin().Column)
	require.Equal(t, uint(13), tk1.End().Index)
	require.Equal(t, uint(2), tk1.End().Line)
	require.Equal(t, uint(9), tk1.End().Column)

	tk6, err := lex.Read()
	require.NoError(t, err)
	require.Nil(t, tk6)
}

func TestLexerReadExactNoMatch(t *testing.T) {
	lex := misc.NewLexer(&parser.SourceFile{
		Name: "test.txt",
		Src:  "abc\r\n\t defg,!",
	})

	tk1, match1, err1 := lex.ReadExact("ac", 1002)
	require.NoError(t, err1)
	require.NotNil(t, tk1)
	require.False(t, match1)
	require.Equal(t, parser.FragmentKind(1002), tk1.Kind())
	require.Equal(t, "ab", tk1.Src())
	require.Equal(t, uint(0), tk1.Begin().Index)
	require.Equal(t, uint(1), tk1.Begin().Line)
	require.Equal(t, uint(1), tk1.Begin().Column)
	require.Equal(t, uint(2), tk1.End().Index)
	require.Equal(t, uint(1), tk1.End().Line)
	require.Equal(t, uint(3), tk1.End().Column)

	tk2, match2, err2 := lex.ReadExact("c", 1003)
	require.NoError(t, err2)
	require.NotNil(t, tk2)
	require.True(t, match2)
	require.Equal(t, parser.FragmentKind(1003), tk2.Kind())
	require.Equal(t, "c", tk2.Src())
	require.Equal(t, uint(2), tk2.Begin().Index)
	require.Equal(t, uint(1), tk2.Begin().Line)
	require.Equal(t, uint(3), tk2.Begin().Column)
	require.Equal(t, uint(3), tk2.End().Index)
	require.Equal(t, uint(1), tk2.End().Line)
	require.Equal(t, uint(4), tk2.End().Column)
}
