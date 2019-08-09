package parser_test

import (
	"testing"

	parser "github.com/romshark/llparser"
	"github.com/stretchr/testify/require"
)

type C struct {
	line   uint
	column uint
}

func checkFrag(
	t *testing.T,
	lexer *TestLexer,
	frag parser.Fragment,
	kind parser.FragmentKind,
	begin C,
	end C,
	elements int,
) {
	require.Equal(t, kind, frag.Kind())
	lexer.CheckCursor(t, frag.Begin(), begin.line, begin.column)
	lexer.CheckCursor(t, frag.End(), end.line, end.column)
	require.Len(t, frag.Elements(), elements)
}

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
		checkFrag(t, lx, mainFrag, expectedKind, C{1, 1}, C{1, 10}, 3)

		// Check elements
		elems := mainFrag.Elements()

		checkFrag(t, lx, elems[0], TestFrSeq, C{1, 1}, C{1, 4}, 0)
		checkFrag(t, lx, elems[1], TestFrSpace, C{1, 4}, C{1, 7}, 0)
		checkFrag(t, lx, elems[2], TestFrSeq, C{1, 7}, C{1, 10}, 0)
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
				testR_bar,
			},
			Kind: expectedKind,
		})
		require.NoError(t, err)
		checkFrag(t, lx, mainFrag, expectedKind, C{1, 1}, C{1, 10}, 3)

		// Check elements
		elems := mainFrag.Elements()

		checkFrag(t, lx, elems[0], TestFrFoo, C{1, 1}, C{1, 4}, 1)
		checkFrag(t, lx, elems[1], TestFrSpace, C{1, 4}, C{1, 7}, 0)
		checkFrag(t, lx, elems[2], TestFrBar, C{1, 7}, C{1, 10}, 1)
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
		checkFrag(t, lx, mainFrag, expectedKind, C{1, 1}, C{1, 4}, 1)

		// Check elements
		elems := mainFrag.Elements()

		checkFrag(t, lx, elems[0], TestFrBar, C{1, 1}, C{1, 4}, 1)
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
		checkFrag(t, lx, mainFrag, expectedKind, C{1, 1}, C{1, 8}, 3)

		// Check elements
		elems := mainFrag.Elements()

		checkFrag(t, lx, elems[0], TestFrFoo, C{1, 1}, C{1, 4}, 1)
		checkFrag(t, lx, elems[1], TestFrSpace, C{1, 4}, C{1, 5}, 0)
		checkFrag(t, lx, elems[2], TestFrBar, C{1, 5}, C{1, 8}, 1)
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
	checkFrag(t, lx, mainFrag, expectedKind, C{1, 1}, C{1, 8}, 1)

	// Check elements
	elems := mainFrag.Elements()

	checkFrag(t, lx, elems[0], TestFrSeq, C{1, 1}, C{1, 8}, 0)
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
			Pattern: parser.Sequence{
				parser.ZeroOrMore{
					parser.Sequence{
						parser.Term(TestFrSpace),
						testR_foo,
					},
				},
				parser.Optional{&parser.Rule{
					Designation: "?foo",
					Pattern:     testR_foo,
					Kind:        200,
				}},
			},
			Kind: expectedKind,
		})

		require.NoError(t, err)
		checkFrag(t, lx, mainFrag, expectedKind, C{1, 1}, C{1, 4}, 1)

		// Check elements
		elems := mainFrag.Elements()

		checkFrag(t, lx, elems[0], 200, C{1, 1}, C{1, 4}, 1)
	})

	t.Run("One", func(t *testing.T) {
		pr := parser.NewParser()
		lx := NewTestLexer(" foo")
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
		checkFrag(t, lx, mainFrag, expectedKind, C{1, 1}, C{1, 5}, 2)

		// Check elements
		elems := mainFrag.Elements()

		checkFrag(t, lx, elems[0], TestFrSpace, C{1, 1}, C{1, 2}, 0)
		checkFrag(t, lx, elems[1], TestFrFoo, C{1, 2}, C{1, 5}, 1)
	})

	t.Run("Multiple", func(t *testing.T) {
		pr := parser.NewParser()
		lx := NewTestLexer(" foo foo foo")
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
		checkFrag(t, lx, mainFrag, expectedKind, C{1, 1}, C{1, 13}, 6)

		// Check elements
		elements := mainFrag.Elements()

		checkFrag(t, lx, elements[0], TestFrSpace, C{1, 1}, C{1, 2}, 0)
		checkFrag(t, lx, elements[1], TestFrFoo, C{1, 2}, C{1, 5}, 1)
		checkFrag(t, lx, elements[2], TestFrSpace, C{1, 5}, C{1, 6}, 0)
		checkFrag(t, lx, elements[3], TestFrFoo, C{1, 6}, C{1, 9}, 1)
		checkFrag(t, lx, elements[4], TestFrSpace, C{1, 9}, C{1, 10}, 0)
		checkFrag(t, lx, elements[5], TestFrFoo, C{1, 10}, C{1, 13}, 1)
	})
}

