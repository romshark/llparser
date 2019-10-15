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
) (frag Fragment, err error) {
	switch pt := pattern.(type) {
	case *Rule:
		frag, err = pr.parseRule(scan.New(), pt)
		if err, ok := err.(*ErrUnexpectedToken); ok {
			// Override expected pattern to the higher-order rule
			err.Expected = pt
		}

	case Exact:
		if scan.Lexer.reachedEOF() {
			return nil, errEOF{}
		}
		// Exact terminal
		frag, err = pr.parseExact(scan, pt)

	case Lexed:
		if scan.Lexer.reachedEOF() {
			return nil, errEOF{}
		}
		frag, err = pr.parseLexed(scan, pt)

	case Repeated:
		err = pr.parseRepeated(scan, pt.Min, pt.Max, pt.Pattern)

	case Sequence:
		// Sequence
		err = pr.parseSequence(scan, pt)

	case Either:
		// Choice
		frag, err = pr.parseEither(scan, pt)

	case Not:
		// Expect no match
		err = pr.parseNot(scan, pt)

	default:
		panic(fmt.Errorf(
			"unsupported pattern type: %s",
			reflect.TypeOf(pattern),
		))
	}
	return
}

func (pr Parser) parseNot(scan *scanner, ptr Not) error {
	beforeCr := scan.Lexer.cr
	_, err := pr.handlePattern(scan, ptr.Pattern)
	switch err := err.(type) {
	case *ErrUnexpectedToken:
		scan.Set(beforeCr)
		return nil
	case errEOF:
		scan.Set(beforeCr)
		return nil
	case nil:
		return &ErrUnexpectedToken{
			At:       beforeCr,
			Expected: ptr,
		}
	default:
		return err
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

func (pr Parser) parseRepeated(
	scanner *scanner,
	min uint,
	max uint,
	pattern Pattern,
) error {
	if max != 0 && min > max {
		panic(fmt.Errorf(
			"min (%d) > max (%d) while parsing pattern Repeated(%s)",
			min,
			max,
			pattern.Desig(),
		))
	}

	num := uint(0)
	lastPosition := scanner.Lexer.cr
	for {
		if max != 0 && num >= max {
			break
		}

		frag, err := pr.handlePattern(scanner, pattern)
		switch err := err.(type) {
		case *ErrUnexpectedToken:
			if min != 0 && num < min {
				// Mismatch before the minimum is read
				return err
			}
			// Reset scanner to the last match
			scanner.Set(lastPosition)
			return nil

		case errEOF:
			if min != 0 && num < min {
				// Mismatch before the minimum is read
				return &ErrUnexpectedToken{
					At:       scanner.Lexer.cr,
					Expected: pattern,
				}
			}
			// Reset scanner to the last match
			scanner.Set(lastPosition)
			return nil

		case nil:
			num++
			lastPosition = scanner.Lexer.cr
			// Append rule patterns, other patterns are appended automatically
			if !pattern.Container() {
				scanner.Append(pattern, frag)
			}

		default:
			return err
		}
	}

	return nil
}

func (pr Parser) parseSequence(
	scanner *scanner,
	patterns Sequence,
) error {
	for _, pt := range patterns {
		frag, err := pr.handlePattern(scanner, pt)
		if err != nil {
			return err
		}
		// Append rule patterns, other patterns are appended automatically
		if !pt.Container() {
			scanner.Append(pt, frag)
		}
	}
	return nil
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

func (pr Parser) parseExact(
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
) (frag Fragment, err error) {
	frag, err = pr.handlePattern(scanner, rule.Pattern)
	if err != nil {
		return
	}
	if !rule.Pattern.Container() {
		scanner.Append(rule.Pattern, frag)
	}
	frag = scanner.Fragment(rule.Kind)

	if rule.Action != nil {
		// Execute rule action callback
		if err := rule.Action(frag); err != nil {
			return nil, &Err{Err: err, At: frag.Begin()}
		}
	}
	return
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
	switch err := err.(type) {
	case errEOF:
		// Ignore EOF errors
		return mainFrag, nil
	case nil:
	default:
		// Report unexpected errors
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
