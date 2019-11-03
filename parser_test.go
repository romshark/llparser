package parser_test

import (
	"errors"
	"fmt"
	"testing"

	llp "github.com/romshark/llparser"
	"github.com/stretchr/testify/require"
)

type FragKind = llp.FragmentKind

const (
	_ FragKind = iota
	FrSpace
	FrSeparator
	FrWord
	FrFoo
	FrBar
)

// Basic terminal types
var (
	termSpace = &llp.Lexed{
		Designation: "space",
		Fn: func(_ uint, crs llp.Cursor) bool {
			switch crs.File.Src[crs.Index] {
			case ' ':
				return true
			case '\t':
				return true
			case '\n':
				return true
			case '\r':
				return true
			}
			return false
		},
		Kind: FrSpace,
	}
	termLatinWord = &llp.Lexed{
		Designation: "latin word",
		Fn: func(_ uint, crs llp.Cursor) bool {
			rn := crs.File.Src[crs.Index]
			if rn >= 48 && rn <= 57 ||
				rn >= 65 && rn <= 90 ||
				rn >= 97 && rn <= 122 {
				return true
			}
			return false
		},
		Kind: FrWord,
	}
	termSeparator = &llp.Exact{
		Expectation: []rune(","),
		Kind:        FrSeparator,
	}
	testR_foo = &llp.Rule{
		Designation: "keyword foo",
		Pattern:     &llp.Exact{Expectation: []rune("foo")},
		Kind:        FrFoo,
	}
	testR_bar = &llp.Rule{
		Designation: "keyword bar",
		Pattern:     &llp.Exact{Expectation: []rune("bar")},
		Kind:        FrBar,
	}
)

