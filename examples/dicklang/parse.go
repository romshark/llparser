package main

import (
	"errors"

	parser "github.com/romshark/llparser"
)

// FragKind represents a dick-lang fragment kind
type FragKind = parser.FragmentKind

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

// Parse parses a dick-lang file
func Parse(fileName string, source []rune) (*ModelDicks, error) {

	// Initialize model
	mod := &ModelDicks{}

	// Define the grammar
	termHeadLeft := parser.Exact{Kind: FrHead, Expectation: []rune("<")}
	termHeadRight := parser.Exact{Kind: FrHead, Expectation: []rune(">")}
	termBalls1 := parser.Exact{Kind: FrBalls, Expectation: []rune("8")}
	termBallsRight1 := parser.Exact{Kind: FrBalls, Expectation: []rune("B")}
	termBallsLeft1 := parser.Exact{Kind: FrBalls, Expectation: []rune("3")}

	termSpace := parser.Lexed{
		Fn: func(crs parser.Cursor) uint {
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

	ruleShaft := &parser.Rule{
		Designation: "shaft",
		Kind:        FrShaft,
		Pattern: parser.OneOrMore{
			Pattern: parser.Either{
				parser.Exact{Expectation: []rune("=")},
				parser.Exact{Expectation: []rune(":")},
				parser.Exact{Expectation: []rune("x")},
			},
		},
	}

	ruleDickRight := &parser.Rule{
		Designation: "dick(right)",
		Kind:        FrDick,
		Pattern: parser.Sequence{
			parser.Either{
				termBalls1,
				termBallsRight1,
			},
			ruleShaft,
			termHeadRight,
		},
		Action: mod.onDickDetected,
	}

	ruleDickLeft := &parser.Rule{
		Designation: "dick(left)",
		Kind:        FrDick,
		Pattern: parser.Sequence{
			termHeadLeft,
			ruleShaft,
			parser.Either{
				termBalls1,
				termBallsLeft1,
			},
		},
		Action: mod.onDickDetected,
	}

	ruleFile := &parser.Rule{
		Designation: "file",
		Pattern: parser.Sequence{
			parser.Optional{Pattern: termSpace},
			parser.ZeroOrMore{
				Pattern: parser.Sequence{
					parser.Either{
						ruleDickLeft,
						ruleDickRight,
					},
					parser.Optional{Pattern: termSpace},
				},
			},
		},
	}

	// Define error patterns
	errRule := &parser.Rule{
		Pattern: parser.Either{
			// Dick (right) without a head
			&parser.Rule{
				Pattern: parser.Sequence{
					parser.Either{
						termBalls1,
						termBallsRight1,
					},
					ruleShaft,
				},
				Action: func(parser.Fragment) error {
					return errors.New("that dick is missing a head")
				},
			},
			// Dick (left) without a head
			&parser.Rule{
				Pattern: parser.Sequence{
					ruleShaft,
					parser.Either{
						termBalls1,
						termBallsLeft1,
					},
				},
				Action: func(parser.Fragment) error {
					return errors.New("that dick is missing a head")
				},
			},
			// Dick (right) without balls
			&parser.Rule{
				Pattern: parser.Sequence{
					ruleShaft,
					termHeadRight,
				},
				Action: func(parser.Fragment) error {
					return errors.New("that dick is missing its balls")
				},
			},
			// Dick (left) without balls
			&parser.Rule{
				Pattern: parser.Sequence{
					termHeadLeft,
					ruleShaft,
				},
				Action: func(parser.Fragment) error {
					return errors.New("that dick is missing its balls")
				},
			},
		},
	}

	// Initialize lexer and parser
	par := parser.NewParser()

	// Parse the source file
	mainFrag, err := par.Parse(&parser.SourceFile{
		Name: fileName,
		Src:  source,
	}, ruleFile, errRule)
	if err != nil {
		return nil, err
	}

	mod.Frag = mainFrag
	return mod, nil
}
