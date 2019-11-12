package parser

import (
	"errors"
	"fmt"
	"reflect"
)

// Parser represents a parser
type Parser struct {
	grammar           *Rule
	errGrammar        *Rule
	recursionRegister recursionRegister

	// MaxRecursionLevel defines the maximum tolerated recursion level.
	// The limitation is disabled when MaxRecursionLevel is set to 0
	MaxRecursionLevel uint
}

// NewParser creates a new parser instance
func NewParser(grammar *Rule, errGrammar *Rule) (*Parser, error) {
	if grammar == nil {
		return nil, errors.New("missing grammar")
	}
	if err := ValidatePattern(grammar); err != nil {
		return nil, fmt.Errorf("invalid grammar: %w", err)
	}
	if errGrammar != nil {
		if err := ValidatePattern(errGrammar); err != nil {
			return nil, fmt.Errorf("invalid error-grammar: %w", err)
		}
	}

	recRegister := recursionRegister{}
	findRules(grammar, recRegister)
	findRules(errGrammar, recRegister)

	return &Parser{
		grammar:           grammar,
		errGrammar:        errGrammar,
		recursionRegister: recRegister,

		// Disable recursion limitation by default
		MaxRecursionLevel: uint(0),
	}, nil
}

func (pr Parser) handlePattern(
	debug *DebugProfile,
	scan *scanner,
	pattern Pattern,
	level uint,
) (frag Fragment, err error) {
	switch pt := pattern.(type) {
	case *Rule:
		frag, err = pr.parseRule(debug, scan.New(), pt, level)
		if err, ok := err.(*ErrUnexpectedToken); ok {
			// Override expected pattern to the higher-order rule
			err.Expected = pt
		}

	case *Exact:
		frag, err = pr.parseExact(debug, scan, pt, level)

	case *Lexed:
		frag, err = pr.parseLexed(debug, scan, pt, level)

	case *Repeated:
		err = pr.parseRepeated(debug, scan, pt.Min, pt.Max, pt, level)

	case Sequence:
		err = pr.parseSequence(debug, scan, pt, level)

	case Either:
		frag, err = pr.parseEither(debug, scan, pt, level)

	case Not:
		err = pr.parseNot(debug, scan, pt, level)

	default:
		panic(fmt.Errorf(
			"unsupported pattern type: %s",
			reflect.TypeOf(pattern),
		))
	}
	return
}

func (pr Parser) parseNot(
	debug *DebugProfile,
	scan *scanner,
	ptr Not,
	level uint,
) error {
	debugIndex := debug.record(ptr, scan.Lexer.cr, level)

	beforeCr := scan.Lexer.cr
	_, err := pr.handlePattern(debug, scan, ptr.Pattern, level+1)
	switch err := err.(type) {
	case *ErrUnexpectedToken:
		scan.Set(beforeCr)
		return nil
	case errEOF:
		scan.Set(beforeCr)
		return nil
	case nil:
		debug.markMismatch(debugIndex)
		return &ErrUnexpectedToken{
			At:       beforeCr,
			Expected: ptr,
		}
	default:
		return err
	}
}

func (pr Parser) parseLexed(
	debug *DebugProfile,
	scanner *scanner,
	expected *Lexed,
	level uint,
) (Fragment, error) {
	debugIndex := debug.record(expected, scanner.Lexer.cr, level)

	if scanner.Lexer.reachedEOF() {
		debug.markMismatch(debugIndex)
		return nil, errEOF{}
	}

	beforeCr := scanner.Lexer.cr
	tk, err := scanner.ReadUntil(expected.Fn, expected.Kind)
	if err != nil {
		return nil, err
	}
	if tk == nil || tk.VEnd.Index-tk.VBegin.Index < expected.MinLen {
		debug.markMismatch(debugIndex)
		return nil, &ErrUnexpectedToken{
			At:       beforeCr,
			Expected: expected,
		}
	}
	return tk, nil
}

func (pr Parser) parseRepeated(
	debug *DebugProfile,
	scanner *scanner,
	min uint,
	max uint,
	repeated *Repeated,
	level uint,
) error {
	debugIndex := debug.record(repeated, scanner.Lexer.cr, level)

	num := uint(0)
	lastPosition := scanner.Lexer.cr
	for {
		if max != 0 && num >= max {
			break
		}

		frag, err := pr.handlePattern(
			debug,
			scanner,
			repeated.Pattern,
			level+1,
		)
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
				debug.markMismatch(debugIndex)
				return &ErrUnexpectedToken{
					At:       scanner.Lexer.cr,
					Expected: repeated,
				}
			}
			// Reset scanner to the last match
			scanner.Set(lastPosition)
			return nil

		case nil:
			num++
			lastPosition = scanner.Lexer.cr
			// Append rule patterns, other patterns are appended automatically
			if !repeated.Pattern.Container() {
				scanner.Append(repeated.Pattern, frag)
			}

		default:
			return err
		}
	}

	return nil
}

