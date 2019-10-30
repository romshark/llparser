package parser

import llp "github.com/romshark/llparser"

func optional(pattern llp.Pattern) *llp.Repeated {
	return &llp.Repeated{
		Min:     0,
		Max:     1,
		Pattern: pattern,
	}
}

// newGrammar returns the grammar of a boolean expression
func newGrammar() *llp.Rule {
	termSpace := &llp.Lexed{
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

	termAnything := &llp.Lexed{
		Fn: func(uint, llp.Cursor) bool { return false },
	}

	termConstTrue := &llp.Exact{
		Kind:        FrConstTrue,
		Expectation: []rune("true"),
	}

	termConstFalse := &llp.Exact{
		Kind:        FrConstFalse,
		Expectation: []rune("false"),
	}

	termParOpen := &llp.Exact{
		Kind:        FrParOpen,
		Expectation: []rune("("),
	}

	termParClose := &llp.Exact{
		Kind:        FrParClose,
		Expectation: []rune(")"),
	}

	termOprOr := &llp.Exact{
		Kind:        FrOprOr,
		Expectation: []rune("||"),
	}

	termOprAnd := &llp.Exact{
		Kind:        FrOprAnd,
		Expectation: []rune("&&"),
	}

	termOprNeg := &llp.Exact{
		Kind:        FrOprNeg,
		Expectation: []rune("!"),
	}

	expression := &llp.Rule{
		Kind:        FrExpr,
		Designation: "boolean expression",
	}

	parentheses := &llp.Rule{
		Kind:        FrExprParentheses,
		Designation: "parentheses",
		Pattern: llp.Sequence{
			termParOpen,
			optional(termSpace),
			expression,
			optional(termSpace),
			termParClose,
		},
	}

	factor := &llp.Rule{
		Kind:        FrExprFactor,
		Designation: "factor",
	}
	factor.Pattern = llp.Either{
		llp.Either{
			termConstTrue,
			termConstFalse,
		},
		llp.Sequence{
			termOprNeg,
			factor,
		},
		parentheses,
		llp.Not{Pattern: termAnything},
		expression,
	}

	term := &llp.Rule{
		Designation: "term",
		Kind:        FrExprTerm,
		Pattern: llp.Sequence{
			factor,
			&llp.Repeated{
				Pattern: llp.Sequence{
					optional(termSpace),
					termOprAnd,
					optional(termSpace),
					factor,
				},
			},
		},
	}

	expression.Pattern = llp.Sequence{
		term,
		&llp.Repeated{
			Pattern: llp.Sequence{
				optional(termSpace),
				termOprOr,
				optional(termSpace),
				term,
			},
		},
	}

	return expression
}
