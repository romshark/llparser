package parser_test

import (
	"errors"
	"testing"

	parser "github.com/romshark/llparser"
	"github.com/romshark/llparser/misc"
	"github.com/stretchr/testify/require"
)

type FragKind = parser.FragmentKind

const (
	_ FragKind = misc.FrSign + iota
	TestFrFoo
	TestFrBar
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
	testR_foo = &parser.Rule{
		Designation: "keyword foo",
		Pattern: parser.Checked{
			Designation: "keyword foo",
			Fn: func(src []rune) bool {
				return rncmp(src, []rune("foo"))
			},
		},
		Kind: TestFrFoo,
	}
	testR_bar = &parser.Rule{
		Designation: "keyword bar",
		Pattern: parser.Checked{
			Designation: "keyword bar",
			Fn: func(src []rune) bool {
				return rncmp(src, []rune("bar"))
			},
		},
		Kind: TestFrBar,
	}
)

func newLexer(src string) parser.Lexer {
	return misc.NewLexer(&parser.SourceFile{
		Name: "test.txt",
		Src:  []rune(src),
	})
}

type C struct {
	line   uint
	column uint
}

// CheckCursor checks a cursor relative to the lexer
func CheckCursor(
	t *testing.T,
	lexer parser.Lexer,
	cursor parser.Cursor,
	line,
	column uint,
) {
	require.Equal(t, lexer.Position().File, cursor.File)
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
	lexer parser.Lexer,
	frag parser.Fragment,
	kind parser.FragmentKind,
	begin C,
	end C,
	elements int,
) {
	require.Equal(t, kind, frag.Kind())
	CheckCursor(t, lexer, frag.Begin(), begin.line, begin.column)
	CheckCursor(t, lexer, frag.End(), end.line, end.column)
	require.Len(t, frag.Elements(), elements)
}

func TestTokenString(t *testing.T) {
	lx := newLexer("abcdefg")
	tk, matched, err := lx.ReadExact([]rune("abc"), parser.FragmentKind(100))

	require.NoError(t, err)
	require.True(t, matched)
	require.Equal(t, "100(test.txt: 1:1-1:4 'abc')", tk.String())

	tk, matched, err = lx.ReadExact([]rune("defg"), parser.FragmentKind(101))

	require.NoError(t, err)
	require.True(t, matched)
	require.Equal(t, "101(test.txt: 1:4-1:8 'defg')", tk.String())
}

func TestParserSequence(t *testing.T) {
	t.Run("SingleLevel", func(t *testing.T) {
		pr := parser.NewParser()
		lx := newLexer("foo   bar")
		expectedKind := parser.FragmentKind(100)
		mainFrag, err := pr.Parse(lx, &parser.Rule{
			Designation: "foobar",
			Pattern: parser.Sequence{
				parser.Term(misc.FrWord),
				parser.Term(misc.FrSpace),
				parser.Term(misc.FrWord),
			},
			Kind: expectedKind,
		})

		require.NoError(t, err)
		checkFrag(t, lx, mainFrag, expectedKind, C{1, 1}, C{1, 10}, 3)

		// Check elements
		elems := mainFrag.Elements()

		checkFrag(t, lx, elems[0], misc.FrWord, C{1, 1}, C{1, 4}, 0)
		checkFrag(t, lx, elems[1], misc.FrSpace, C{1, 4}, C{1, 7}, 0)
		checkFrag(t, lx, elems[2], misc.FrWord, C{1, 7}, C{1, 10}, 0)
	})

	t.Run("TwoLevels", func(t *testing.T) {
		pr := parser.NewParser()
		lx := newLexer("foo   bar")
		expectedKind := parser.FragmentKind(100)
		mainFrag, err := pr.Parse(lx, &parser.Rule{
			Designation: "foobar",
			Pattern: parser.Sequence{
				testR_foo,
				parser.Term(misc.FrSpace),
				testR_bar,
			},
			Kind: expectedKind,
		})
		require.NoError(t, err)
		checkFrag(t, lx, mainFrag, expectedKind, C{1, 1}, C{1, 10}, 3)

		// Check elements
		elems := mainFrag.Elements()

		checkFrag(t, lx, elems[0], TestFrFoo, C{1, 1}, C{1, 4}, 1)
		checkFrag(t, lx, elems[1], misc.FrSpace, C{1, 4}, C{1, 7}, 0)
		checkFrag(t, lx, elems[2], TestFrBar, C{1, 7}, C{1, 10}, 1)
	})
}

