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
		require.Equal(t, TestFrFoo, elem1.Kind())
		lx.CheckCursor(t, elem1.Begin(), 1, 1)
		lx.CheckCursor(t, elem1.End(), 1, 4)
		require.Len(t, elem1.Elements(), 1)

		// Check element 2
		elem2 := elements[1]
		require.Equal(t, TestFrSpace, elem2.Kind())
		lx.CheckCursor(t, elem2.Begin(), 1, 4)
		lx.CheckCursor(t, elem2.End(), 1, 7)
		require.Nil(t, elem2.Elements())

		// Check element 3
		elem3 := elements[2]
		require.Equal(t, TestFrBar, elem3.Kind())
		lx.CheckCursor(t, elem3.Begin(), 1, 7)
		lx.CheckCursor(t, elem3.End(), 1, 10)
		require.Len(t, elem3.Elements(), 1)
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

func TestParserOptionalInSequence(t *testing.T) {
	t.Run("Missing", func(t *testing.T) {
		pr := parser.NewParser()
		lx := NewTestLexer("bar")
		expectedKind := parser.FragmentKind(100)
		mainFrag, err := pr.Parse(lx, &parser.Rule{
			Designation: "?foo bar",
			Pattern: parser.Sequence{
				parser.Optional{parser.Sequence{
					testR_foo,
					parser.Term(TestFrSpace),
				}},
				testR_bar,
			},
			Kind: expectedKind,
		})

		require.NoError(t, err)
		require.NotNil(t, mainFrag)
		require.Equal(t, expectedKind, mainFrag.Kind())
		lx.CheckCursor(t, mainFrag.Begin(), 1, 1)
		lx.CheckCursor(t, mainFrag.End(), 1, 4)

		// Check elements
		elements := mainFrag.Elements()
		require.Len(t, elements, 1)

		// Check element 1
		elem1 := elements[0]
		require.Equal(t, TestFrBar, elem1.Kind())
		require.Len(t, elem1.Elements(), 1)
		lx.CheckCursor(t, elem1.Begin(), 1, 1)
		lx.CheckCursor(t, elem1.End(), 1, 4)
	})

	t.Run("Present", func(t *testing.T) {
		pr := parser.NewParser()
		lx := NewTestLexer("foo bar")
		expectedKind := parser.FragmentKind(100)
		mainFrag, err := pr.Parse(lx, &parser.Rule{
			Designation: "?foo bar",
			Pattern: parser.Sequence{
				parser.Optional{parser.Sequence{
					testR_foo,
					parser.Term(TestFrSpace),
				}},
				testR_bar,
			},
			Kind: expectedKind,
		})
		require.NoError(t, err)
		require.NotNil(t, mainFrag)
		require.Equal(t, expectedKind, mainFrag.Kind())
		lx.CheckCursor(t, mainFrag.Begin(), 1, 1)
		lx.CheckCursor(t, mainFrag.End(), 1, 8)

		// Check elements
		elements := mainFrag.Elements()
		require.Len(t, elements, 3)

		// Check element 1
		elem1 := elements[0]
		require.Equal(t, TestFrFoo, elem1.Kind())
		lx.CheckCursor(t, elem1.Begin(), 1, 1)
		lx.CheckCursor(t, elem1.End(), 1, 4)
		require.Len(t, elem1.Elements(), 1)

		// Check element 2
		elem2 := elements[1]
		require.Equal(t, TestFrSpace, elem2.Kind())
		lx.CheckCursor(t, elem2.Begin(), 1, 4)
		lx.CheckCursor(t, elem2.End(), 1, 5)
		require.Nil(t, elem2.Elements())

		// Check element 3
		elem3 := elements[2]
		require.Equal(t, TestFrBar, elem3.Kind())
		lx.CheckCursor(t, elem3.Begin(), 1, 5)
		lx.CheckCursor(t, elem3.End(), 1, 8)
		require.Len(t, elem3.Elements(), 1)
	})
}

