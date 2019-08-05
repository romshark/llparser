package parser_test

import (
	"testing"

	parser "github.com/romshark/llparser"
	"github.com/stretchr/testify/require"
)

func TestParserSequence(t *testing.T) {
	t.Run("SingleLevel", func(t *testing.T) {
		pr := parser.NewParser()
		lx := NewTestLexer("foo   bar")
		expectedKind := parser.FragmentKind(100)
		mainFrag, err := pr.Parse(lx, &parser.Rule{
			Designation: "foobar",
			Pattern: parser.Sequence{
				parser.Term(TestFrSeq),
				parser.Term(TestFrSpace),
				parser.Term(TestFrSeq),
			},
			Kind: expectedKind,
		})

		require.NoError(t, err)
		require.NotNil(t, mainFrag)
		require.Equal(t, expectedKind, mainFrag.Kind())
		lx.CheckCursor(t, mainFrag.Begin(), 1, 1)
		lx.CheckCursor(t, mainFrag.End(), 1, 10)

		// Check elements
		elements := mainFrag.Elements()
		require.Len(t, elements, 3)

		// Check element 1
		elem1 := elements[0]
		require.Equal(t, TestFrSeq, elem1.Kind())
		require.Nil(t, elem1.Elements())
		lx.CheckCursor(t, elem1.Begin(), 1, 1)
		lx.CheckCursor(t, elem1.End(), 1, 4)

		// Check element 2
		elem2 := elements[1]
		require.Equal(t, TestFrSpace, elem2.Kind())
		require.Nil(t, elem2.Elements())
		lx.CheckCursor(t, elem2.Begin(), 1, 4)
		lx.CheckCursor(t, elem2.End(), 1, 7)

		// Check element 3
		elem3 := elements[2]
		require.Equal(t, TestFrSeq, elem3.Kind())
		require.Nil(t, elem3.Elements())
		lx.CheckCursor(t, elem3.Begin(), 1, 7)
		lx.CheckCursor(t, elem3.End(), 1, 10)
	})

	t.Run("TwoLevels", func(t *testing.T) {
		pr := parser.NewParser()
		lx := NewTestLexer("foo   bar")
		expectedKind := parser.FragmentKind(100)
		mainFrag, err := pr.Parse(lx, &parser.Rule{
			Designation: "foobar",
			Pattern: parser.Sequence{
				testR_foo,
				parser.Term(TestFrSpace),
				testR_bar},
			Kind: expectedKind,
		})
		require.NoError(t, err)
		require.NotNil(t, mainFrag)
		require.Equal(t, expectedKind, mainFrag.Kind())
		lx.CheckCursor(t, mainFrag.Begin(), 1, 1)
		lx.CheckCursor(t, mainFrag.End(), 1, 10)

		// Check elements
		elements := mainFrag.Elements()
		require.Len(t, elements, 3)

		// Check element 1
		elem1 := elements[0]
		require.Equal(t, TestFrA, elem1.Kind())
		require.Nil(t, elem1.Elements())
		lx.CheckCursor(t, elem1.Begin(), 1, 1)
		lx.CheckCursor(t, elem1.End(), 1, 4)

		// Check element 2
		elem2 := elements[1]
		require.Equal(t, TestFrSpace, elem2.Kind())
		require.Nil(t, elem2.Elements())
		lx.CheckCursor(t, elem2.Begin(), 1, 4)
		lx.CheckCursor(t, elem2.End(), 1, 7)

		// Check element 3
		elem3 := elements[2]
		require.Equal(t, TestFrB, elem3.Kind())
		require.Nil(t, elem3.Elements())
		lx.CheckCursor(t, elem3.Begin(), 1, 7)
		lx.CheckCursor(t, elem3.End(), 1, 10)
	})
}

// TestParserSequenceErr tests sequence parsing errors
func TestParserSequenceErr(t *testing.T) {
	t.Run("UnexpectedToken(expect_rule, first_item)", func(t *testing.T) {
		pr := parser.NewParser()
		lx := NewTestLexer("foo")
		expectedKind := parser.FragmentKind(100)
		mainFrag, err := pr.Parse(lx, &parser.Rule{
			Designation: "foobar",
			Pattern: parser.Sequence{
				testR_bar,
				testR_foo,
			},
			Kind: expectedKind,
		})

		require.Error(t, err)
		require.Equal(
			t,
			"unexpected token 'foo', expected {keyword bar} at test.txt:1:1",
			err.Error(),
		)
		require.Nil(t, mainFrag)
	})
	t.Run("UnexpectedToken(expect_rule, second_item)", func(t *testing.T) {
		pr := parser.NewParser()
		lx := NewTestLexer("foo foo")
		expectedKind := parser.FragmentKind(100)
		mainFrag, err := pr.Parse(lx, &parser.Rule{
			Designation: "foobar",
			Pattern: parser.Sequence{
				testR_foo,
				parser.Term(TestFrSpace),
				testR_bar,
			},
			Kind: expectedKind,
		})

		require.Error(t, err)
		require.Equal(
			t,
			"unexpected token 'foo', expected {keyword bar} at test.txt:1:5",
			err.Error(),
		)
		require.Nil(t, mainFrag)
	})
}