// TestParserSequenceErr tests sequence parsing errors
func TestParserSequenceErr(t *testing.T) {
	t.Run("UnexpectedTokenTermExact", func(t *testing.T) {
		pr := parser.NewParser()
		lx := newLexer("foo")
		expectedKind := parser.FragmentKind(100)
		mainFrag, err := pr.Parse(lx, &parser.Rule{
			Designation: "foobar",
			Pattern: parser.Sequence{
				parser.TermExact{
					Kind:        TestFrBar,
					Expectation: []rune("bar"),
				},
				testR_foo,
			},
			Kind: expectedKind,
		})

		require.Error(t, err)
		require.Equal(
			t,
			"unexpected token 'f', expected {'bar'} at test.txt:1:1",
			err.Error(),
		)
		require.Nil(t, mainFrag)
	})
	t.Run("UnexpectedTokenChecked", func(t *testing.T) {
		pr := parser.NewParser()
		lx := newLexer("foo foo")
		expectedKind := parser.FragmentKind(100)
		mainFrag, err := pr.Parse(lx, &parser.Rule{
			Designation: "foobar",
			Pattern: parser.Sequence{
				testR_foo,
				parser.Term(misc.FrSpace),
				parser.Checked{
					Designation: "checked token",
					Fn: func(str []rune) bool {
						return rncmp(str, []rune("bar"))
					},
				},
			},
			Kind: expectedKind,
		})

		require.Error(t, err)
		require.Equal(
			t,
			"unexpected token 'foo', expected {checked token} at test.txt:1:5",
			err.Error(),
		)
		require.Nil(t, mainFrag)
	})
}

