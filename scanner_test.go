package parser_test

import (
	"testing"

	parser "github.com/romshark/llparser"
	"github.com/romshark/llparser/misc"
	"github.com/stretchr/testify/require"
)

func TestScannerRead(t *testing.T) {
	scan := parser.NewScanner(newLexer("abc\n\t defg,"))

	tk1, err := scan.Read()
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

	tk2, err := scan.Read()
	require.NoError(t, err)
	require.NotNil(t, tk2)
	require.Equal(t, misc.FrSpace, tk2.Kind())
	require.Equal(t, "\n\t ", tk2.Src())
	require.Equal(t, uint(3), tk2.Begin().Index)
	require.Equal(t, uint(1), tk2.Begin().Line)
	require.Equal(t, uint(4), tk2.Begin().Column)
	require.Equal(t, uint(6), tk2.End().Index)
	require.Equal(t, uint(2), tk2.End().Line)
	require.Equal(t, uint(3), tk2.End().Column)

	tk3, err := scan.Read()
	require.NoError(t, err)
	require.NotNil(t, tk3)
	require.Equal(t, misc.FrWord, tk3.Kind())
	require.Equal(t, "defg", tk3.Src())
	require.Equal(t, uint(6), tk3.Begin().Index)
	require.Equal(t, uint(2), tk3.Begin().Line)
	require.Equal(t, uint(3), tk3.Begin().Column)
	require.Equal(t, uint(10), tk3.End().Index)
	require.Equal(t, uint(2), tk3.End().Line)
	require.Equal(t, uint(7), tk3.End().Column)

	tk4, err := scan.Read()
	require.NoError(t, err)
	require.NotNil(t, tk4)
	require.Equal(t, misc.FrSign, tk4.Kind())
	require.Equal(t, ",", tk4.Src())
	require.Equal(t, uint(10), tk4.Begin().Index)
	require.Equal(t, uint(2), tk4.Begin().Line)
	require.Equal(t, uint(7), tk4.Begin().Column)
	require.Equal(t, uint(11), tk4.End().Index)
	require.Equal(t, uint(2), tk4.End().Line)
	require.Equal(t, uint(8), tk4.End().Column)

	tk5, err := scan.Read()
	require.NoError(t, err)
	require.Nil(t, tk5)
}
