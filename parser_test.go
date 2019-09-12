package parser_test

import (
	"errors"
	"testing"

	parser "github.com/romshark/llparser"
	"github.com/stretchr/testify/require"
)

type FragKind = parser.FragmentKind

const (
	_ FragKind = iota
	FrSpace
	FrSeparator
	FrWord
	FrFoo
	FrBar
)

func rncmp(a, b []rune) bool {
	for i, x := range b {
		if a[i] != x {
			return false
		}
	}
	return true
}

// Basic terminal types
var (
	termSpace = parser.Lexed{
		Designation: "space",
		Fn: func(crs parser.Cursor) uint {
			switch crs.File.Src[crs.Index] {
			case ' ':
				return 1
			case '\t':
				return 1
			case '\n':
				return 1
			case '\r':
				next := crs.Index + 1
				if next < uint(len(crs.File.Src)) &&
					crs.File.Src[next] == '\n' {
					return 2
				}
			}
			return 0
		},
		Kind: FrSpace,
	}
	termLatinWord = parser.Lexed{
		Fn: func(crs parser.Cursor) uint {
			rn := crs.File.Src[crs.Index]
			if rn >= 48 && rn <= 57 ||
				rn >= 65 && rn <= 90 ||
				rn >= 97 && rn <= 122 {
				return 1
			}
			return 0
		},
		Kind: FrWord,
	}
	termSeparator = parser.TermExact{
		Expectation: []rune(","),
		Kind:        FrSeparator,
	}
	testR_foo = &parser.Rule{
		Designation: "keyword foo",
		Pattern:     parser.TermExact{Expectation: []rune("foo")},
		Kind:        FrFoo,
	}
	testR_bar = &parser.Rule{
		Designation: "keyword bar",
		Pattern:     parser.TermExact{Expectation: []rune("bar")},
		Kind:        FrBar,
	}
)

func newSource(src string) *parser.SourceFile {
	return &parser.SourceFile{
		Name: "test.txt",
		Src:  []rune(src),
	}
}

type C struct {
	line   uint
	column uint
}

// CheckCursor checks a cursor relative to the lexer
func CheckCursor(
	t *testing.T,
	src *parser.SourceFile,
	cursor parser.Cursor,
	line,
	column uint,
) {
	require.Equal(t, src, cursor.File)
	require.Equal(t, line, cursor.Line)
	require.Equal(t, column, cursor.Column)
	if column > 1 || line > 1 {
		require.True(t, cursor.Index > 0)
	} else if column == 1 && line == 1 {
		require.Equal(t, uint(0), cursor.Index)
	}
}

func checkFrag(
	t *testing.T,
	src *parser.SourceFile,
	frag parser.Fragment,
	expectedKind parser.FragmentKind,
	expectedBegin C,
	expectedEnd C,
	expectedElementsNum int,
) {
	require.Equal(t, expectedKind, frag.Kind())
	CheckCursor(t, src, frag.Begin(), expectedBegin.line, expectedBegin.column)
	CheckCursor(t, src, frag.End(), expectedEnd.line, expectedEnd.column)
	require.Len(t, frag.Elements(), expectedElementsNum)
}

func TestParserSequence(t *testing.T) {
	t.Run("SingleLevel", func(t *testing.T) {
		pr := parser.NewParser()
		src := newSource("foo   bar")
		expectedKind := parser.FragmentKind(100)
		mainFrag, err := pr.Parse(src, &parser.Rule{
			Designation: "foobar",
			Pattern: parser.Sequence{
				termLatinWord,
				termSpace,
				termLatinWord,
			},
			Kind: expectedKind,
		})

		require.NoError(t, err)
		checkFrag(t, src, mainFrag, expectedKind, C{1, 1}, C{1, 10}, 3)

		// Check elements
		elems := mainFrag.Elements()

		checkFrag(t, src, elems[0], FrWord, C{1, 1}, C{1, 4}, 0)
		checkFrag(t, src, elems[1], FrSpace, C{1, 4}, C{1, 7}, 0)
		checkFrag(t, src, elems[2], FrWord, C{1, 7}, C{1, 10}, 0)
	})

	t.Run("TwoLevels", func(t *testing.T) {
		pr := parser.NewParser()
		src := newSource("foo   bar")
		expectedKind := parser.FragmentKind(100)
		mainFrag, err := pr.Parse(src, &parser.Rule{
			Designation: "foobar",
			Pattern: parser.Sequence{
				testR_foo,
				termSpace,
				testR_bar,
			},
			Kind: expectedKind,
		})
		require.NoError(t, err)
		checkFrag(t, src, mainFrag, expectedKind, C{1, 1}, C{1, 10}, 3)

		// Check elements
		elems := mainFrag.Elements()

		checkFrag(t, src, elems[0], FrFoo, C{1, 1}, C{1, 4}, 1)
		checkFrag(t, src, elems[1], FrSpace, C{1, 4}, C{1, 7}, 0)
		checkFrag(t, src, elems[2], FrBar, C{1, 7}, C{1, 10}, 1)
	})
}

