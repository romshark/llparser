package parser

import (
	"errors"
	"fmt"
	"reflect"
)

// Lexer defines the interface of an abstract lexer implementation
type Lexer interface {
	Next() (*Token, error)
	Position() Cursor
	Fork() Lexer
	Set(Cursor)
}

// Parser represents a parser
type Parser struct{}

// NewParser creates a new parser instance
func NewParser() Parser {
	return Parser{}
}

func (pr Parser) handlePattern(
	scanner *Scanner,
	pattern Pattern,
) (Fragment, error) {
	switch pt := pattern.(type) {
	case *Rule:
		return pr.parseRule(scanner.New(), pt)
	case Term:
		// Terminal
		return pr.parseTerm(scanner, pt)
	case TermExact:
		// Exact terminal
		return pr.parseTermExact(scanner, pt)
	case Checked:
		// Checked terminal
		panic("not yet implemented")
	case ZeroOrMore:
		// ZeroOrMore
		panic("not yet implemented")
	case OneOrMore:
		// OneOrMore
		panic("not yet implemented")
	case Optional:
		// Optional
		panic("not yet implemented")
	case Sequence:
		// Sequence
		return pr.parseSequence(scanner, pt)
	case Either:
		// Choice
		panic("not yet implemented")
	default:
		panic(fmt.Errorf(
			"unsupported pattern type: %s",
			reflect.TypeOf(pattern),
		))
	}
}

func (pr Parser) parseSequence(
	scanner *Scanner,
	patterns []Pattern,
) (Fragment, error) {
	for _, pt := range patterns {
		frag, err := pr.handlePattern(scanner, pt)
		if err != nil {
			return frag, err
		}
		if frag == nil {
			return nil, nil
		}
		// Append rule patterns, other patterns are appended automatically
		if _, isRule := pt.(*Rule); isRule {
			scanner.Append(frag)
		}
	}
	return nil, nil
}

func (pr Parser) parseTermExact(
	scanner *Scanner,
	term TermExact,
) (Fragment, error) {
	tk, err := scanner.Next()
	if err != nil {
		return nil, err
	}
	if string(term) != tk.Src() {
		return nil, nil
	}
	return tk, nil
}

func (pr Parser) parseTerm(
	scanner *Scanner,
	term Term,
) (Fragment, error) {
	tk, err := scanner.Next()
	if err != nil {
		return nil, err
	}
	if tk == nil {
		return nil, nil
	}
	if tk.VKind != FragmentKind(term) {
		return nil, nil
	}
	return tk, nil
}

func (pr Parser) parseRule(
	scanner *Scanner,
	rule *Rule,
) (Fragment, error) {
	frag, err := pr.handlePattern(scanner, rule.Pattern)
	if err != nil {
		return frag, err
	}
	return scanner.Fragment(rule.Kind), nil
}

// Parse parses the given rule
func (pr Parser) Parse(lexer Lexer, rule *Rule) (Fragment, error) {
	if lexer == nil {
		return nil, errors.New("missing lexer while parsing")
	}
	return pr.parseRule(NewScanner(lexer), rule)
}