func newSource(src string) *llp.SourceFile {
	return &llp.SourceFile{
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
	src *llp.SourceFile,
	cursor llp.Cursor,
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
	src *llp.SourceFile,
	frag llp.Fragment,
	expectedKind llp.FragmentKind,
	expectedBegin C,
	expectedEnd C,
	expectedElementsNum int,
) {
	require.Equal(t, expectedKind, frag.Kind())
	CheckCursor(t, src, frag.Begin(), expectedBegin.line, expectedBegin.column)
	CheckCursor(t, src, frag.End(), expectedEnd.line, expectedEnd.column)
	require.Len(t, frag.Elements(), expectedElementsNum)
}

func newParser(t *testing.T, grammar, errGrammar *llp.Rule) *llp.Parser {
	pr, err := llp.NewParser(grammar, errGrammar)
	require.NoError(t, err)
	return pr
}

func TestParserSequence(t *testing.T) {
	t.Run("SingleLevel", func(t *testing.T) {
		expectedKind := llp.FragmentKind(100)
		pr := newParser(t, &llp.Rule{
			Designation: "foobar",
			Pattern: llp.Sequence{
				termLatinWord,
				termSpace,
				termLatinWord,
			},
			Kind: expectedKind,
		}, nil)

		src := newSource("foo   bar")
		mainFrag, err := pr.Parse(src)

		require.NoError(t, err)
		checkFrag(t, src, mainFrag, expectedKind, C{1, 1}, C{1, 10}, 3)

		// Check elements
		elems := mainFrag.Elements()

		checkFrag(t, src, elems[0], FrWord, C{1, 1}, C{1, 4}, 0)
		checkFrag(t, src, elems[1], FrSpace, C{1, 4}, C{1, 7}, 0)
		checkFrag(t, src, elems[2], FrWord, C{1, 7}, C{1, 10}, 0)
	})

	t.Run("TwoLevels", func(t *testing.T) {
		expectedKind := llp.FragmentKind(100)
		pr := newParser(t, &llp.Rule{
			Designation: "foobar",
			Pattern: llp.Sequence{
				testR_foo,
				termSpace,
				testR_bar,
			},
			Kind: expectedKind,
		}, nil)

		src := newSource("foo   bar")
		mainFrag, err := pr.Parse(src)
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
		expectedKind := llp.FragmentKind(100)
		pr := newParser(t, &llp.Rule{
			Designation: "foobar",
			Pattern: llp.Sequence{
				&llp.Exact{
					Kind:        FrBar,
					Expectation: []rune("bar"),
				},
				testR_foo,
			},
			Kind: expectedKind,
		}, nil)

		src := newSource("foo")
		mainFrag, err := pr.Parse(src)
		require.Error(t, err)
		require.Equal(
			t,
			"unexpected token, expected {'bar'} at test.txt:1:1",
			err.Error(),
		)
		require.Nil(t, mainFrag)
	})
	t.Run("UnexpectedTokenLexed", func(t *testing.T) {
		expectedKind := llp.FragmentKind(100)
		pr := newParser(t, &llp.Rule{
			Designation: "foobar",
			Pattern: llp.Sequence{
				testR_foo,
				termSpace,
				&llp.Lexed{
					Designation: "lexed token",
					Fn: func(_ uint, crs llp.Cursor) bool {
						return crs.File.Src[crs.Index] == 'b'
					},
				},
			},
			Kind: expectedKind,
		}, nil)

		src := newSource("foo foo")
		mainFrag, err := pr.Parse(src)

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
		expectedKind := llp.FragmentKind(100)
		pr := newParser(t, &llp.Rule{
			Designation: "?foo bar",
			Pattern: llp.Sequence{
				&llp.Repeated{
					Min: 0,
					Max: 1,
					Pattern: llp.Sequence{
						testR_foo,
						termSpace,
					},
				},
				testR_bar,
			},
			Kind: expectedKind,
		}, nil)

		src := newSource("bar")
		mainFrag, err := pr.Parse(src)

		require.NoError(t, err)
		checkFrag(t, src, mainFrag, expectedKind, C{1, 1}, C{1, 4}, 1)

		// Check elements
		elems := mainFrag.Elements()

		checkFrag(t, src, elems[0], FrBar, C{1, 1}, C{1, 4}, 1)
	})

	t.Run("Present", func(t *testing.T) {
		expectedKind := llp.FragmentKind(100)
		pr := newParser(t, &llp.Rule{
			Designation: "?foo bar",
			Pattern: llp.Sequence{
				&llp.Repeated{
					Min: 0,
					Max: 1,
					Pattern: llp.Sequence{
						testR_foo,
						termSpace,
					},
				},
				testR_bar,
			},
			Kind: expectedKind,
		}, nil)

		src := newSource("foo bar")
		mainFrag, err := pr.Parse(src)
		require.NoError(t, err)
		checkFrag(t, src, mainFrag, expectedKind, C{1, 1}, C{1, 8}, 3)

		// Check elements
		elems := mainFrag.Elements()

		checkFrag(t, src, elems[0], FrFoo, C{1, 1}, C{1, 4}, 1)
		checkFrag(t, src, elems[1], FrSpace, C{1, 4}, C{1, 5}, 0)
		checkFrag(t, src, elems[2], FrBar, C{1, 5}, C{1, 8}, 1)
	})
}

func TestParserRepeatedZeroOrMany(t *testing.T) {
	t.Run("None", func(t *testing.T) {
		expectedKind := llp.FragmentKind(100)
		pr := newParser(t, &llp.Rule{
			Designation: "(space foo)*",
			Pattern: llp.Sequence{
				&llp.Repeated{
					Pattern: llp.Sequence{
						termSpace,
						testR_foo,
					},
				},
				&llp.Repeated{
					Min: 0,
					Max: 1,
					Pattern: &llp.Rule{
						Designation: "?foo",
						Pattern:     testR_foo,
						Kind:        200,
					},
				},
			},
			Kind: expectedKind,
		}, nil)

		src := newSource("foo")
		mainFrag, err := pr.Parse(src)
		require.NoError(t, err)
		checkFrag(t, src, mainFrag, expectedKind, C{1, 1}, C{1, 4}, 1)

		// Check elements
		elems := mainFrag.Elements()

		checkFrag(t, src, elems[0], 200, C{1, 1}, C{1, 4}, 1)
	})

	t.Run("One", func(t *testing.T) {
		expectedKind := llp.FragmentKind(100)
		pr := newParser(t, &llp.Rule{
			Designation: "(space foo)*",
			Pattern: &llp.Repeated{
				Pattern: llp.Sequence{
					termSpace,
					testR_foo,
				},
			},
			Kind: expectedKind,
		}, nil)

		src := newSource(" foo")
		mainFrag, err := pr.Parse(src)
		require.NoError(t, err)
		checkFrag(t, src, mainFrag, expectedKind, C{1, 1}, C{1, 5}, 2)

		// Check elements
		elems := mainFrag.Elements()

		checkFrag(t, src, elems[0], FrSpace, C{1, 1}, C{1, 2}, 0)
		checkFrag(t, src, elems[1], FrFoo, C{1, 2}, C{1, 5}, 1)
	})

	t.Run("Multiple", func(t *testing.T) {
		expectedKind := llp.FragmentKind(100)
		pr := newParser(t, &llp.Rule{
			Designation: "(space foo)*",
			Pattern: &llp.Repeated{
				Pattern: llp.Sequence{
					termSpace,
					testR_foo,
				},
			},
			Kind: expectedKind,
		}, nil)

		src := newSource(" foo foo foo")
		mainFrag, err := pr.Parse(src)
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

func TestParserRepeatedMin1(t *testing.T) {
	expectedKind := llp.FragmentKind(100)
	pr := newParser(t, &llp.Rule{
		Designation: "foo{1,}",
		Pattern: &llp.Repeated{
			Min:     1,
			Pattern: testR_foo,
		},
		Kind: expectedKind,
	}, nil)

	t.Run("None", func(t *testing.T) {
		src := newSource("bar")
		mainFrag, err := pr.Parse(src)

		require.Error(t, err)
		require.Equal(
			t,
			"unexpected token, expected {keyword foo} at test.txt:1:1",
			err.Error(),
		)
		require.Nil(t, mainFrag)
	})

	t.Run("One", func(t *testing.T) {
		src := newSource("foo")
		mainFrag, err := pr.Parse(src)
		require.NoError(t, err)
		checkFrag(t, src, mainFrag, expectedKind, C{1, 1}, C{1, 4}, 1)

		// Check elements
		elements := mainFrag.Elements()
		checkFrag(t, src, elements[0], FrFoo, C{1, 1}, C{1, 4}, 1)
	})

	t.Run("Multiple(3)", func(t *testing.T) {
		src := newSource("foofoofoo")
		mainFrag, err := pr.Parse(src)
		require.NoError(t, err)
		require.NotNil(t, mainFrag)
		checkFrag(t, src, mainFrag, expectedKind, C{1, 1}, C{1, 10}, 3)

		// Check elements
		elements := mainFrag.Elements()
		checkFrag(t, src, elements[0], FrFoo, C{1, 1}, C{1, 4}, 1)
		checkFrag(t, src, elements[1], FrFoo, C{1, 4}, C{1, 7}, 1)
		checkFrag(t, src, elements[2], FrFoo, C{1, 7}, C{1, 10}, 1)
	})
}

func TestParserRepeatedMin2(t *testing.T) {
	expectedKind := llp.FragmentKind(100)
	pr := newParser(t, &llp.Rule{
		Designation: "foo{2,}",
		Pattern: &llp.Repeated{
			Min:     2,
			Pattern: testR_foo,
		},
		Kind: expectedKind,
	}, nil)

	t.Run("None", func(t *testing.T) {
		src := newSource("bar")
		mainFrag, err := pr.Parse(src)

		require.Error(t, err)
		require.Equal(
			t,
			"unexpected token, expected {keyword foo} at test.txt:1:1",
			err.Error(),
		)
		require.Nil(t, mainFrag)
	})

	t.Run("One", func(t *testing.T) {
		src := newSource("foo")
		mainFrag, err := pr.Parse(src)

		require.Error(t, err)
		require.Equal(
			t,
			"unexpected token, expected {keyword foo} at test.txt:1:4",
			err.Error(),
		)
		require.Nil(t, mainFrag)
	})

	t.Run("Multiple(min)", func(t *testing.T) {
		src := newSource("foofoo")
		mainFrag, err := pr.Parse(src)
		require.NoError(t, err)
		require.NotNil(t, mainFrag)
		checkFrag(t, src, mainFrag, expectedKind, C{1, 1}, C{1, 7}, 2)

		// Check elements
		elements := mainFrag.Elements()
		checkFrag(t, src, elements[0], FrFoo, C{1, 1}, C{1, 4}, 1)
		checkFrag(t, src, elements[1], FrFoo, C{1, 4}, C{1, 7}, 1)
	})

	t.Run("Multiple(3)", func(t *testing.T) {
		src := newSource("foofoofoo")
		mainFrag, err := pr.Parse(src)
		require.NoError(t, err)
		require.NotNil(t, mainFrag)
		checkFrag(t, src, mainFrag, expectedKind, C{1, 1}, C{1, 10}, 3)

		// Check elements
		elements := mainFrag.Elements()
		checkFrag(t, src, elements[0], FrFoo, C{1, 1}, C{1, 4}, 1)
		checkFrag(t, src, elements[1], FrFoo, C{1, 4}, C{1, 7}, 1)
		checkFrag(t, src, elements[2], FrFoo, C{1, 7}, C{1, 10}, 1)
	})
}

func TestParserRepeatedMin1Max2(t *testing.T) {
	expectedKind := llp.FragmentKind(100)
	pr := newParser(t, &llp.Rule{
		Designation: "foo{1,2}",
		Pattern: &llp.Repeated{
			Min:     1,
			Max:     2,
			Pattern: testR_foo,
		},
		Kind: expectedKind,
	}, nil)

	t.Run("None", func(t *testing.T) {
		src := newSource("bar")
		mainFrag, err := pr.Parse(src)

		require.Error(t, err)
		require.Equal(
			t,
			"unexpected token, expected {keyword foo} at test.txt:1:1",
			err.Error(),
		)
		require.Nil(t, mainFrag)
	})

	t.Run("One", func(t *testing.T) {
		src := newSource("foo")
		mainFrag, err := pr.Parse(src)
		require.NoError(t, err)
		require.NotNil(t, mainFrag)
		checkFrag(t, src, mainFrag, expectedKind, C{1, 1}, C{1, 4}, 1)

		// Check elements
		elements := mainFrag.Elements()
		checkFrag(t, src, elements[0], FrFoo, C{1, 1}, C{1, 4}, 1)
	})

	t.Run("Multiple(max)", func(t *testing.T) {
		src := newSource("foofoo")
		mainFrag, err := pr.Parse(src)
		require.NoError(t, err)
		require.NotNil(t, mainFrag)
		checkFrag(t, src, mainFrag, expectedKind, C{1, 1}, C{1, 7}, 2)

		// Check elements
		elements := mainFrag.Elements()
		checkFrag(t, src, elements[0], FrFoo, C{1, 1}, C{1, 4}, 1)
		checkFrag(t, src, elements[1], FrFoo, C{1, 4}, C{1, 7}, 1)
	})

	t.Run("Multiple(max+1)", func(t *testing.T) {
		src := newSource("foofoofoo")
		mainFrag, err := pr.Parse(src)

		require.Error(t, err)
		require.Equal(t, "unexpected token at test.txt:1:7", err.Error())
		require.Nil(t, mainFrag)
	})
}

func TestParserRepeatedOptional(t *testing.T) {
	expectedKind := llp.FragmentKind(100)
	pr := newParser(t, &llp.Rule{
		Designation: "foo? bar?",
		Pattern: llp.Sequence{
			&llp.Repeated{
				Min:     0,
				Max:     1,
				Pattern: testR_foo,
			},
			&llp.Repeated{
				Min:     0,
				Max:     1,
				Pattern: testR_bar,
			},
		},
		Kind: expectedKind,
	}, nil)

	t.Run("None", func(t *testing.T) {
		src := newSource("")
		mainFrag, err := pr.Parse(src)
		require.NoError(t, err)
		require.NotNil(t, mainFrag)
		checkFrag(t, src, mainFrag, expectedKind, C{1, 1}, C{1, 1}, 0)
	})

	t.Run("Bar", func(t *testing.T) {
		src := newSource("bar")
		mainFrag, err := pr.Parse(src)
		require.NoError(t, err)
		require.NotNil(t, mainFrag)
		checkFrag(t, src, mainFrag, expectedKind, C{1, 1}, C{1, 4}, 1)

		// Check elements
		elements := mainFrag.Elements()
		checkFrag(t, src, elements[0], FrBar, C{1, 1}, C{1, 4}, 1)
	})

	t.Run("Foo", func(t *testing.T) {
		src := newSource("foo")
		mainFrag, err := pr.Parse(src)
		require.NoError(t, err)
		require.NotNil(t, mainFrag)
		checkFrag(t, src, mainFrag, expectedKind, C{1, 1}, C{1, 4}, 1)

		// Check elements
		elements := mainFrag.Elements()
		checkFrag(t, src, elements[0], FrFoo, C{1, 1}, C{1, 4}, 1)
	})

	t.Run("FooBar", func(t *testing.T) {
		src := newSource("foobar")
		mainFrag, err := pr.Parse(src)
		require.NoError(t, err)
		require.NotNil(t, mainFrag)
		checkFrag(t, src, mainFrag, expectedKind, C{1, 1}, C{1, 7}, 2)

		// Check elements
		elements := mainFrag.Elements()
		checkFrag(t, src, elements[0], FrFoo, C{1, 1}, C{1, 4}, 1)
		checkFrag(t, src, elements[1], FrBar, C{1, 4}, C{1, 7}, 1)
	})

	t.Run("FooFoo", func(t *testing.T) {
		src := newSource("foofoo")
		mainFrag, err := pr.Parse(src)

		require.Error(t, err)
		require.Equal(t, "unexpected token at test.txt:1:4", err.Error())
		require.Nil(t, mainFrag)
	})
}

func TestParserSuperfluousInput(t *testing.T) {
	expectedKind := llp.FragmentKind(100)
	pr := newParser(t, &llp.Rule{
		Designation: "single foo",
		Pattern:     testR_foo,
		Kind:        expectedKind,
	}, nil)
	mainFrag, err := pr.Parse(newSource("foo "))

	require.Error(t, err)
	require.Equal(t, "unexpected token at test.txt:1:4", err.Error())
	require.Nil(t, mainFrag)
}

func TestParserEither(t *testing.T) {
	t.Run("Neither", func(t *testing.T) {
		expectedKind := llp.FragmentKind(100)
		pr := newParser(t, &llp.Rule{
			Designation: "(Foo / Bar)",
			Pattern: llp.Either{
				testR_foo,
				testR_bar,
			},
			Kind: expectedKind,
		}, nil)
		mainFrag, err := pr.Parse(newSource("far"))

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
		expectedKind := llp.FragmentKind(100)
		pr := newParser(t, &llp.Rule{
			Designation: "(Foo / Bar)",
			Pattern: llp.Either{
				testR_foo,
				testR_bar,
			},
			Kind: expectedKind,
		}, nil)
		src := newSource("foo")
		mainFrag, err := pr.Parse(src)
		require.NoError(t, err)
		require.NotNil(t, mainFrag)
		checkFrag(t, src, mainFrag, expectedKind, C{1, 1}, C{1, 4}, 1)

		// Check elements
		elements := mainFrag.Elements()
		checkFrag(t, src, elements[0], FrFoo, C{1, 1}, C{1, 4}, 1)
	})

	t.Run("Second", func(t *testing.T) {
		expectedKind := llp.FragmentKind(100)
		pr := newParser(t, &llp.Rule{
			Designation: "(Foo / Bar)",
			Pattern: llp.Either{
				testR_foo,
				testR_bar,
			},
			Kind: expectedKind,
		}, nil)
		src := newSource("bar")
		mainFrag, err := pr.Parse(src)
		require.NoError(t, err)
		require.NotNil(t, mainFrag)
		checkFrag(t, src, mainFrag, expectedKind, C{1, 1}, C{1, 4}, 1)

		// Check elements
		elements := mainFrag.Elements()
		checkFrag(t, src, elements[0], FrBar, C{1, 1}, C{1, 4}, 1)
	})
}

func TestParserRecursiveRule(t *testing.T) {
	expectedKind := llp.FragmentKind(100)
	recursiveRule := &llp.Rule{
		Designation: "recursive",
		Kind:        expectedKind,
	}
	recursiveRule.Pattern = llp.Sequence{
		testR_foo,
		termSeparator,
		&llp.Repeated{
			Min:     0,
			Max:     1,
			Pattern: recursiveRule,
		},
	}
	pr := newParser(t, recursiveRule, nil)
	src := newSource("foo,foo,foo,")
	mainFrag, err := pr.Parse(src)

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
	aFrags := make([]llp.Fragment, 0, 2)
	bFrags := make([]llp.Fragment, 0, 2)
	aKind := llp.FragmentKind(905)
	ruleA := &llp.Rule{
		Designation: "a",
		Kind:        aKind,
		Pattern:     &llp.Exact{FrWord, []rune("a")},
		Action: func(f llp.Fragment) error {
			aFrags = append(aFrags, f)
			return nil
		},
	}
	bKind := llp.FragmentKind(906)
	ruleB := &llp.Rule{
		Designation: "b",
		Kind:        bKind,
		Pattern:     &llp.Exact{FrWord, []rune("b")},
		Action: func(f llp.Fragment) error {
			bFrags = append(bFrags, f)
			return nil
		},
	}
	pr := newParser(t, &llp.Rule{
		Designation: "list",
		Pattern: &llp.Repeated{
			Min: 1,
			Pattern: &llp.Rule{
				Designation: "list item",
				Pattern: llp.Sequence{
					llp.Either{ruleA, ruleB},
					termSeparator,
				},
			},
		},
	}, nil)

	src := newSource("a,b,b,a,")
	mainFrag, err := pr.Parse(src)

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
	expectedErr := errors.New("custom error")
	pr := newParser(t, &llp.Rule{
		Designation: "a",
		Kind:        900,
		Pattern:     &llp.Exact{FrWord, []rune("a")},
		Action: func(f llp.Fragment) error {
			return expectedErr
		},
	}, nil)

	mainFrag, err := pr.Parse(newSource("a"))

	require.Error(t, err)
	require.IsType(t, &llp.Err{}, err)
	er := err.(*llp.Err)
	require.Equal(t, expectedErr, er.Err)
	require.Equal(t, uint(0), er.At.Index)
	require.Equal(t, uint(1), er.At.Line)
	require.Equal(t, uint(1), er.At.Column)
	require.Nil(t, mainFrag)
}

func TestParserLexed(t *testing.T) {
	fn := func(_ uint, crs llp.Cursor) bool {
		rn := crs.File.Src[crs.Index]
		if (rn >= 0x0410 && rn <= 0x044F) || rn == '\n' {
			return true
		}
		return false
	}
	expectedKind := llp.FragmentKind(100)

	pr := newParser(t, &llp.Rule{
		Pattern: &llp.Lexed{
			Kind:        expectedKind,
			Designation: "lexed token",
			Fn:          fn,
		},
		Kind: expectedKind,
	}, nil)

	src := newSource("абв\nгде")
	mainFrag, err := pr.Parse(src)

	require.NoError(t, err)
	checkFrag(t, src, mainFrag, expectedKind, C{1, 1}, C{2, 4}, 1)

	// Check elements
	elems := mainFrag.Elements()

	checkFrag(t, src, elems[0], expectedKind, C{1, 1}, C{2, 4}, 0)
}

func TestParserLexedErr(t *testing.T) {
	fn := func(_ uint, crs llp.Cursor) bool {
		rn := crs.File.Src[crs.Index]
		if (rn >= 0x0410 && rn <= 0x044F) || rn == '\n' {
			return true
		}
		return false
	}
	expectedKind := llp.FragmentKind(100)

	pr := newParser(t, &llp.Rule{
		Pattern: &llp.Lexed{
			Kind:        expectedKind,
			Designation: "lexed token",
			Fn:          fn,
		},
		Kind: expectedKind,
	}, nil)

	mainFrag, err := pr.Parse(newSource("abc"))
	require.Error(t, err)
	require.Nil(t, mainFrag)
}

func TestParserLexedErrBelowMinLen(t *testing.T) {
	minLen := uint(3)
	fn := func(idx uint, _ llp.Cursor) bool { return idx < minLen-1 }
	expectedKind := llp.FragmentKind(100)

	pr := newParser(t, &llp.Rule{
		Pattern: &llp.Lexed{
			Kind:        expectedKind,
			Designation: "lexed token",
			MinLen:      minLen,
			Fn:          fn,
		},
		Kind: expectedKind,
	}, nil)

	mainFrag, err := pr.Parse(newSource("abc"))
	require.Error(t, err)
	require.Nil(t, mainFrag)
}

func TestParserErrRule(t *testing.T) {
	t.Run("MatchErr", func(t *testing.T) {
		expectedKind := llp.FragmentKind(100)
		grammar := &llp.Rule{
			Designation: "foo list",
			Pattern: llp.Sequence{
				testR_foo,
				&llp.Exact{Expectation: []rune("...")},
			},
			Kind: expectedKind,
		}
		errGrammar := &llp.Rule{
			Pattern: llp.Either{
				&llp.Repeated{
					Min:     1,
					Pattern: &llp.Exact{Expectation: []rune(";")},
				},
				&llp.Repeated{
					Min:     1,
					Pattern: &llp.Exact{Expectation: []rune(".")},
				},
			},
			Action: func(fr llp.Fragment) error {
				return fmt.Errorf("expected 3 dots, got %d", len(fr.Src()))
			},
		}
		pr := newParser(t, grammar, errGrammar)

		mainFrag, err := pr.Parse(newSource("foo.."))
		require.Error(t, err)
		require.Equal(t, "expected 3 dots, got 2 at test.txt:1:4", err.Error())
		require.Nil(t, mainFrag)
	})

	t.Run("NoMatch", func(t *testing.T) {
		expectedKind := llp.FragmentKind(100)

		grammar := &llp.Rule{
			Designation: "foo list",
			Pattern: llp.Sequence{
				testR_foo,
				&llp.Exact{Expectation: []rune("...")},
			},
			Kind: expectedKind,
		}

		errGrammar := &llp.Rule{
			Pattern: &llp.Repeated{
				Min:     1,
				Pattern: &llp.Exact{Expectation: []rune(";")},
			},
			Action: func(fr llp.Fragment) error {
				return fmt.Errorf(
					"expected 3 semicolons, got %d",
					len(fr.Src()),
				)
			},
		}

		pr := newParser(t, grammar, errGrammar)

		mainFrag, err := pr.Parse(newSource("foo.."))
		require.Error(t, err)
		require.Equal(t, "unexpected token at test.txt:1:4", err.Error())
		require.Nil(t, mainFrag)
	})
}

func TestRepeatedRecursiveRuleUntilEOF(t *testing.T) {
	for _, src := range []string{
		"x",
		"xx",
		"xxx",
	} {
		t.Run(src, func(t *testing.T) {
			ruleA := &llp.Rule{Designation: "A"}
			ruleA.Pattern = llp.Either{
				&llp.Exact{Expectation: []rune("x")},
				ruleA,
			}
			ruleFile := &llp.Rule{
				Designation: "file",
				Pattern: &llp.Repeated{
					Min:     1,
					Pattern: ruleA,
				},
			}
			pr := newParser(t, ruleFile, nil)

			mainFrag, err := pr.Parse(newSource(src))
			require.NoError(t, err)
			require.NotNil(t, mainFrag)
		})
	}
}

func TestParserNot(t *testing.T) {
	t.Run("NoMatch", func(t *testing.T) {
		expectedKind := llp.FragmentKind(100)
		pr := newParser(t, &llp.Rule{
			Designation: "Foo",
			Pattern:     llp.Not{Pattern: testR_foo},
			Kind:        expectedKind,
		}, nil)

		mainFrag, err := pr.Parse(newSource("foo"))
		require.Error(t, err)
		require.Equal(
			t,
			"unexpected token, expected {not a "+
				"keyword foo} at test.txt:1:1",
			err.Error(),
		)
		require.Nil(t, mainFrag)
	})

	t.Run("Match", func(t *testing.T) {
		expectedKind := llp.FragmentKind(100)
		pr := newParser(t, &llp.Rule{
			Designation: "Foo !Foo Bar",
			Pattern: llp.Sequence{
				testR_foo,
				llp.Not{Pattern: testR_foo},
				testR_bar,
			},
			Kind: expectedKind,
		}, nil)

		src := newSource("foobar")
		mainFrag, err := pr.Parse(src)
		require.NoError(t, err)
		require.NotNil(t, mainFrag)
		checkFrag(t, src, mainFrag, expectedKind, C{1, 1}, C{1, 7}, 2)

		// Check elements
		elements := mainFrag.Elements()
		checkFrag(t, src, elements[0], FrFoo, C{1, 1}, C{1, 4}, 1)
		checkFrag(t, src, elements[1], FrBar, C{1, 4}, C{1, 7}, 1)
	})
}

func TestRecursionLimit(t *testing.T) {
	rc := &llp.Rule{
		Designation: "C",
	}
	rb := &llp.Rule{
		Designation: "B",
		Pattern:     rc,
	}
	ra := &llp.Rule{
		Designation: "A",
		Pattern:     rb,
	}
	rc.Pattern = ra

	pr := newParser(t, ra, nil)
	pr.MaxRecursionLevel = 10

	mainFrag, err := pr.Parse(newSource("a"))
	require.Error(t, err)
	require.Equal(
		t,
		fmt.Sprintf(
			"max recursion level exceeded at rule %p (%q) at test.txt:1:1",
			ra,
			ra.Designation,
		),
		err.Error(),
	)
	require.Nil(t, mainFrag)
}

func TestRecursionLimitErrorGrammar(t *testing.T) {
	ec := &llp.Rule{
		Designation: "EC",
	}
	eb := &llp.Rule{
		Designation: "EB",
		Pattern:     ec,
	}
	ea := &llp.Rule{
		Designation: "EA",
		Pattern:     eb,
	}
	ec.Pattern = ea

	pr := newParser(t, &llp.Rule{
		Designation: "error grammar",
		Pattern:     &llp.Exact{Expectation: []rune("okay")},
	}, ea)
	pr.MaxRecursionLevel = 10

	mainFrag, err := pr.Parse(newSource("notokay"))
	require.Error(t, err)
	require.Equal(
		t,
		fmt.Sprintf(
			"max recursion level exceeded at rule %p (%q) at test.txt:1:1",
			ea,
			ea.Designation,
		),
		err.Error(),
	)
	require.Nil(t, mainFrag)
}
