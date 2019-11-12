package parser_test

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	llp "github.com/romshark/llparser"
	parser "github.com/romshark/llparser"
	"github.com/stretchr/testify/require"
)

func stringifyPattern(pt llp.Pattern) string {
	switch tp := pt.(type) {
	case *llp.Rule:
		return fmt.Sprintf("rule (%s)", tp.Designation)
	case *llp.Exact:
		return fmt.Sprintf("exact (%d)", tp.Kind)
	case *llp.Lexed:
		return fmt.Sprintf("lexed (%s)", tp.Designation)
	case llp.Sequence:
		elementNames := make([]string, len(tp))
		for ix, elem := range tp {
			elementNames[ix] = stringifyPattern(elem)
		}
		return fmt.Sprintf(
			"sequence <- %s",
			strings.Join(elementNames, ", "),
		)
	case llp.Not:
		return fmt.Sprintf("not <- %s", stringifyPattern(tp.Pattern))
	case *llp.Repeated:
		return fmt.Sprintf(
			"repeated (min: %d, max: %d) <- %s",
			tp.Min,
			tp.Max,
			stringifyPattern(tp.Pattern),
		)
	case llp.Either:
		optionNames := make([]string, len(tp))
		for ix, elem := range tp {
			optionNames[ix] = stringifyPattern(elem)
		}
		return fmt.Sprintf(
			"either <- %s",
			strings.Join(optionNames, " / "),
		)
	default:
		panic(fmt.Errorf("unexpected pattern type: %s", reflect.TypeOf(tp)))
	}
}

