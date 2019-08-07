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
		tk, err := pr.parseRule(scanner.New(), pt)
		if err, ok := err.(*ErrUnexpectedToken); ok {
			// Override expected pattern to the higher-order rule
			err.Expected = pt
			return nil, err
		} else if err != nil {
			return nil, err
		}
		return tk, nil
	case Term:
		// Terminal
		return pr.parseTerm(scanner, pt)
	case TermExact:
		// Exact terminal
		return pr.parseTermExact(scanner, pt)
	case Checked:
		return pr.parseChecked(scanner, pt)
	case ZeroOrMore:
		// ZeroOrMore
		return pr.parseZeroOrMore(scanner, pt.Pattern)
	case OneOrMore:
		// OneOrMore
		return pr.parseOneOrMore(scanner, pt.Pattern)
	case Optional:
		// Optional
		return pr.parseOptional(scanner, pt.Pattern)
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

func (pr Parser) parseChecked(
	scanner *Scanner,
	expected Checked,
) (Fragment, error) {
	beforeCr := scanner.Lexer.Position()
	tk, err := scanner.Next()
	if err != nil {
		return nil, err
	}
	if tk == nil || !expected.Fn(tk.Src()) {
		return nil, &ErrUnexpectedToken{
			At:       beforeCr,
			Expected: expected,
			Actual:   tk,
		}
	}
	return tk, nil
}

func (pr Parser) parseOptional(
	scanner *Scanner,
	pattern Pattern,
) (Fragment, error) {
	beforeCr := scanner.Lexer.Position()
	tk, err := pr.handlePattern(scanner, pattern)
	if err != nil {
		if _, ok := err.(*ErrUnexpectedToken); ok {
			// Reset scanner to the initial position
			scanner.Lexer.Set(beforeCr)
			return nil, nil
		}
		return nil, err
	}
	return tk, nil
}

func (pr Parser) parseZeroOrMore(
	scanner *Scanner,
	pattern Pattern,
) (Fragment, error) {
	lastPosition := scanner.Lexer.Position()
	for {
		frag, err := pr.handlePattern(scanner, pattern)
		if err != nil {
			if _, ok := err.(*ErrUnexpectedToken); ok {
				// Reset scanner to the last match
				scanner.Set(lastPosition)
				return nil, nil
			}
			return nil, err
		}
		lastPosition = scanner.Lexer.Position()
		// Append rule patterns, other patterns are appended automatically
		if _, isRule := pattern.(*Rule); isRule {
			scanner.Append(frag)
		}
	}
}

func (pr Parser) parseOneOrMore(
	scanner *Scanner,
	pattern Pattern,
) (Fragment, error) {
	num := 0
	lastPosition := scanner.Lexer.Position()
	for {
		frag, err := pr.handlePattern(scanner, pattern)
		if err != nil {
			if num < 1 {
				return nil, err
			}
			if _, ok := err.(*ErrUnexpectedToken); ok {
				// Reset scanner to the last match
				scanner.Set(lastPosition)
				return nil, nil
			}
			return nil, err
		}
		num++
		lastPosition = scanner.Lexer.Position()
		// Append rule patterns, other patterns are appended automatically
		if _, isRule := pattern.(*Rule); isRule {
			scanner.Append(frag)
		}
	}
}

func (pr Parser) parseSequence(
	scanner *Scanner,
	patterns []Pattern,
) (Fragment, error) {
	for _, pt := range patterns {
		frag, err := pr.handlePattern(scanner, pt)
		if err != nil {
			return nil, err
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
	expected TermExact,
) (Fragment, error) {
	beforeCr := scanner.Lexer.Position()
	tk, err := scanner.Next()
	if err != nil {
		return nil, err
	}
	if tk == nil || tk.Src() != string(expected) {
		return nil, &ErrUnexpectedToken{
			At:       beforeCr,
			Expected: expected,
			Actual:   tk,
		}
	}
	return tk, nil
}

func (pr Parser) parseTerm(
	scanner *Scanner,
	expected Term,
) (Fragment, error) {
	beforeCr := scanner.Lexer.Position()
	tk, err := scanner.Next()
	if err != nil {
		return nil, err
	}
	if tk == nil || tk.VKind != FragmentKind(expected) {
		return nil, &ErrUnexpectedToken{
			At:       beforeCr,
			Expected: expected,
			Actual:   tk,
		}
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
	mainFrag, err := pr.parseRule(NewScanner(lexer), rule)
	if err != nil {
		return nil, err
	}

	// Ensure EOF
	before := lexer.Position()
	last, err := lexer.Next()
	if err != nil {
		return nil, err
	}
	if last != nil {
		return nil, &ErrUnexpectedToken{
			At:       before,
			Expected: nil,
			Actual:   last,
		}
	}

	return mainFrag, nil
}
