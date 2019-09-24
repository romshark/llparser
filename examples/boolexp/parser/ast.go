package parser

import (
	llp "github.com/romshark/llparser"
)

// ASTExpression represents an abstract expression
type ASTExpression interface {
	Frag() llp.Fragment
	IsConst() bool
	Val() bool
}

// ASTRoot represents the AST root
type ASTRoot struct {
	Fragment   llp.Fragment
	Expression ASTExpression
}

// Frag implements the ASTExpression interface
func (ex *ASTRoot) Frag() llp.Fragment { return ex.Fragment }

// IsConst implements the ASTExpression interface
func (ex *ASTRoot) IsConst() bool { return false }

// Val implements the ASTExpression interface
func (ex *ASTRoot) Val() bool { return ex.Expression.Val() }

// ASTConstant represents a constant expression
type ASTConstant struct {
	Fragment llp.Fragment
	Value    bool
}

// Frag implements the ASTExpression interface
func (ex *ASTConstant) Frag() llp.Fragment { return ex.Fragment }

// IsConst implements the ASTExpression interface
func (ex *ASTConstant) IsConst() bool { return false }

// Val implements the ASTExpression interface
func (ex *ASTConstant) Val() bool { return ex.Value }

// ASTNegated represents a constant expression
type ASTNegated struct {
	Fragment   llp.Fragment
	Expression ASTExpression
}

// Frag implements the ASTExpression interface
func (ex *ASTNegated) Frag() llp.Fragment { return ex.Fragment }

// IsConst implements the ASTExpression interface
func (ex *ASTNegated) IsConst() bool { return false }

// Val implements the ASTExpression interface
func (ex *ASTNegated) Val() bool { return !ex.Expression.Val() }

// ASTOr represents an or-expression
type ASTOr struct {
	Fragment    llp.Fragment
	Expressions []ASTExpression
}

// Frag implements the ASTExpression interface
func (ex *ASTOr) Frag() llp.Fragment { return ex.Fragment }

// IsConst implements the ASTExpression interface
func (ex *ASTOr) IsConst() bool { return false }

// Val implements the ASTExpression interface
func (ex *ASTOr) Val() bool {
	for _, expr := range ex.Expressions {
		if expr.Val() {
			return true
		}
	}
	return false
}

// ASTAnd represents an and-expression
type ASTAnd struct {
	Fragment    llp.Fragment
	Expressions []ASTExpression
}

// Frag implements the ASTExpression interface
func (ex *ASTAnd) Frag() llp.Fragment { return ex.Fragment }

// IsConst implements the ASTExpression interface
func (ex *ASTAnd) IsConst() bool { return false }

// Val implements the ASTExpression interface
func (ex *ASTAnd) Val() bool {
	for _, expr := range ex.Expressions {
		if expr.Val() == false {
			return false
		}
	}
	return true
}

// ASTParentheses represents an expression enclosed in parentheses
type ASTParentheses struct {
	Fragment   llp.Fragment
	Expression ASTExpression
}

// Frag implements the ASTExpression interface
func (ex *ASTParentheses) Frag() llp.Fragment { return ex.Fragment }

// IsConst implements the ASTExpression interface
func (ex *ASTParentheses) IsConst() bool { return false }

// Val implements the ASTExpression interface
func (ex *ASTParentheses) Val() bool { return ex.Expression.Val() }

// AST represents the abstract syntax tree of a boolean expression
type AST struct {
	Root *ASTRoot
}
