package parser_test

import (
	"testing"

	parser "github.com/romshark/llparser"
	"github.com/stretchr/testify/require"
)

func TestScannerNext(t *testing.T) {
	scan := parser.NewScanner(NewTestLexer("abc\n\t defg,"))

	tk1, err := scan.Next()
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

	tk2, err := scan.Next()
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

	tk3, err := scan.Next()
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

	tk4, err := scan.Next()
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

	tk5, err := scan.Next()
	require.NoError(t, err)
	require.Nil(t, tk5)
}

func TestScannerBack(t *testing.T) {
	t.Run("OneStep", func(t *testing.T) {
		scan := parser.NewScanner(NewTestLexer("abc"))

		tk1, err := scan.Next()
		require.NoError(t, err)
		require.NotNil(t, tk1)
		require.Equal(t, TestFrSeq, tk1.Kind())
		require.Equal(t, "abc", tk1.Src())
		require.Equal(t, uint(0), tk1.Begin().Index)
		require.Equal(t, uint(1), tk1.Begin().Line)
		require.Equal(t, uint(1), tk1.Begin().Column)

		scan.Back(1)

		tk2, err := scan.Next()
		require.NoError(t, err)
		require.NotNil(t, tk2)
		require.Equal(t, TestFrSeq, tk2.Kind())
		require.Equal(t, "abc", tk2.Src())
		require.Equal(t, uint(0), tk1.Begin().Index)
		require.Equal(t, uint(1), tk2.Begin().Line)
		require.Equal(t, uint(1), tk2.Begin().Column)
	})
	t.Run("MultipleSteps", func(t *testing.T) {
		scan := parser.NewScanner(NewTestLexer("abc\ndef"))

		testSequence := func() {
			tk1, err := scan.Next()
			require.NoError(t, err)
			require.NotNil(t, tk1)
			require.Equal(t, TestFrSeq, tk1.Kind())
			require.Equal(t, "abc", tk1.Src())
			require.Equal(t, uint(0), tk1.Begin().Index)
			require.Equal(t, uint(1), tk1.Begin().Line)
			require.Equal(t, uint(1), tk1.Begin().Column)

			tk2, err := scan.Next()
			require.NoError(t, err)
			require.NotNil(t, tk2)
			require.Equal(t, TestFrSpace, tk2.Kind())
			require.Equal(t, "\n", tk2.Src())
			require.Equal(t, uint(3), tk2.Begin().Index)
			require.Equal(t, uint(1), tk2.Begin().Line)
			require.Equal(t, uint(4), tk2.Begin().Column)

			tk3, err := scan.Next()
			require.NoError(t, err)
			require.NotNil(t, tk3)
			require.Equal(t, TestFrSeq, tk3.Kind())
			require.Equal(t, "def", tk3.Src())
			require.Equal(t, uint(4), tk3.Begin().Index)
			require.Equal(t, uint(2), tk3.Begin().Line)
			require.Equal(t, uint(1), tk3.Begin().Column)

			tk4, err := scan.Next()
			require.NoError(t, err)
			require.Nil(t, tk4)
		}

		testSequence()
		scan.Back(3)
		testSequence()
	})
}