// TestParserSequenceErr tests sequence parsing errors
func TestParserSequenceErr(t *testing.T) {
	t.Run("UnexpectedTokenTermExact", func(t *testing.T) {
		pr := parser.NewParser()
		src := newSource("foo")
		expectedKind := parser.FragmentKind(100)
		mainFrag, err := pr.Parse(src, &parser.Rule{
			Designation: "foobar",
			Pattern: parser.Sequence{
				parser.TermExact{
					Kind:        FrBar,
					Expectation: []rune("bar"),
				},
				testR_foo,
			},
			Kind: expectedKind,
		})

		require.Error(t, err)
		require.Equal(
			t,
			"unexpected token, expected {'bar'} at test.txt:1:1",
			err.Error(),
		)
		require.Nil(t, mainFrag)
	})
	t.Run("UnexpectedTokenLexed", func(t *testing.T) {
		pr := parser.NewParser()
		src := newSource("foo foo")
		expectedKind := parser.FragmentKind(100)
		mainFrag, err := pr.Parse(src, &parser.Rule{
			Designation: "foobar",
			Pattern: parser.Sequence{
				testR_foo,
				termSpace,
				parser.Lexed{
					Designation: "lexed token",
					Fn: func(crs parser.Cursor) uint {
						if crs.File.Src[crs.Index] == 'b' {
							return 1
						}
						return 0
					},
				},
			},
			Kind: expectedKind,
		})

		require.Error(t, err)
		require.Equal(
			t,
			"unexpected token, expected {lexed token} at test.txt:1:5",
			err.Error(),
		)
		require.Nil(t, mainFrag)
	})
}

func TestParserOptionalInSequence(t *testing.T) {
	t.Run("Missing", func(t *testing.T) {
		pr := parser.NewParser()
		src := newSource("bar")
		expectedKind := parser.FragmentKind(100)
		mainFrag, err := pr.Parse(src, &parser.Rule{
			Designation: "?foo bar",
			Pattern: parser.Sequence{
				parser.Optional{parser.Sequence{
					testR_foo,
					termSpace,
				}},
				testR_bar,
			},
			Kind: expectedKind,
		})

		require.NoError(t, err)
		checkFrag(t, src, mainFrag, expectedKind, C{1, 1}, C{1, 4}, 1)

		// Check elements
		elems := mainFrag.Elements()

		checkFrag(t, src, elems[0], FrBar, C{1, 1}, C{1, 4}, 1)
	})

	t.Run("Present", func(t *testing.T) {
		pr := parser.NewParser()
		src := newSource("foo bar")
		expectedKind := parser.FragmentKind(100)
		mainFrag, err := pr.Parse(src, &parser.Rule{
			Designation: "?foo bar",
			Pattern: parser.Sequence{
				parser.Optional{parser.Sequence{
					testR_foo,
					termSpace,
				}},
				testR_bar,
			},
			Kind: expectedKind,
		})
		require.NoError(t, err)
		checkFrag(t, src, mainFrag, expectedKind, C{1, 1}, C{1, 8}, 3)

		// Check elements
		elems := mainFrag.Elements()

		checkFrag(t, src, elems[0], FrFoo, C{1, 1}, C{1, 4}, 1)
		checkFrag(t, src, elems[1], FrSpace, C{1, 4}, C{1, 5}, 0)
		checkFrag(t, src, elems[2], FrBar, C{1, 5}, C{1, 8}, 1)
	})
}