func (pr Parser) parseSequence(
	debug *DebugProfile,
	scanner *scanner,
	patterns Sequence,
	level uint,
) error {
	debugIndex := debug.record(patterns, scanner.Lexer.cr, level)

	for _, pt := range patterns {
		frag, err := pr.handlePattern(debug, scanner, pt, level+1)
		if err != nil {
			debug.markMismatch(debugIndex)
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
	debug *DebugProfile,
	scanner *scanner,
	patternOptions Either,
	level uint,
) (Fragment, error) {
	debugIndex := debug.record(patternOptions, scanner.Lexer.cr, level)

	beforeCr := scanner.Lexer.cr
	for ix, pt := range patternOptions {
		lastOption := ix >= len(patternOptions)-1

		frag, err := pr.handlePattern(debug, scanner, pt, level+1)
		if err != nil {
			if er, ok := err.(*ErrUnexpectedToken); ok {
				if lastOption {
					// Set actual expected pattern
					er.Expected = patternOptions
					debug.markMismatch(debugIndex)
				} else {
					// Reset scanner to the initial position
					scanner.Set(beforeCr)
					// Continue checking other options
					continue
				}
			} else {
				// Unexpected error
				debug.markMismatch(debugIndex)
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
	debug *DebugProfile,
	scanner *scanner,
	exact *Exact,
	level uint,
) (Fragment, error) {
	debugIndex := debug.record(exact, scanner.Lexer.cr, level)

	if scanner.Lexer.reachedEOF() {
		debug.markMismatch(debugIndex)
		return nil, errEOF{}
	}

	beforeCr := scanner.Lexer.cr
	tk, match, err := scanner.ReadExact(
		exact.Expectation,
		exact.Kind,
	)
	if err != nil {
		return nil, err
	}
	if !match {
		debug.markMismatch(debugIndex)
		return nil, &ErrUnexpectedToken{
			At:       beforeCr,
			Expected: exact,
		}
	}
	return tk, nil
}

func (pr Parser) parseRule(
	debug *DebugProfile,
	scanner *scanner,
	rule *Rule,
	level uint,
) (frag Fragment, err error) {
	debugIndex := debug.record(rule, scanner.Lexer.cr, level)

	if pr.MaxRecursionLevel > 0 {
		pr.recursionRegister[rule]++
		if pr.recursionRegister[rule] > pr.MaxRecursionLevel {
			// Max recursion level exceeded
			return nil, &Err{
				Err: fmt.Errorf(
					"max recursion level exceeded at rule %p (%q)",
					rule,
					rule.Designation,
				),
				At: scanner.Lexer.cr,
			}
		}
	}

	frag, err = pr.handlePattern(debug, scanner, rule.Pattern, level+1)
	if err != nil {
		debug.markMismatch(debugIndex)
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
	debug *DebugProfile,
	lex *lexer,
	errRule *Rule,
	previousUnexpErr error,
) error {
	if errRule != nil {
		_, err := pr.parseRule(debug, newScanner(lex), errRule, 0)
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

// Debug parses the given source file in debug mode generating a debug profile
func (pr *Parser) Debug(source *SourceFile) (*DebugProfile, Fragment, error) {
	debug := newDebugProfile()
	mainFrag, err := pr.parse(source, debug)
	return debug, mainFrag, err
}

// Parse parses the given source file.
//
// WARNING: Parse isn't safe for concurrent use and shall therefore
// not be executed by multiple goroutines concurrently!
func (pr *Parser) Parse(source *SourceFile) (Fragment, error) {
	return pr.parse(source, nil)
}

func (pr *Parser) parse(
	source *SourceFile,
	debug *DebugProfile,
) (Fragment, error) {
	if pr.MaxRecursionLevel > 0 {
		// Reset the recursion register when recursion limitation is enabled
		pr.recursionRegister.Reset()
	}
	cr := NewCursor(source)
	lex := &lexer{cr: cr}

	mainFrag, err := pr.parseRule(debug, newScanner(lex), pr.grammar, 0)
	if err != nil {
		if err, ok := err.(*ErrUnexpectedToken); ok {
			// Reset the lexer to the start position of the error
			lex.cr = err.At
		}
		if err := pr.tryErrRule(debug, lex, pr.errGrammar, err); err != nil {
			return nil, err
		}
		return nil, err
	}

	// Ensure EOF
	last, err := lex.ReadUntil(
		func(uint, Cursor) bool { return true },
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
		if pr.errGrammar != nil {
			// Try to match an error-pattern
			lex.cr = last.VBegin
		}

		unexpErr := &ErrUnexpectedToken{At: last.VBegin}

		if err := pr.tryErrRule(debug, lex, pr.errGrammar, unexpErr); err != nil {
			return nil, err
		}

		// Fallback to default unexpected-token error
		return nil, unexpErr
	}

	return mainFrag, nil
}
