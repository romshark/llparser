package parser

import (
	"fmt"

	llp "github.com/romshark/llparser"
)

// Parser represents a boolean expression parser
type Parser struct {
	prs *llp.Parser
}

// NewParser creates a new parser instance
func NewParser() (*Parser, error) {
	parser, err := llp.NewParser(newGrammar(), nil)
	if err != nil {
		return nil, err
	}
	return &Parser{prs: parser}, nil
}

// Parse parses a boolean expression
func (pr *Parser) Parse(
	fileName string,
	src []rune,
) (*AST, error) {
	if fileName == "" {
		return nil, fmt.Errorf("invalid file name")
	}

	if len(src) < 1 {
		return nil, fmt.Errorf("empty input")
	}

	mainFrag, err := pr.prs.Parse(&llp.SourceFile{
		Name: fileName,
		Src:  src,
	})
	if err != nil {
		return nil, err
	}

	ast := &AST{
		Root: &ASTRoot{
			Fragment:   mainFrag,
			Expression: parseExpr(mainFrag),
		},
	}

	// Turn the parse tree into an AST
	if err != nil {
		return nil, err
	}

	return ast, nil
}

func findKind(
	frags []llp.Fragment,
	offset int,
	kinds ...llp.FragmentKind,
) (int, llp.Fragment) {
	for fix := offset; fix < len(frags); fix++ {
		frag := frags[fix]
		for _, kind := range kinds {
			if frag.Kind() == kind {
				return fix, frag
			}
		}
	}
	return -1, nil
}

func parseExpr(frag llp.Fragment) ASTExpression {
	elems := frag.Elements()
	terms := []ASTExpression{}

	// Parse all terms
	for ix := 0; ix < len(elems); ix++ {
		elem := elems[ix]
		if elem.Kind() == FrExprTerm {
			terms = append(terms, parseTerm(elem))
		}
	}

	// Check for or-statements
	if _, fr := findKind(elems, 1, FrOprOr); fr == nil {
		if len(terms) < 1 {
			return nil
		}
		return terms[0]
	}
	return &ASTOr{
		Fragment:    frag,
		Expressions: terms,
	}
}

func parseTerm(frag llp.Fragment) ASTExpression {
	elems := frag.Elements()
	factors := []ASTExpression{}

	// Parse all factors
	for ix := 0; ix < len(elems); ix++ {
		elem := elems[ix]
		if elem.Kind() == FrExprFactor {
			factors = append(factors, parseFactor(elem))
		}
	}

	// Check for and-statements
	if _, fr := findKind(elems, 1, FrOprAnd); fr == nil {
		if len(factors) < 1 {
			return nil
		}
		return factors[0]
	}
	return &ASTAnd{
		Fragment:    frag,
		Expressions: factors,
	}
}

func parseFactor(frag llp.Fragment) ASTExpression {
	elems := frag.Elements()
	firstElem := elems[0]

	switch firstElem.Kind() {
	case FrConstTrue:
		// Constant bool value
		return &ASTConstant{
			Fragment: firstElem,
			Value:    true,
		}
	case FrConstFalse:
		// Constant bool value
		return &ASTConstant{
			Fragment: firstElem,
			Value:    false,
		}
	case FrOprNeg:
		// Negated factor
		return &ASTNegated{
			Fragment:   frag,
			Expression: parseFactor(elems[1]),
		}
	case FrExprParentheses:
		// Expression enclosed in parentheses
		_, expr := findKind(elems[0].Elements(), 1, FrExpr)
		return &ASTParentheses{
			Fragment:   frag,
			Expression: parseExpr(expr),
		}
	}

	return nil
}