func TestParserOneOrMore(t *testing.T) {
	t.Run("None", func(t *testing.T) {
		pr := parser.NewParser()
		lx := NewTestLexer("foo")
		expectedKind := parser.FragmentKind(100)
		mainFrag, err := pr.Parse(lx, &parser.Rule{
			Designation: "(space foo)*",
			Pattern: parser.OneOrMore{
				parser.Sequence{
					parser.Term(TestFrSpace),
					testR_foo,
				},
			},
			Kind: expectedKind,
		})

		require.Error(t, err)
		require.Equal(
			t,
			"unexpected token 'foo', "+
				"expected {terminal(1)} at test.txt:1:1",
			err.Error(),
		)
		require.Nil(t, mainFrag)
	})

	t.Run("One", func(t *testing.T) {
		pr := parser.NewParser()
		lx := NewTestLexer(" foo")
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
		checkFrag(t, lx, mainFrag, expectedKind, C{1, 1}, C{1, 5}, 2)

		// Check elements
		elems := mainFrag.Elements()

		checkFrag(t, lx, elems[0], TestFrSpace, C{1, 1}, C{1, 2}, 0)
		checkFrag(t, lx, elems[1], TestFrFoo, C{1, 2}, C{1, 5}, 1)
	})

	t.Run("Multiple", func(t *testing.T) {
		pr := parser.NewParser()
		lx := NewTestLexer(" foo foo foo")
		expectedKind := parser.FragmentKind(100)
		mainFrag, err := pr.Parse(lx, &parser.Rule{
			Designation: "(space foo)*",
			Pattern: parser.OneOrMore{
				parser.Sequence{
					parser.Term(TestFrSpace),
					testR_foo,
				},
			},
			Kind: expectedKind,
		})
		require.NoError(t, err)
		require.NotNil(t, mainFrag)
		checkFrag(t, lx, mainFrag, expectedKind, C{1, 1}, C{1, 13}, 6)

		// Check elements
		elements := mainFrag.Elements()
		checkFrag(t, lx, elements[0], TestFrSpace, C{1, 1}, C{1, 2}, 0)
		checkFrag(t, lx, elements[1], TestFrFoo, C{1, 2}, C{1, 5}, 1)
		checkFrag(t, lx, elements[2], TestFrSpace, C{1, 5}, C{1, 6}, 0)
		checkFrag(t, lx, elements[3], TestFrFoo, C{1, 6}, C{1, 9}, 1)
		checkFrag(t, lx, elements[4], TestFrSpace, C{1, 9}, C{1, 10}, 0)
		checkFrag(t, lx, elements[5], TestFrFoo, C{1, 10}, C{1, 13}, 1)
	})
}

func TestParserSuperfluousInput(t *testing.T) {
	pr := parser.NewParser()
	lx := NewTestLexer("foo ")
	expectedKind := parser.FragmentKind(100)
	mainFrag, err := pr.Parse(lx, &parser.Rule{
		Designation: "single foo",
		Pattern:     testR_foo,
		Kind:        expectedKind,
	})

	require.Error(t, err)
	require.Equal(t, "unexpected token ' ' at test.txt:1:4", err.Error())
	require.Nil(t, mainFrag)
}

func TestParserEither(t *testing.T) {
	t.Run("Neither", func(t *testing.T) {
		pr := parser.NewParser()
		lx := NewTestLexer("  ")
		expectedKind := parser.FragmentKind(100)
		mainFrag, err := pr.Parse(lx, &parser.Rule{
			Designation: "(Foo / Bar)",
			Pattern: parser.Either{
				testR_foo,
				testR_bar,
			},
			Kind: expectedKind,
		})

		require.Error(t, err)
		require.Equal(
			t,
			"unexpected token '  ', expected {either of "+
				"[keyword foo, keyword bar]} at test.txt:1:1",
			err.Error(),
		)
		require.Nil(t, mainFrag)
	})

	t.Run("First", func(t *testing.T) {
		pr := parser.NewParser()
		lx := NewTestLexer("foo")
		expectedKind := parser.FragmentKind(100)
		mainFrag, err := pr.Parse(lx, &parser.Rule{
			Designation: "(Foo / Bar)",
			Pattern: parser.Either{
				testR_foo,
				testR_bar,
			},
			Kind: expectedKind,
		})
		require.NoError(t, err)
		require.NotNil(t, mainFrag)
		checkFrag(t, lx, mainFrag, expectedKind, C{1, 1}, C{1, 4}, 1)

		// Check elements
		elements := mainFrag.Elements()
		require.Len(t, elements, 1)

		checkFrag(t, lx, elements[0], TestFrFoo, C{1, 1}, C{1, 4}, 1)
	})

	t.Run("Second", func(t *testing.T) {
		pr := parser.NewParser()
		lx := NewTestLexer("bar")
		expectedKind := parser.FragmentKind(100)
		mainFrag, err := pr.Parse(lx, &parser.Rule{
			Designation: "(Foo / Bar)",
			Pattern: parser.Either{
				testR_foo,
				testR_bar,
			},
			Kind: expectedKind,
		})
		require.NoError(t, err)
		require.NotNil(t, mainFrag)
		checkFrag(t, lx, mainFrag, expectedKind, C{1, 1}, C{1, 4}, 1)

		// Check elements
		elements := mainFrag.Elements()
		require.Len(t, elements, 1)

		checkFrag(t, lx, elements[0], TestFrBar, C{1, 1}, C{1, 4}, 1)
	})
}
