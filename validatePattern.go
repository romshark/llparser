package parser

import (
	"fmt"
	"reflect"
)

func validateRule(
	ptr *Rule,
	validated map[Pattern]struct{},
) error {
	if ptr.Pattern == nil {
		return fmt.Errorf("rule %p is missing a pattern", ptr)
	}
	return validatePattern(ptr.Pattern, validated)
}

func validateSequence(
	ptr Sequence,
	validated map[Pattern]struct{},
) error {
	if len(ptr) < 1 {
		return fmt.Errorf("sequence is empty")
	}
	for _, elem := range ptr {
		if err := validatePattern(elem, validated); err != nil {
			return err
		}
	}
	return nil
}

func validateRepeated(
	ptr *Repeated,
	validated map[Pattern]struct{},
) error {
	if ptr.Max != 0 && ptr.Min > ptr.Max {
		return fmt.Errorf(
			"repeated %p min (%d) greater max (%d)",
			ptr,
			ptr.Min,
			ptr.Max,
		)
	}
	if ptr.Pattern == nil {
		return fmt.Errorf("repeated %p is missing a pattern", ptr)
	}
	return validatePattern(ptr.Pattern, validated)
}

func validateEither(
	ptr Either,
	validated map[Pattern]struct{},
) error {
	if len(ptr) < 2 {
		return fmt.Errorf("either-combinator has %d option(s)", len(ptr))
	}

	options := map[Pattern]struct{}{}
	ix := 0

	for _, opt := range ptr {
		// Check for duplicate options
		checkDuplicate := false
		switch opt.(type) {
		case *Rule:
			checkDuplicate = true
		case *Lexed:
			checkDuplicate = true
		case *Exact:
			checkDuplicate = true
		case *Repeated:
			checkDuplicate = true
		}

		if checkDuplicate {
			if _, ok := options[opt]; ok {
				return fmt.Errorf(
					"either-combinator has duplicate options (at index %d)",
					ix,
				)
			}
			options[opt] = struct{}{}
		}
		if err := validatePattern(opt, validated); err != nil {
			return err
		}

		ix++
	}
	return nil
}

func validateNot(
	ptr Not,
	validated map[Pattern]struct{},
) error {
	if _, ok := ptr.Pattern.(Not); ok {
		return fmt.Errorf("not-combinator is nested")
	}
	return validatePattern(ptr.Pattern, validated)
}

func validateLexed(ptr *Lexed) error {
	if ptr.Fn == nil {
		return fmt.Errorf("lexed-terminal %p is missing the lexer function", ptr)
	}
	return nil
}

func validateExact(ptr *Exact) error {
	if len(ptr.Expectation) < 1 {
		return fmt.Errorf("exact-terminal %p is missing an expectation", ptr)
	}
	return nil
}

// ValidatePattern recursively validates the given pattern
func ValidatePattern(ptr Pattern) error {
	return validatePattern(ptr, nil)
}

func validatePattern(ptr Pattern, validated map[Pattern]struct{}) error {
	if validated == nil {
		validated = map[Pattern]struct{}{}
	}

	isValidated := func() bool {
		if _, isValidated := validated[ptr]; !isValidated {
			validated[ptr] = struct{}{}
			return false
		}
		return true
	}

	switch ptr := ptr.(type) {
	case nil:
		return nil
	case *Rule:
		if isValidated() {
			return nil
		}
		if err := validateRule(ptr, validated); err != nil {
			return err
		}
	case Sequence:
		if err := validateSequence(ptr, validated); err != nil {
			return err
		}
	case *Repeated:
		if isValidated() {
			return nil
		}
		if err := validateRepeated(ptr, validated); err != nil {
			return err
		}
	case Either:
		if err := validateEither(ptr, validated); err != nil {
			return err
		}
	case Not:
		if err := validateNot(ptr, validated); err != nil {
			return err
		}
	case *Lexed:
		if isValidated() {
			return nil
		}
		if err := validateLexed(ptr); err != nil {
			return err
		}
	case *Exact:
		if isValidated() {
			return nil
		}
		if err := validateExact(ptr); err != nil {
			return err
		}
	default:
		return fmt.Errorf(
			"unsupported pattern type: %s",
			reflect.TypeOf(ptr),
		)
	}

	return nil
}