func drawStackTree(out io.Writer, log []*llp.DebugLogEntry) error {
	indent := []byte(". ")
	for ix, ent := range log {
		if _, err := fmt.Fprintf(out, "%d\t| ", ix); err != nil {
			return err
		}
		for i := uint(0); i < ent.Level; i++ {
			if _, err := out.Write(indent); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(
			out,
			"%d:%d %s%s%s %t%s\n",
			ent.At.Line,
			ent.At.Column,
			"\x1b[90m",
			stringifyPattern(ent.Pattern),
			"\x1b[91m",
			ent.Matched,
			"\x1b[0m",
		); err != nil {
			return err
		}
	}
	return nil
}

type E struct {
	At      string
	Pattern string
	Level   int
	Match   bool
}

func checkExpectations(
	t *testing.T,
	debugProfile *parser.DebugProfile,
	expectedEntries ...E,
) {
	for ix, expected := range expectedEntries {
		require.True(
			t,
			ix < len(debugProfile.Log),
			"missing entry at index %d",
			ix,
		)
		actual := debugProfile.Log[ix]
		require.Equal(
			t,
			expected.At,
			actual.At.String(),
			"unexpected position for %d", ix,
		)
		require.Equal(
			t,
			expected.Pattern,
			stringifyPattern(actual.Pattern),
			"unexpected pattern for %d", ix,
		)
		require.Equal(
			t,
			uint(expected.Level),
			actual.Level,
			"unexpected level for %d", ix,
		)
		require.Equal(
			t,
			expected.Match,
			actual.Matched,
			"unexpected matched-flag for %d", ix,
		)
	}
}

func TestDebug(t *testing.T) {
	const (
		kindMain = 100 + iota
		kindA
		kindB
		kindC
		kindD
		kindE
	)

	lexedE := &llp.Lexed{
		Kind:        kindE,
		Designation: "E",
		MinLen:      1,
		Fn: func(ix uint, cr llp.Cursor) bool {
			// Number (0-9)
			rn := cr.File.Src[cr.Index]
			return rn >= 0x30 && rn <= 0x39
		},
	}

	exactC := &llp.Exact{Kind: kindC, Expectation: []rune("c")}
	exactD := &llp.Exact{Kind: kindD, Expectation: []rune("d")}

	ruleA := &llp.Rule{
		Kind:        kindA,
		Designation: "A",
		Pattern:     llp.Sequence{exactC, exactD},
	}

	ruleB := &llp.Rule{
		Kind:        kindB,
		Designation: "B",
		Pattern:     lexedE,
	}

	grammar := &llp.Rule{
		Kind:        kindMain,
		Designation: "main",
		Pattern: &llp.Repeated{
			Min:     1,
			Max:     10,
			Pattern: llp.Either{ruleA, ruleB},
		},
	}

	parser, err := llp.NewParser(grammar, nil)
	require.NoError(t, err)

	profile, parseTree, err := parser.Debug(&llp.SourceFile{
		Name: "test.txt",
		Src:  []rune("cdcd1234cd999"),
	})

	require.NotNil(t, profile)
	require.NoError(t, err)
	require.NotNil(t, parseTree)

	require.NoError(t, drawStackTree(os.Stdout, profile.Log))

	const (
		dRuleMain = "rule (main)"
		dRuleA    = "rule (A)"
		dRuleB    = "rule (B)"
		dChs1     = "either <- " + dRuleA + " / " + dRuleB
		dRep1     = "repeated (min: 1, max: 10) <- " + dChs1
		dExt103   = "exact (103)"
		dExt104   = "exact (104)"
		dSeq1     = "sequence <- " + dExt103 + ", " + dExt104
		dLexE     = "lexed (E)"
	)

	checkExpectations(t, profile,
		E{"test.txt:1:1", dRuleMain, 0, true}, // 0
		E{"test.txt:1:1", dRep1, 1, true},     // 1
		E{"test.txt:1:1", dChs1, 2, true},     // 2
		E{"test.txt:1:1", dRuleA, 3, true},    // 3
		E{"test.txt:1:1", dSeq1, 4, true},     // 4
		E{"test.txt:1:1", dExt103, 5, true},   // 5
		E{"test.txt:1:2", dExt104, 5, true},   // 6
		E{"test.txt:1:3", dChs1, 2, true},     // 7
		E{"test.txt:1:3", dRuleA, 3, true},    // 8
		E{"test.txt:1:3", dSeq1, 4, true},     // 9
		E{"test.txt:1:3", dExt103, 5, true},   // 10
		E{"test.txt:1:4", dExt104, 5, true},   // 11
		E{"test.txt:1:5", dChs1, 2, true},     // 12
		E{"test.txt:1:5", dRuleA, 3, false},   // 13
		E{"test.txt:1:5", dSeq1, 4, false},    // 14
		E{"test.txt:1:5", dExt103, 5, false},  // 15
		E{"test.txt:1:5", dRuleB, 3, true},    // 16
		E{"test.txt:1:5", dLexE, 4, true},     // 17
		E{"test.txt:1:9", dChs1, 2, true},     // 18
		E{"test.txt:1:9", dRuleA, 3, true},    // 19
		E{"test.txt:1:9", dSeq1, 4, true},     // 20
		E{"test.txt:1:9", dExt103, 5, true},   // 21
		E{"test.txt:1:10", dExt104, 5, true},  // 22
		E{"test.txt:1:11", dChs1, 2, true},    // 23
		E{"test.txt:1:11", dRuleA, 3, false},  // 24
		E{"test.txt:1:11", dSeq1, 4, false},   // 25
		E{"test.txt:1:11", dExt103, 5, false}, // 26
		E{"test.txt:1:11", dRuleB, 3, true},   // 27
		E{"test.txt:1:11", dLexE, 4, true},    // 28
		E{"test.txt:1:14", dChs1, 2, false},   // 29
		E{"test.txt:1:14", dRuleA, 3, false},  // 30
		E{"test.txt:1:14", dSeq1, 4, false},   // 31
		E{"test.txt:1:14", dExt103, 5, false}, // 32
	)
}

func TestDebugRecursion(t *testing.T) {
	grammar := &llp.Rule{Designation: "main"}
	grammar.Pattern = grammar

	parser, err := llp.NewParser(grammar, nil)
	require.NoError(t, err)
	parser.MaxRecursionLevel = 3

	profile, parseTree, err := parser.Debug(&llp.SourceFile{
		Name: "test.txt",
		Src:  []rune("cdcd1234cd999"),
	})

	require.Error(t, err)
	require.Nil(t, parseTree)
	require.NotNil(t, profile)

	require.NoError(t, drawStackTree(os.Stdout, profile.Log))
	require.Len(t, profile.Log, 4)

	const dRuleMain = "rule (main)"

	checkExpectations(t, profile,
		E{"test.txt:1:1", dRuleMain, 0, false}, // 0
		E{"test.txt:1:1", dRuleMain, 1, false}, // 1
		E{"test.txt:1:1", dRuleMain, 2, false}, // 2
		E{"test.txt:1:1", dRuleMain, 3, true},  // 3
	)
}

func TestDebugMismatch(t *testing.T) {
	const (
		kindA = 100 + iota
		kindB
	)

	exactA := &llp.Exact{Kind: kindA, Expectation: []rune("aaa")}
	exactB := &llp.Exact{Kind: kindB, Expectation: []rune("bbb")}
	parser, err := llp.NewParser(&llp.Rule{
		Designation: "main",
		Pattern:     llp.Either{exactA, exactB},
	}, nil)
	require.NoError(t, err)
	parser.MaxRecursionLevel = 3

	profile, parseTree, err := parser.Debug(&llp.SourceFile{
		Name: "test.txt",
		Src:  []rune("aabbb"),
	})

	require.Error(t, err)
	require.Nil(t, parseTree)
	require.NotNil(t, profile)

	require.NoError(t, drawStackTree(os.Stdout, profile.Log))

	const (
		dRlMain = "rule (main)"
		dExA    = "exact (100)"
		dExB    = "exact (101)"
		dCh1    = "either <- " + dExA + " / " + dExB
	)

	checkExpectations(t, profile,
		E{"test.txt:1:1", dRlMain, 0, false}, // 0
		E{"test.txt:1:1", dCh1, 1, false},    // 1
		E{"test.txt:1:1", dExA, 2, false},    // 2
		E{"test.txt:1:1", dExB, 2, false},    // 3
	)
}

func TestDebugMismatchSequence(t *testing.T) {
	const (
		kindA = 100 + iota
		kindB
	)

	exactA := &llp.Exact{Kind: kindA, Expectation: []rune("aaa")}
	exactB := &llp.Exact{Kind: kindB, Expectation: []rune("bbb")}
	parser, err := llp.NewParser(&llp.Rule{
		Designation: "main",
		Pattern:     llp.Sequence{exactA, exactB},
	}, nil)
	require.NoError(t, err)

	profile, parseTree, err := parser.Debug(&llp.SourceFile{
		Name: "test.txt",
		Src:  []rune("aaabb"),
	})

	require.Error(t, err)
	require.Nil(t, parseTree)
	require.NotNil(t, profile)

	require.NoError(t, drawStackTree(os.Stdout, profile.Log))

	const (
		dRlMain = "rule (main)"
		dExA    = "exact (100)"
		dExB    = "exact (101)"
		dSq1    = "sequence <- " + dExA + ", " + dExB
	)

	checkExpectations(t, profile,
		E{"test.txt:1:1", dRlMain, 0, false}, // 0
		E{"test.txt:1:1", dSq1, 1, false},    // 1
		E{"test.txt:1:1", dExA, 2, true},     // 2
		E{"test.txt:1:4", dExB, 2, false},    // 3
	)
}

func TestDebugErrorGrammar(t *testing.T) {
	const (
		kindA = 100 + iota
		kindE
		kindEA
		kindEB
		kindEC
	)

	exactA := &llp.Exact{Kind: kindA, Expectation: []rune("cba")}
	grammar := &llp.Rule{Designation: "main", Pattern: exactA}
	errorGrammar := &llp.Rule{
		Designation: "error",
		Kind:        kindE,
		Pattern: llp.Sequence{
			&llp.Exact{Kind: kindEA, Expectation: []rune("a")},
			&llp.Exact{Kind: kindEB, Expectation: []rune("b")},
			&llp.Exact{Kind: kindEC, Expectation: []rune("c")},
		},
	}

	parser, err := llp.NewParser(grammar, errorGrammar)
	require.NoError(t, err)

	profile, parseTree, err := parser.Debug(&llp.SourceFile{
		Name: "test.txt",
		Src:  []rune("abc"),
	})

	require.Error(t, err)
	require.Nil(t, parseTree)
	require.NotNil(t, profile)

	require.NoError(t, drawStackTree(os.Stdout, profile.Log))

	const (
		dRlMain  = "rule (main)"
		dExA     = "exact (100)"
		dExEA    = "exact (102)"
		dExEB    = "exact (103)"
		dExEC    = "exact (104)"
		dSq1     = "sequence <- " + dExEA + ", " + dExEB + ", " + dExEC
		dRlError = "rule (error)"
	)

	checkExpectations(t, profile,
		E{"test.txt:1:1", dRlMain, 0, false}, // 0
		E{"test.txt:1:1", dExA, 1, false},    // 1
		E{"test.txt:1:1", dRlError, 0, true}, // 2
		E{"test.txt:1:1", dSq1, 1, true},     // 3
		E{"test.txt:1:1", dExEA, 2, true},    // 4
		E{"test.txt:1:2", dExEB, 2, true},    // 5
		E{"test.txt:1:3", dExEC, 2, true},    // 6
	)
}