func TestParserZeroOrMore(t *testing.T) {
	t.Run("None", func(t *testing.T) {
		pr := parser.NewParser()
		src := newSource("foo")
		expectedKind := parser.FragmentKind(100)
		mainFrag, err := pr.Parse(src, &parser.Rule{
			Designation: "(space foo)*",
			Pattern: parser.Sequence{
				parser.ZeroOrMore{
					parser.Sequence{
						termSpace,
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
		checkFrag(t, src, mainFrag, expectedKind, C{1, 1}, C{1, 4}, 1)

		// Check elements
		elems := mainFrag.Elements()

		checkFrag(t, src, elems[0], 200, C{1, 1}, C{1, 4}, 1)
	})

	t.Run("One", func(t *testing.T) {
		pr := parser.NewParser()
		src := newSource(" foo")
		expectedKind := parser.FragmentKind(100)
		mainFrag, err := pr.Parse(src, &parser.Rule{
			Designation: "(space foo)*",
			Pattern: parser.ZeroOrMore{
				parser.Sequence{
					termSpace,
					testR_foo,
				},
			},
			Kind: expectedKind,
		})
		require.NoError(t, err)
		checkFrag(t, src, mainFrag, expectedKind, C{1, 1}, C{1, 5}, 2)

		// Check elements
		elems := mainFrag.Elements()

		checkFrag(t, src, elems[0], FrSpace, C{1, 1}, C{1, 2}, 0)
		checkFrag(t, src, elems[1], FrFoo, C{1, 2}, C{1, 5}, 1)
	})

	t.Run("Multiple", func(t *testing.T) {
		pr := parser.NewParser()
		src := newSource(" foo foo foo")
		expectedKind := parser.FragmentKind(100)
		mainFrag, err := pr.Parse(src, &parser.Rule{
			Designation: "(space foo)*",
			Pattern: parser.ZeroOrMore{
				parser.Sequence{
					termSpace,
					testR_foo,
				},
			},
			Kind: expectedKind,
		})
		require.NoError(t, err)
		checkFrag(t, src, mainFrag, expectedKind, C{1, 1}, C{1, 13}, 6)

		// Check elements
		elements := mainFrag.Elements()

		checkFrag(t, src, elements[0], FrSpace, C{1, 1}, C{1, 2}, 0)
		checkFrag(t, src, elements[1], FrFoo, C{1, 2}, C{1, 5}, 1)
		checkFrag(t, src, elements[2], FrSpace, C{1, 5}, C{1, 6}, 0)
		checkFrag(t, src, elements[3], FrFoo, C{1, 6}, C{1, 9}, 1)
		checkFrag(t, src, elements[4], FrSpace, C{1, 9}, C{1, 10}, 0)
		checkFrag(t, src, elements[5], FrFoo, C{1, 10}, C{1, 13}, 1)
	})
}

func TestParserOneOrMore(t *testing.T) {
	t.Run("None", func(t *testing.T) {
		pr := parser.NewParser()
		src := newSource("foo")
		expectedKind := parser.FragmentKind(100)
		mainFrag, err := pr.Parse(src, &parser.Rule{
			Designation: "(space foo)*",
			Pattern: parser.OneOrMore{
				parser.Sequence{
					termSpace,
					testR_foo,
				},
			},
			Kind: expectedKind,
		})

		require.Error(t, err)

		require.Equal(
			t,
			"unexpected token, expected {space} at test.txt:1:1",
			err.Error(),
		)
		require.Nil(t, mainFrag)
	})

	t.Run("One", func(t *testing.T) {
		pr := parser.NewParser()
		src := newSource(" foo")
		expectedKind := parser.FragmentKind(100)
		mainFrag, err := pr.Parse(src, &parser.Rule{
			Designation: "(space foo)*",
			Pattern: parser.ZeroOrMore{
				parser.Sequence{
					termSpace,
					testR_foo,
				},
			},
			Kind: expectedKind,
		})
		require.NoError(t, err)
		checkFrag(t, src, mainFrag, expectedKind, C{1, 1}, C{1, 5}, 2)

		// Check elements
		elems := mainFrag.Elements()

		checkFrag(t, src, elems[0], FrSpace, C{1, 1}, C{1, 2}, 0)
		checkFrag(t, src, elems[1], FrFoo, C{1, 2}, C{1, 5}, 1)
	})

	t.Run("Multiple", func(t *testing.T) {
		pr := parser.NewParser()
		src := newSource(" foo foo foo")
		expectedKind := parser.FragmentKind(100)
		mainFrag, err := pr.Parse(src, &parser.Rule{
			Designation: "(space foo)*",
			Pattern: parser.OneOrMore{
				parser.Sequence{
					termSpace,
					testR_foo,
				},
			},
			Kind: expectedKind,
		})
		require.NoError(t, err)
		require.NotNil(t, mainFrag)
		checkFrag(t, src, mainFrag, expectedKind, C{1, 1}, C{1, 13}, 6)

		// Check elements
		elements := mainFrag.Elements()
		checkFrag(t, src, elements[0], FrSpace, C{1, 1}, C{1, 2}, 0)
		checkFrag(t, src, elements[1], FrFoo, C{1, 2}, C{1, 5}, 1)
		checkFrag(t, src, elements[2], FrSpace, C{1, 5}, C{1, 6}, 0)
		checkFrag(t, src, elements[3], FrFoo, C{1, 6}, C{1, 9}, 1)
		checkFrag(t, src, elements[4], FrSpace, C{1, 9}, C{1, 10}, 0)
		checkFrag(t, src, elements[5], FrFoo, C{1, 10}, C{1, 13}, 1)
	})
}

func TestParserSuperfluousInput(t *testing.T) {
	pr := parser.NewParser()
	src := newSource("foo ")
	expectedKind := parser.FragmentKind(100)
	mainFrag, err := pr.Parse(src, &parser.Rule{
		Designation: "single foo",
		Pattern:     testR_foo,
		Kind:        expectedKind,
	})

	require.Error(t, err)
	require.Equal(t, "unexpected token at test.txt:1:4", err.Error())
	require.Nil(t, mainFrag)
}

func TestParserEither(t *testing.T) {
	t.Run("Neither", func(t *testing.T) {
		pr := parser.NewParser()
		src := newSource("far")
		expectedKind := parser.FragmentKind(100)
		mainFrag, err := pr.Parse(src, &parser.Rule{
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
			"unexpected token, expected {either of "+
				"[keyword foo, keyword bar]} at test.txt:1:1",
			err.Error(),
		)
		require.Nil(t, mainFrag)
	})

	t.Run("First", func(t *testing.T) {
		pr := parser.NewParser()
		src := newSource("foo")
		expectedKind := parser.FragmentKind(100)
		mainFrag, err := pr.Parse(src, &parser.Rule{
			Designation: "(Foo / Bar)",
			Pattern: parser.Either{
				testR_foo,
				testR_bar,
			},
			Kind: expectedKind,
		})
		require.NoError(t, err)
		require.NotNil(t, mainFrag)
		checkFrag(t, src, mainFrag, expectedKind, C{1, 1}, C{1, 4}, 1)

		// Check elements
		elements := mainFrag.Elements()
		require.Len(t, elements, 1)

		checkFrag(t, src, elements[0], FrFoo, C{1, 1}, C{1, 4}, 1)
	})

	t.Run("Second", func(t *testing.T) {
		pr := parser.NewParser()
		src := newSource("bar")
		expectedKind := parser.FragmentKind(100)
		mainFrag, err := pr.Parse(src, &parser.Rule{
			Designation: "(Foo / Bar)",
			Pattern: parser.Either{
				testR_foo,
				testR_bar,
			},
			Kind: expectedKind,
		})
		require.NoError(t, err)
		require.NotNil(t, mainFrag)
		checkFrag(t, src, mainFrag, expectedKind, C{1, 1}, C{1, 4}, 1)

		// Check elements
		elements := mainFrag.Elements()
		require.Len(t, elements, 1)

		checkFrag(t, src, elements[0], FrBar, C{1, 1}, C{1, 4}, 1)
	})
}

func TestParserRecursiveRule(t *testing.T) {
	pr := parser.NewParser()
	src := newSource("foo,foo,foo,")
	expectedKind := parser.FragmentKind(100)
	recursiveRule := &parser.Rule{
		Designation: "recursive",
		Kind:        expectedKind,
	}
	recursiveRule.Pattern = parser.Sequence{
		testR_foo,
		termSeparator,
		parser.Optional{recursiveRule},
	}
	mainFrag, err := pr.Parse(src, recursiveRule)

	require.NoError(t, err)
	checkFrag(t, src, mainFrag, expectedKind, C{1, 1}, C{1, 13}, 3)

	// First level
	elems := mainFrag.Elements()

	checkFrag(t, src, elems[0], FrFoo, C{1, 1}, C{1, 4}, 1)
	checkFrag(t, src, elems[1], FrSeparator, C{1, 4}, C{1, 5}, 0)
	checkFrag(t, src, elems[2], expectedKind, C{1, 5}, C{1, 13}, 3)

	// Second levels
	elems2 := elems[2].Elements()
	checkFrag(t, src, elems2[0], FrFoo, C{1, 5}, C{1, 8}, 1)
	checkFrag(t, src, elems2[1], FrSeparator, C{1, 8}, C{1, 9}, 0)
	checkFrag(t, src, elems2[2], expectedKind, C{1, 9}, C{1, 13}, 2)

	// Second levels
	elems3 := elems2[2].Elements()
	checkFrag(t, src, elems3[0], FrFoo, C{1, 9}, C{1, 12}, 1)
	checkFrag(t, src, elems3[1], FrSeparator, C{1, 12}, C{1, 13}, 0)
}

func TestParserAction(t *testing.T) {
	pr := parser.NewParser()

	aFrags := make([]parser.Fragment, 0, 2)
	bFrags := make([]parser.Fragment, 0, 2)

	src := newSource("a,b,b,a,")
	aKind := parser.FragmentKind(905)
	ruleA := &parser.Rule{
		Designation: "a",
		Kind:        aKind,
		Pattern:     parser.TermExact{FrWord, []rune("a")},
		Action: func(f parser.Fragment) error {
			aFrags = append(aFrags, f)
			return nil
		},
	}
	bKind := parser.FragmentKind(906)
	ruleB := &parser.Rule{
		Designation: "b",
		Kind:        bKind,
		Pattern:     parser.TermExact{FrWord, []rune("b")},
		Action: func(f parser.Fragment) error {
			bFrags = append(bFrags, f)
			return nil
		},
	}
	mainFrag, err := pr.Parse(src, &parser.Rule{
		Designation: "list",
		Pattern: parser.OneOrMore{&parser.Rule{
			Designation: "list item",
			Pattern: parser.Sequence{
				parser.Either{ruleA, ruleB},
				termSeparator,
			},
		}},
	})

	require.NoError(t, err)
	require.NotNil(t, mainFrag)

	require.Len(t, aFrags, 2)
	checkFrag(t, src, aFrags[0], aKind, C{1, 1}, C{1, 2}, 1)
	checkFrag(t, src, aFrags[1], aKind, C{1, 7}, C{1, 8}, 1)

	require.Len(t, bFrags, 2)
	checkFrag(t, src, bFrags[0], bKind, C{1, 3}, C{1, 4}, 1)
	checkFrag(t, src, bFrags[1], bKind, C{1, 5}, C{1, 6}, 1)
}

func TestParserActionErr(t *testing.T) {
	pr := parser.NewParser()
	src := newSource("a")

	expectedErr := errors.New("custom error")
	mainFrag, err := pr.Parse(src, &parser.Rule{
		Designation: "a",
		Kind:        900,
		Pattern:     parser.TermExact{FrWord, []rune("a")},
		Action: func(f parser.Fragment) error {
			return expectedErr
		},
	})

	require.Error(t, err)
	require.IsType(t, &parser.Err{}, err)
	er := err.(*parser.Err)
	require.Equal(t, expectedErr, er.Err)
	require.Equal(t, uint(0), er.At.Index)
	require.Equal(t, uint(1), er.At.Line)
	require.Equal(t, uint(1), er.At.Column)
	require.Nil(t, mainFrag)
}

func TestParserLexed(t *testing.T) {
	fn := func(crs parser.Cursor) uint {
		rn := crs.File.Src[crs.Index]
		if (rn >= 0x0410 && rn <= 0x044F) || rn == '\n' {
			return 1
		}
		return 0
	}
	expectedKind := parser.FragmentKind(100)

	pr := parser.NewParser()
	src := newSource("абв\nгде")
	mainFrag, err := pr.Parse(src, &parser.Rule{
		Pattern: parser.Lexed{
			Kind:        expectedKind,
			Designation: "lexed token",
			Fn:          fn,
		},
		Kind: expectedKind,
	})

	require.NoError(t, err)
	checkFrag(t, src, mainFrag, expectedKind, C{1, 1}, C{2, 4}, 1)

	// Check elements
	elems := mainFrag.Elements()

	checkFrag(t, src, elems[0], expectedKind, C{1, 1}, C{2, 4}, 0)
}

func TestParserLexedErr(t *testing.T) {
	fn := func(crs parser.Cursor) uint {
		rn := crs.File.Src[crs.Index]
		if (rn >= 0x0410 && rn <= 0x044F) || rn == '\n' {
			return 1
		}
		return 0
	}
	expectedKind := parser.FragmentKind(100)

	pr := parser.NewParser()
	src := newSource("abc")
	mainFrag, err := pr.Parse(src, &parser.Rule{
		Pattern: parser.Lexed{
			Kind:        expectedKind,
			Designation: "lexed token",
			Fn:          fn,
		},
		Kind: expectedKind,
	})

	require.Error(t, err)
	require.Nil(t, mainFrag)
}
