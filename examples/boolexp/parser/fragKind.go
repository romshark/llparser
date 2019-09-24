package parser

import llp "github.com/romshark/llparser"

// FragKind represents a dick-lang fragment kind
type FragKind = llp.FragmentKind

const (
	_ FragKind = iota

	// FrSpace represents an arbitrary sequence of spaces, tabs and line-breaks
	FrSpace

	// FrExpr represents a boolean expression
	FrExpr

	// FrExprParentheses represents an expression enclosed by parentheses
	FrExprParentheses

	// FrExprFactor represents a factor expression
	FrExprFactor

	// FrExprTerm represents a term expression
	FrExprTerm

	// FrConstTrue represents the constant boolean value "true"
	FrConstTrue

	// FrConstFalse represents the constant boolean value "false"
	FrConstFalse

	// FrVarRef represents a variable reference
	FrVarRef

	// FrOprOr represents the logical or-operator
	FrOprOr

	// FrOprAnd represents the logical and-operator
	FrOprAnd

	// FrOprNeg represents the boolean negation operator
	FrOprNeg

	// FrParOpen represents the opening parenthesis
	FrParOpen

	// FrParClose represents the closing parenthesis
	FrParClose
)

// FragKindString translates the kind identifier to its name
func FragKindString(kind llp.FragmentKind) string {
	switch kind {
	case FrSpace:
		return "Space"
	case FrExpr:
		return "Expr"
	case FrExprParentheses:
		return "ExprParentheses"
	case FrExprFactor:
		return "ExprFactor"
	case FrExprTerm:
		return "ExprTerm"
	case FrConstTrue:
		return "ConstTrue"
	case FrConstFalse:
		return "ConstFalse"
	case FrVarRef:
		return "VarRef"
	case FrOprOr:
		return "OprOr"
	case FrOprAnd:
		return "OprAnd"
	case FrOprNeg:
		return "OprNeg"
	case FrParOpen:
		return "ParOpen"
	case FrParClose:
		return "ParClose"
	}
	return ""
}
