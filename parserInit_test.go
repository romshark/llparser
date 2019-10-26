package parser_test

import (
	"fmt"
	"testing"

	llp "github.com/romshark/llparser"
	"github.com/stretchr/testify/require"
)

func str(format string, a ...interface{}) string {
	return fmt.Sprintf(format, a...)
}

func test(t *testing.T, pattern llp.Pattern, expectedErrMsg string) {
	// Wrap non-rules into dummy rules
	var rule *llp.Rule
	if ptr, ok := pattern.(*llp.Rule); ok {
		rule = ptr
	} else {
		rule = &llp.Rule{
			Pattern: pattern,
		}
	}

	pr, err := llp.NewParser(rule, rule)
	require.Error(t, err)
	require.Nil(t, pr)
	require.Equal(t, expectedErrMsg, err.Error())
}

type UnsupportedPatternType struct{}

func (UnsupportedPatternType) Container() bool              { return true }
func (UnsupportedPatternType) TerminalPattern() llp.Pattern { return nil }
func (UnsupportedPatternType) Desig() string {
	return "UnsupportedPatternType"
}

func TestUnsupportedPatternType(t *testing.T) {
	test(
		t,
		UnsupportedPatternType{},
		"invalid grammar: unsupported pattern type: "+
			"parser_test.UnsupportedPatternType",
	)
}

func TestRuleMissingPattern(t *testing.T) {
	rl := &llp.Rule{}
	test(t, rl, str("invalid grammar: rule %p is missing a pattern", rl))
}

func TestRepeatedMinGreaterMax(t *testing.T) {
	min := uint(2)
	max := uint(1)
	rp := &llp.Repeated{
		Min:     min,
		Max:     max,
		Pattern: &llp.Exact{Expectation: []rune("test")},
	}
	test(t, rp, str(
		"invalid grammar: repeated %p min (%d) greater max (%d)",
		rp,
		min,
		max,
	))
}

func TestRepeatedMissingPattern(t *testing.T) {
	rp := &llp.Repeated{}
	test(t, rp, str("invalid grammar: repeated %p is missing a pattern", rp))
}

func TestSequenceEmpty(t *testing.T) {
	test(t, llp.Sequence{}, "invalid grammar: sequence is empty")
}

func TestEitherEmpty(t *testing.T) {
	test(
		t,
		llp.Either{},
		"invalid grammar: either-combinator has 0 option(s)",
	)
}

func TestEitherOneOption(t *testing.T) {
	test(t, llp.Either{
		&llp.Exact{Expectation: []rune("test")},
	}, "invalid grammar: either-combinator has 1 option(s)")
}

func TestEitherDuplicateOptions(t *testing.T) {
	opt1 := &llp.Exact{Expectation: []rune("opt1")}
	opt2 := &llp.Exact{Expectation: []rune("opt2")}
	test(t, llp.Either{
		opt1,
		opt2,
		opt1,
	}, "invalid grammar: either-combinator has "+
		"duplicate options (at index 2)",
	)
}

func TestLexedMissingFn(t *testing.T) {
	lx := &llp.Lexed{}
	test(t, lx, str(
		"invalid grammar: lexed-terminal %p is missing the lexer function",
		lx,
	))
}

func TestExactMissingExpectation(t *testing.T) {
	ex := &llp.Exact{}
	test(t, ex, str(
		"invalid grammar: exact-terminal %p is missing an expectation",
		ex,
	))
}

func TestNotNested(t *testing.T) {
	ex := &llp.Exact{Expectation: []rune("test")}
	test(
		t,
		llp.Not{Pattern: llp.Not{Pattern: ex}},
		"invalid grammar: not-combinator is nested",
	)
}
