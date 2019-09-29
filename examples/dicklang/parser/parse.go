package parser

import (
	"errors"

	llp "github.com/romshark/llparser"
)

// FragKind represents a dick-lang fragment kind
type FragKind = llp.FragmentKind

const (
	_ FragKind = iota

	// FrSpace represents an arbitrary sequence of spaces, tabs and line-breaks
	FrSpace

	// FrBalls represents the balls
	FrBalls

	// FrShaft represents the shaft
	FrShaft

	// FrHead represents the head
	FrHead

	// FrDick represents the entire dick
	FrDick
)

// FragKindString translates the kind identifier to its name
func FragKindString(kind llp.FragmentKind) string {
	switch kind {
	case FrSpace:
		return "Space"
	case FrBalls:
		return "Balls"
	case FrShaft:
		return "Shaft"
	case FrHead:
		return "Head"
	case FrDick:
		return "Dick"
	}
	return ""
}

// Parse parses a dick-lang file
func Parse(fileName string, source []rune) (*ModelDicks, error) {

	// Initialize model
	mod := &ModelDicks{}

	// Define the grammar
	termHeadLeft := llp.Exact{Kind: FrHead, Expectation: []rune("<")}
	termHeadRight := llp.Exact{Kind: FrHead, Expectation: []rune(">")}
	termBalls1 := llp.Exact{Kind: FrBalls, Expectation: []rune("8")}
	termBallsRight1 := llp.Exact{Kind: FrBalls, Expectation: []rune("B")}
	termBallsLeft1 := llp.Exact{Kind: FrBalls, Expectation: []rune("3")}

	termSpace := llp.Lexed{
		Fn: func(crs llp.Cursor) uint {
			switch crs.File.Src[crs.Index] {
			case ' ':
				return 1
			case '\t':
				return 1
			case '\n':
				return 1
			case '\r':
				next := crs.Index + 1
				if next < uint(len(crs.File.Src)) &&
					crs.File.Src[next] == '\n' {
					return 2
				}
			}
			return 0
		},
		Kind: FrSpace,
	}

	shaftElement := llp.Either{
		llp.Exact{Expectation: []rune("=")},
		llp.Exact{Expectation: []rune(":")},
		llp.Exact{Expectation: []rune("x")},
	}

	ruleShaft := &llp.Rule{
		Designation: "shaft",
		Kind:        FrShaft,
		Pattern:     llp.Repeated{Pattern: shaftElement, Min: 2},
	}

	ruleDickRight := &llp.Rule{
		Designation: "dick(right)",
		Kind:        FrDick,
		Pattern: llp.Sequence{
			llp.Either{
				termBalls1,
				termBallsRight1,
			},
			ruleShaft,
			termHeadRight,
		},
		Action: mod.onDickDetected,
	}

	ruleDickLeft := &llp.Rule{
		Designation: "dick(left)",
		Kind:        FrDick,
		Pattern: llp.Sequence{
			termHeadLeft,
			ruleShaft,
			llp.Either{
				termBalls1,
				termBallsLeft1,
			},
		},
		Action: mod.onDickDetected,
	}

	ruleFile := &llp.Rule{
		Designation: "file",
		Pattern: llp.Sequence{
			llp.Repeated{
				Min:     0,
				Max:     1,
				Pattern: termSpace,
			},
			llp.Repeated{
				Pattern: llp.Sequence{
					llp.Either{
						ruleDickLeft,
						ruleDickRight,
					},
					llp.Repeated{
						Min:     0,
						Max:     1,
						Pattern: termSpace,
					},
				},
			},
		},
	}

	// Define error patterns
	errRule := &llp.Rule{
		Pattern: llp.Either{
			// Dick (right) without a head
			&llp.Rule{
				Pattern: llp.Sequence{
					llp.Either{
						termBalls1,
						termBallsRight1,
					},
					ruleShaft,
				},
				Action: func(llp.Fragment) error {
					return errors.New("that dick is missing a head")
				},
			},
			// Dick (left) without a head
			&llp.Rule{
				Pattern: llp.Sequence{
					ruleShaft,
					llp.Either{
						termBalls1,
						termBallsLeft1,
					},
				},
				Action: func(llp.Fragment) error {
					return errors.New("that dick is missing a head")
				},
			},
			// Dick (right) without balls
			&llp.Rule{
				Pattern: llp.Sequence{
					ruleShaft,
					termHeadRight,
				},
				Action: func(llp.Fragment) error {
					return errors.New("that dick is missing its balls")
				},
			},
			// Dick (left) without balls
			&llp.Rule{
				Pattern: llp.Sequence{
					termHeadLeft,
					ruleShaft,
				},
				Action: func(llp.Fragment) error {
					return errors.New("that dick is missing its balls")
				},
			},
			// Dick (right) too small
			&llp.Rule{
				Pattern: llp.Sequence{
					llp.Either{
						termBalls1,
						termBallsRight1,
					},
					shaftElement,
					termHeadRight,
				},
				Action: func(llp.Fragment) error {
					return errors.New("that dick is too small")
				},
			},
			// Dick (left) too small
			&llp.Rule{
				Pattern: llp.Sequence{
					termHeadLeft,
					shaftElement,
					llp.Either{
						termBalls1,
						termBallsLeft1,
					},
				},
				Action: func(llp.Fragment) error {
					return errors.New("that dick is too small")
				},
			},
		},
	}

	// Initialize lexer and parser
	par := llp.NewParser()

	// Parse the source file
	mainFrag, err := par.Parse(&llp.SourceFile{
		Name: fileName,
		Src:  source,
	}, ruleFile, errRule)
	if err != nil {
		return nil, err
	}

	mod.Frag = mainFrag
	return mod, nil
}