func TestParserOptionalInSequence(t *testing.T) {
	t.Run("Missing", func(t *testing.T) {
		pr := parser.NewParser()
		lx := newLexer("bar")
		expectedKind := parser.FragmentKind(100)
		mainFrag, err := pr.Parse(lx, &parser.Rule{
			Designation: "?foo bar",
			Pattern: parser.Sequence{
				parser.Optional{parser.Sequence{
					testR_foo,
					parser.Term(misc.FrSpace),
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
		lx := newLexer("foo bar")
		expectedKind := parser.FragmentKind(100)
		mainFrag, err := pr.Parse(lx, &parser.Rule{
			Designation: "?foo bar",
			Pattern: parser.Sequence{
				parser.Optional{parser.Sequence{
					testR_foo,
					parser.Term(misc.FrSpace),
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
		checkFrag(t, lx, elems[1], misc.FrSpace, C{1, 4}, C{1, 5}, 0)
		checkFrag(t, lx, elems[2], TestFrBar, C{1, 5}, C{1, 8}, 1)
	})
}

func TestParserChecked(t *testing.T) {
	pr := parser.NewParser()
	lx := newLexer("example")
	expectedKind := parser.FragmentKind(100)
	mainFrag, err := pr.Parse(lx, &parser.Rule{
		Designation: "keyword 'example'",
		Pattern: parser.Checked{"keyword 'example'", func(str []rune) bool {
			return rncmp(str, []rune("example"))
		}},
		Kind: expectedKind,
	})

	require.NoError(t, err)
	checkFrag(t, lx, mainFrag, expectedKind, C{1, 1}, C{1, 8}, 1)

	// Check elements
	elems := mainFrag.Elements()

	checkFrag(t, lx, elems[0], misc.FrWord, C{1, 1}, C{1, 8}, 0)
}

// TestParserCheckedErr tests checked parsing errors
func TestParserCheckedErr(t *testing.T) {
	pr := parser.NewParser()
	lx := newLexer("elpmaxe")
	expectedKind := parser.FragmentKind(100)
	mainFrag, err := pr.Parse(lx, &parser.Rule{
		Designation: "keyword 'example'",
		Pattern: parser.Checked{"keyword 'example'", func(str []rune) bool {
			return rncmp(str, []rune("example"))
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
		lx := newLexer("foo")
		expectedKind := parser.FragmentKind(100)
		mainFrag, err := pr.Parse(lx, &parser.Rule{
			Designation: "(space foo)*",
			Pattern: parser.Sequence{
				parser.ZeroOrMore{
					parser.Sequence{
						parser.Term(misc.FrSpace),
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
		lx := newLexer(" foo")
		expectedKind := parser.FragmentKind(100)
		mainFrag, err := pr.Parse(lx, &parser.Rule{
			Designation: "(space foo)*",
			Pattern: parser.ZeroOrMore{
				parser.Sequence{
					parser.Term(misc.FrSpace),
					testR_foo,
				},
			},
			Kind: expectedKind,
		})
		require.NoError(t, err)
		checkFrag(t, lx, mainFrag, expectedKind, C{1, 1}, C{1, 5}, 2)

		// Check elements
		elems := mainFrag.Elements()

		checkFrag(t, lx, elems[0], misc.FrSpace, C{1, 1}, C{1, 2}, 0)
		checkFrag(t, lx, elems[1], TestFrFoo, C{1, 2}, C{1, 5}, 1)
	})

	t.Run("Multiple", func(t *testing.T) {
		pr := parser.NewParser()
		lx := newLexer(" foo foo foo")
		expectedKind := parser.FragmentKind(100)
		mainFrag, err := pr.Parse(lx, &parser.Rule{
			Designation: "(space foo)*",
			Pattern: parser.ZeroOrMore{
				parser.Sequence{
					parser.Term(misc.FrSpace),
					testR_foo,
				},
			},
			Kind: expectedKind,
		})
		require.NoError(t, err)
		checkFrag(t, lx, mainFrag, expectedKind, C{1, 1}, C{1, 13}, 6)

		// Check elements
		elements := mainFrag.Elements()

		checkFrag(t, lx, elements[0], misc.FrSpace, C{1, 1}, C{1, 2}, 0)
		checkFrag(t, lx, elements[1], TestFrFoo, C{1, 2}, C{1, 5}, 1)
		checkFrag(t, lx, elements[2], misc.FrSpace, C{1, 5}, C{1, 6}, 0)
		checkFrag(t, lx, elements[3], TestFrFoo, C{1, 6}, C{1, 9}, 1)
		checkFrag(t, lx, elements[4], misc.FrSpace, C{1, 9}, C{1, 10}, 0)
		checkFrag(t, lx, elements[5], TestFrFoo, C{1, 10}, C{1, 13}, 1)
	})
}

func TestParserOneOrMore(t *testing.T) {
	t.Run("None", func(t *testing.T) {
		pr := parser.NewParser()
		lx := newLexer("foo")
		expectedKind := parser.FragmentKind(100)
		mainFrag, err := pr.Parse(lx, &parser.Rule{
			Designation: "(space foo)*",
			Pattern: parser.OneOrMore{
				parser.Sequence{
					parser.Term(misc.FrSpace),
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
		lx := newLexer(" foo")
		expectedKind := parser.FragmentKind(100)
		mainFrag, err := pr.Parse(lx, &parser.Rule{
			Designation: "(space foo)*",
			Pattern: parser.ZeroOrMore{
				parser.Sequence{
					parser.Term(misc.FrSpace),
					testR_foo,
				},
			},
			Kind: expectedKind,
		})
		require.NoError(t, err)
		checkFrag(t, lx, mainFrag, expectedKind, C{1, 1}, C{1, 5}, 2)

		// Check elements
		elems := mainFrag.Elements()

		checkFrag(t, lx, elems[0], misc.FrSpace, C{1, 1}, C{1, 2}, 0)
		checkFrag(t, lx, elems[1], TestFrFoo, C{1, 2}, C{1, 5}, 1)
	})

	t.Run("Multiple", func(t *testing.T) {
		pr := parser.NewParser()
		lx := newLexer(" foo foo foo")
		expectedKind := parser.FragmentKind(100)
		mainFrag, err := pr.Parse(lx, &parser.Rule{
			Designation: "(space foo)*",
			Pattern: parser.OneOrMore{
				parser.Sequence{
					parser.Term(misc.FrSpace),
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
		checkFrag(t, lx, elements[0], misc.FrSpace, C{1, 1}, C{1, 2}, 0)
		checkFrag(t, lx, elements[1], TestFrFoo, C{1, 2}, C{1, 5}, 1)
		checkFrag(t, lx, elements[2], misc.FrSpace, C{1, 5}, C{1, 6}, 0)
		checkFrag(t, lx, elements[3], TestFrFoo, C{1, 6}, C{1, 9}, 1)
		checkFrag(t, lx, elements[4], misc.FrSpace, C{1, 9}, C{1, 10}, 0)
		checkFrag(t, lx, elements[5], TestFrFoo, C{1, 10}, C{1, 13}, 1)
	})
}

func TestParserSuperfluousInput(t *testing.T) {
	pr := parser.NewParser()
	lx := newLexer("foo ")
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
		lx := newLexer("  ")
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
		lx := newLexer("foo")
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
		lx := newLexer("bar")
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

func TestParserRecursiveRule(t *testing.T) {
	pr := parser.NewParser()
	lx := newLexer("foo,foo,foo,")
	expectedKind := parser.FragmentKind(100)
	recursiveRule := &parser.Rule{
		Designation: "recursive",
		Kind:        expectedKind,
	}
	recursiveRule.Pattern = parser.Sequence{
		testR_foo,
		parser.Term(misc.FrSign),
		parser.Optional{recursiveRule},
	}
	mainFrag, err := pr.Parse(lx, recursiveRule)

	require.NoError(t, err)
	checkFrag(t, lx, mainFrag, expectedKind, C{1, 1}, C{1, 13}, 3)

	// First level
	elems := mainFrag.Elements()

	checkFrag(t, lx, elems[0], TestFrFoo, C{1, 1}, C{1, 4}, 1)
	checkFrag(t, lx, elems[1], misc.FrSign, C{1, 4}, C{1, 5}, 0)
	checkFrag(t, lx, elems[2], expectedKind, C{1, 5}, C{1, 13}, 3)

	// Second levels
	elems2 := elems[2].Elements()
	checkFrag(t, lx, elems2[0], TestFrFoo, C{1, 5}, C{1, 8}, 1)
	checkFrag(t, lx, elems2[1], misc.FrSign, C{1, 8}, C{1, 9}, 0)
	checkFrag(t, lx, elems2[2], expectedKind, C{1, 9}, C{1, 13}, 2)

	// Second levels
	elems3 := elems2[2].Elements()
	checkFrag(t, lx, elems3[0], TestFrFoo, C{1, 9}, C{1, 12}, 1)
	checkFrag(t, lx, elems3[1], misc.FrSign, C{1, 12}, C{1, 13}, 0)
}

func TestParserAction(t *testing.T) {
	pr := parser.NewParser()

	aFrags := make([]parser.Fragment, 0, 2)
	bFrags := make([]parser.Fragment, 0, 2)

	lx := newLexer("a,b,b,a,")
	aKind := parser.FragmentKind(905)
	ruleA := &parser.Rule{
		Designation: "a",
		Kind:        aKind,
		Pattern:     parser.TermExact{misc.FrWord, []rune("a")},
		Action: func(f parser.Fragment) error {
			aFrags = append(aFrags, f)
			return nil
		},
	}
	bKind := parser.FragmentKind(906)
	ruleB := &parser.Rule{
		Designation: "b",
		Kind:        bKind,
		Pattern:     parser.TermExact{misc.FrWord, []rune("b")},
		Action: func(f parser.Fragment) error {
			bFrags = append(bFrags, f)
			return nil
		},
	}
	mainFrag, err := pr.Parse(lx, &parser.Rule{
		Designation: "list",
		Pattern: parser.OneOrMore{&parser.Rule{
			Designation: "list item",
			Pattern: parser.Sequence{
				parser.Either{ruleA, ruleB},
				parser.Term(misc.FrSign),
			},
		}},
	})

	require.NoError(t, err)
	require.NotNil(t, mainFrag)

	require.Len(t, aFrags, 2)
	checkFrag(t, lx, aFrags[0], aKind, C{1, 1}, C{1, 2}, 1)
	checkFrag(t, lx, aFrags[1], aKind, C{1, 7}, C{1, 8}, 1)

	require.Len(t, bFrags, 2)
	checkFrag(t, lx, bFrags[0], bKind, C{1, 3}, C{1, 4}, 1)
	checkFrag(t, lx, bFrags[1], bKind, C{1, 5}, C{1, 6}, 1)
}

func TestParserActionErr(t *testing.T) {
	pr := parser.NewParser()
	lx := newLexer("a")

	expectedErr := errors.New("custom error")
	mainFrag, err := pr.Parse(lx, &parser.Rule{
		Designation: "a",
		Kind:        900,
		Pattern:     parser.TermExact{misc.FrWord, []rune("a")},
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
	lx := newLexer("абв\nгде")
	mainFrag, err := pr.Parse(lx, &parser.Rule{
		Pattern: parser.Lexed{
			Kind:        expectedKind,
			Designation: "lexed token",
			Fn:          fn,
		},
		Kind: expectedKind,
	})

	require.NoError(t, err)
	checkFrag(t, lx, mainFrag, expectedKind, C{1, 1}, C{2, 4}, 1)

	// Check elements
	elems := mainFrag.Elements()

	checkFrag(t, lx, elems[0], expectedKind, C{1, 1}, C{2, 4}, 0)
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
	lx := newLexer("abc")
	mainFrag, err := pr.Parse(lx, &parser.Rule{
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