func TestParserChecked(t *testing.T) {
	pr := parser.NewParser()
	lx := NewTestLexer("example")
	expectedKind := parser.FragmentKind(100)
	mainFrag, err := pr.Parse(lx, &parser.Rule{
		Designation: "keyword 'example'",
		Pattern: parser.Checked{"keyword 'example'", func(str string) bool {
			return str == "example"
		}},
		Kind: expectedKind,
	})

	require.NoError(t, err)
	require.NotNil(t, mainFrag)
	require.Equal(t, expectedKind, mainFrag.Kind())
	lx.CheckCursor(t, mainFrag.Begin(), 1, 1)
	lx.CheckCursor(t, mainFrag.End(), 1, 8)

	// Check elements
	elements := mainFrag.Elements()
	require.Len(t, elements, 1)

	// Check element 1
	elem1 := elements[0]
	require.Equal(t, TestFrSeq, elem1.Kind())
	lx.CheckCursor(t, elem1.Begin(), 1, 1)
	lx.CheckCursor(t, elem1.End(), 1, 8)
	require.Nil(t, elem1.Elements())
}

// TestParserCheckedErr tests checked parsing errors
func TestParserCheckedErr(t *testing.T) {
	pr := parser.NewParser()
	lx := NewTestLexer("elpmaxe")
	expectedKind := parser.FragmentKind(100)
	mainFrag, err := pr.Parse(lx, &parser.Rule{
		Designation: "keyword 'example'",
		Pattern: parser.Checked{"keyword 'example'", func(str string) bool {
			return str == "example"
		}},
		Kind: expectedKind,
	})

	require.Error(t, err)
	require.Equal(
		t,
		"unexpected token 'elpmaxe', "+
			"expected {keyword 'example'} at test.txt:1:1",
		err.Error(),
	)
	require.Nil(t, mainFrag)
}

func TestParserZeroOrMore(t *testing.T) {
	t.Run("None", func(t *testing.T) {
		pr := parser.NewParser()
		lx := NewTestLexer("foo")
		expectedKind := parser.FragmentKind(100)
		mainFrag, err := pr.Parse(lx, &parser.Rule{
			Designation: "(space foo)*",
			Pattern: parser.ZeroOrMore{
				parser.Sequence{
					parser.Term(TestFrSpace),
					testR_foo,
				},
			},
			Kind: expectedKind,
		})

		require.NoError(t, err)
		require.NotNil(t, mainFrag)
		require.Len(t, mainFrag.Elements(), 0)
	})

	t.Run("One", func(t *testing.T) {
		pr := parser.NewParser()
		lx := NewTestLexer(" foo ")
		expectedKind := parser.FragmentKind(100)
		mainFrag, err := pr.Parse(lx, &parser.Rule{
			Designation: "(space foo)*",
			Pattern: parser.ZeroOrMore{
				parser.Sequence{
					parser.Term(TestFrSpace),
					testR_foo,
				},
			},
			Kind: expectedKind,
		})
		require.NoError(t, err)
		require.NotNil(t, mainFrag)
		require.Equal(t, expectedKind, mainFrag.Kind())
		lx.CheckCursor(t, mainFrag.Begin(), 1, 1)
		lx.CheckCursor(t, mainFrag.End(), 1, 5)

		// Check elements
		elements := mainFrag.Elements()
		require.Len(t, elements, 2)

		// Check element 1
		elem1 := elements[0]
		require.Equal(t, TestFrSpace, elem1.Kind())
		lx.CheckCursor(t, elem1.Begin(), 1, 1)
		lx.CheckCursor(t, elem1.End(), 1, 2)
		require.Nil(t, elem1.Elements())

		// Check element 2
		elem2 := elements[1]
		require.Equal(t, TestFrFoo, elem2.Kind())
		lx.CheckCursor(t, elem2.Begin(), 1, 2)
		lx.CheckCursor(t, elem2.End(), 1, 5)
		require.Len(t, elem2.Elements(), 1)
	})
}
