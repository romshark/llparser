package parser

import (
	"fmt"
	"io"
	"os"
	"reflect"
)

var snipConstTrue = []byte("const(true)")
var snipConstFalse = []byte("const(false)")
var snipOr = []byte("or{")
var snipAnd = []byte("and{")
var snipNeg = []byte("neg{")
var snipPar = []byte("par{")
var snipBlkEnd = []byte("}")
var snipLineBreak = []byte("\n")
var snipSpace = []byte(" ")

// ASTPrintOptions defines the AST printing options
type ASTPrintOptions struct {
	Out         io.Writer
	Indentation []byte
	Prefix      []byte
}

// Print prints an expression recursively
func (ast *AST) Print(opts ASTPrintOptions) (bytesWritten int, err error) {
	if opts.Out == nil {
		opts.Out = os.Stdout
	}

	// write returns true if there was an error,
	// otherwise returns false
	write := func(str []byte) bool {
		var bw int
		bw, err = opts.Out.Write(str)
		if err != nil {
			// Abort printing
			return true
		}
		bytesWritten += bw
		// Continue printing
		return false
	}

	writePrefix := func() bool {
		// Write the prefix if any
		if len(opts.Prefix) > 0 {
			return write(opts.Prefix)
		}
		return false
	}

	writeLnBrk := func() bool {
		if len(opts.Indentation) < 1 {
			// Write whitespace instead when indentation is disabled
			return write(snipSpace)
		}

		// Write line-break
		if write(snipLineBreak) {
			return true
		}

		return writePrefix()
	}

	printIndent := func(ind uint) bool {
		// Write the indentation
		if len(opts.Indentation) > 0 {
			for ix := uint(0); ix < ind; ix++ {
				if write(opts.Indentation) {
					return true
				}
			}
		}
		return false
	}

	var prt func(ind uint, expr ASTExpression) bool
	prt = func(ind uint, expr ASTExpression) bool {
		if printIndent(ind) {
			return true
		}

		// Write the actual line
		switch expr := expr.(type) {
		case *ASTOr:
			// Print or-expression
			if write(snipOr) {
				return true
			}
			if writeLnBrk() {
				return true
			}
			for _, subExpr := range expr.Expressions {
				if prt(ind+1, subExpr) {
					return true
				}
				if writeLnBrk() {
					return true
				}
			}
			if printIndent(ind) {
				return true
			}
			return write(snipBlkEnd)
		case *ASTAnd:
			// Print and-expression
			if write(snipAnd) {
				return true
			}
			if writeLnBrk() {
				return true
			}
			for _, subExpr := range expr.Expressions {
				if prt(ind+1, subExpr) {
					return true
				}
				if writeLnBrk() {
					return true
				}
			}
			if printIndent(ind) {
				return true
			}
			return write(snipBlkEnd)
		case *ASTConstant:
			// Print constant
			switch expr.Value {
			case true:
				if write(snipConstTrue) {
					return true
				}
			case false:
				if write(snipConstFalse) {
					return true
				}
			}
		case *ASTNegated:
			// Print negated expression
			if write(snipNeg) {
				return true
			}
			if writeLnBrk() {
				return true
			}
			if prt(ind+1, expr.Expression) {
				return true
			}
			if writeLnBrk() {
				return true
			}
			if printIndent(ind) {
				return true
			}
			return write(snipBlkEnd)
		case *ASTParentheses:
			// Print expression enclosed in parantheses
			if write(snipPar) {
				return true
			}
			if writeLnBrk() {
				return true
			}
			if prt(ind+1, expr.Expression) {
				return true
			}
			if writeLnBrk() {
				return true
			}
			if printIndent(ind) {
				return true
			}
			return write(snipBlkEnd)
		default:
			panic(fmt.Errorf("unexpected node type: %s", reflect.TypeOf(expr)))
		}
		return false
	}

	if writePrefix() {
		return
	}
	prt(0, ast.Root.Expression)
	return
}
