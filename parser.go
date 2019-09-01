package parser

import (
	"errors"
	"fmt"
	"reflect"
)

// Lexer defines the interface of an abstract lexer implementation
type Lexer interface {
	Read() (*Token, error)

	ReadExact(
		expectation []rune,
		kind FragmentKind,
	) (
		token *Token,
		matched bool,
		err error,
	)

	ReadUntil(
		fn func(Cursor) uint,
		kind FragmentKind,
	) (
		token *Token,
		err error,
	)

	Position() Cursor

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
		}
		if err != nil {
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
	case Lexed:
		return pr.parseLexed(scanner, pt)
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
		return pr.parseEither(scanner, pt)
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
	tk, err := scanner.Read()
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

func (pr Parser) parseLexed(
	scanner *Scanner,
	expected Lexed,
) (Fragment, error) {
	beforeCr := scanner.Lexer.Position()
	tk, err := scanner.ReadUntil(expected.Fn, expected.Kind)
	if err != nil {
		return nil, err
	}
	if tk == nil {
		return nil, &ErrUnexpectedToken{
			At:       beforeCr,
			Expected: expected,
		}
	}
	return tk, nil
}

func (pr Parser) parseOptional(
	scanner *Scanner,
	pattern Pattern,
) (Fragment, error) {
	beforeCr := scanner.Lexer.Position()
	frag, err := pr.handlePattern(scanner, pattern)
	if err != nil {
		if _, ok := err.(*ErrUnexpectedToken); ok {
			// Reset scanner to the initial position
			scanner.Set(beforeCr)
			return nil, nil
		}
		return nil, err
	}
	// Append rule patterns, other patterns are appended automatically
	if !pattern.Container() {
		scanner.Append(pattern, frag)
	}
	return frag, nil
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
		if !pattern.Container() {
			scanner.Append(pattern, frag)
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
			if _, ok := err.(*ErrUnexpectedToken); ok {
				if num < 1 {
					// No matches so far but already a mismatch
					return nil, err
				}
				// Reset scanner to the last match
				scanner.Set(lastPosition)
				return nil, nil
			}
			return nil, err
		}
		num++
		lastPosition = scanner.Lexer.Position()
		// Append rule patterns, other patterns are appended automatically
		if !pattern.Container() {
			scanner.Append(pattern, frag)
		}
	}
}

func (pr Parser) parseSequence(
	scanner *Scanner,
	patterns Sequence,
) (Fragment, error) {
	for _, pt := range patterns {
		frag, err := pr.handlePattern(scanner, pt)
		if err != nil {
			return nil, err
		}
		// Append rule patterns, other patterns are appended automatically
		if !pt.Container() {
			scanner.Append(pt, frag)
		}
	}
	return nil, nil
}

func (pr Parser) parseEither(
	scanner *Scanner,
	patternOptions Either,
) (Fragment, error) {
	beforeCr := scanner.Lexer.Position()
	for ix, pt := range patternOptions {
		lastOption := ix >= len(patternOptions)-1

		frag, err := pr.handlePattern(scanner, pt)
		if err != nil {
			if er, ok := err.(*ErrUnexpectedToken); ok {
				if lastOption {
					// Set actual expected pattern
					er.Expected = patternOptions
				} else {
					// Reset scanner to the initial position
					scanner.Set(beforeCr)
					// Continue checking other options
					continue
				}
			}
			return nil, err
		}
		// Append rule patterns, other patterns are appended automatically
		if !pt.Container() {
			scanner.Append(pt, frag)
		}
		return frag, nil
	}
	return nil, nil
}

func (pr Parser) parseTermExact(
	scanner *Scanner,
	exact TermExact,
) (Fragment, error) {
	beforeCr := scanner.Lexer.Position()
	tk, match, err := scanner.ReadExact(
		exact.Expectation,
		exact.Kind,
	)
	if err != nil {
		return nil, err
	}
	if !match {
		return nil, &ErrUnexpectedToken{
			At:       beforeCr,
			Expected: exact,
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
	tk, err := scanner.Read()
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
		return nil, err
	}
	if !rule.Pattern.Container() {
		scanner.Append(rule.Pattern, frag)
	}
	composedFrag := scanner.Fragment(rule.Kind)

	if rule.Action != nil {
		// Execute rule action callback
		if err := rule.Action(composedFrag); err != nil {
			return nil, &Err{Err: err, At: composedFrag.Begin()}
		}
	}

	return composedFrag, nil
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
	last, err := lexer.Read()
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
