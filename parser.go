package parser

import (
	"errors"
	"fmt"
	"reflect"
)

// Parser represents a parser
type Parser struct{}

// NewParser creates a new parser instance
func NewParser() Parser {
	return Parser{}
}

func (pr Parser) handlePattern(
	scan *scanner,
	pattern Pattern,
) (Fragment, error) {
	switch pt := pattern.(type) {
	case *Rule:
		tk, err := pr.parseRule(scan.New(), pt)
		if err, ok := err.(*ErrUnexpectedToken); ok {
			// Override expected pattern to the higher-order rule
			err.Expected = pt
			return nil, err
		}
		if err != nil {
			return nil, err
		}
		return tk, nil
	case Exact:
		// Exact terminal
		return pr.parseTermExact(scan, pt)
	case Lexed:
		return pr.parseLexed(scan, pt)
	case ZeroOrMore:
		// ZeroOrMore
		return pr.parseZeroOrMore(scan, pt.Pattern)
	case OneOrMore:
		// OneOrMore
		return pr.parseOneOrMore(scan, pt.Pattern)
	case Optional:
		// Optional
		return pr.parseOptional(scan, pt.Pattern)
	case Sequence:
		// Sequence
		return pr.parseSequence(scan, pt)
	case Either:
		// Choice
		return pr.parseEither(scan, pt)
	default:
		panic(fmt.Errorf(
			"unsupported pattern type: %s",
			reflect.TypeOf(pattern),
		))
	}
}

func (pr Parser) parseLexed(
	scanner *scanner,
	expected Lexed,
) (Fragment, error) {
	beforeCr := scanner.Lexer.cr
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
	scanner *scanner,
	pattern Pattern,
) (Fragment, error) {
	beforeCr := scanner.Lexer.cr
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
	scanner *scanner,
	pattern Pattern,
) (Fragment, error) {
	lastPosition := scanner.Lexer.cr
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
		lastPosition = scanner.Lexer.cr
		// Append rule patterns, other patterns are appended automatically
		if !pattern.Container() {
			scanner.Append(pattern, frag)
		}
	}
}

func (pr Parser) parseOneOrMore(
	scanner *scanner,
	pattern Pattern,
) (Fragment, error) {
	num := 0
	lastPosition := scanner.Lexer.cr
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
		lastPosition = scanner.Lexer.cr
		// Append rule patterns, other patterns are appended automatically
		if !pattern.Container() {
			scanner.Append(pattern, frag)
		}
	}
}

func (pr Parser) parseSequence(
	scanner *scanner,
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
	scanner *scanner,
	patternOptions Either,
) (Fragment, error) {
	beforeCr := scanner.Lexer.cr
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
	scanner *scanner,
	exact Exact,
) (Fragment, error) {
	beforeCr := scanner.Lexer.cr
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
		}
	}
	return tk, nil
}

func (pr Parser) parseRule(
	scanner *scanner,
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

func (pr Parser) tryErrRule(
	lex *lexer,
	errRule *Rule,
	previousUnexpErr error,
) error {
	if errRule != nil {
		_, err := pr.parseRule(newScanner(lex), errRule)
		if err == nil {
			// Return the previous error when no error was returned
			return previousUnexpErr
		}
		if err, ok := err.(*ErrUnexpectedToken); ok {
			// Reset expected token for the error-rule
			err.Expected = nil
		}
		return err
	}
	return nil
}

// Parse parses the given rule
func (pr Parser) Parse(
	source *SourceFile,
	rule *Rule,
	errRule *Rule,
) (Fragment, error) {
	if source == nil {
		return nil, errors.New("missing source file")
	}
	if rule == nil {
		return nil, errors.New("missing main grammar rule")
	}

	lex := &lexer{cr: NewCursor(source)}

	mainFrag, err := pr.parseRule(newScanner(lex), rule)
	if err != nil {
		if err, ok := err.(*ErrUnexpectedToken); ok {
			// Reset the lexer to the start position of the error
			lex.cr = err.At
		}
		if err := pr.tryErrRule(lex, errRule, err); err != nil {
			return nil, err
		}
		return nil, err
	}

	// Ensure EOF
	last, err := lex.ReadUntil(
		func(Cursor) uint { return 1 },
		0,
	)
	if err != nil {
		return nil, err
	}
	if last != nil {
		if errRule != nil {
			// Try to match an error-pattern
			lex.cr = last.VBegin
		}

		unexpErr := &ErrUnexpectedToken{At: last.VBegin}

		if err := pr.tryErrRule(lex, errRule, unexpErr); err != nil {
			return nil, err
		}

		// Fallback to default unexpected-token error
		return nil, unexpErr
	}

	return mainFrag, nil
}
