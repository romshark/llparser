package main

import (
	parser "github.com/romshark/llparser"
	"github.com/romshark/llparser/misc"
)

// FragKind represents a dick-lang fragment kind
type FragKind parser.FragmentKind

const (
	_ = misc.FrSign + iota

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
func Parse(fileName, source string) (*ModelDicks, error) {

	// Initialize model
	mod := &ModelDicks{}

	// Define the grammar
	ruleShaft := &parser.Rule{
		Designation: "shaft",
		Kind:        FrShaft,
		Pattern: parser.OneOrMore{
			Pattern: parser.Either{
				parser.TermExact{Kind: misc.FrSign, Expectation: []rune("=")},
				parser.TermExact{Kind: misc.FrSign, Expectation: []rune(":")},
				parser.TermExact{Kind: misc.FrWord, Expectation: []rune("x")},
			},
		},
	}

	ruleDickRight := &parser.Rule{
		Designation: "dick(right)",
		Kind:        FrDick,
		Pattern: parser.Sequence{
			parser.Either{
				parser.TermExact{Kind: FrBalls, Expectation: []rune("8")},
				parser.TermExact{Kind: FrBalls, Expectation: []rune("B")},
			},
			ruleShaft,
			parser.TermExact{Kind: FrHead, Expectation: []rune(">")},
		},
		Action: mod.onDickDetected,
	}

	ruleDickLeft := &parser.Rule{
		Designation: "dick(left)",
		Kind:        FrDick,
		Pattern: parser.Sequence{
			parser.TermExact{Kind: FrHead, Expectation: []rune("<")},
			ruleShaft,
			parser.Either{
				parser.TermExact{Kind: FrBalls, Expectation: []rune("8")},
				parser.TermExact{Kind: FrBalls, Expectation: []rune("3")},
			},
		},
		Action: mod.onDickDetected,
	}

	ruleFile := &parser.Rule{
		Designation: "file",
		Pattern: parser.Sequence{
			parser.Optional{Pattern: parser.Term(misc.FrSpace)},
			parser.ZeroOrMore{
				Pattern: parser.Sequence{
					parser.Either{
						ruleDickLeft,
						ruleDickRight,
					},
					parser.Optional{Pattern: parser.Term(misc.FrSpace)},
				},
			},
		},
	}

	// Initialize lexer and parser
	par := parser.NewParser()
	lex := misc.NewLexer(&parser.SourceFile{
		Name: fileName,
		Src:  []rune(source),
	})

	// Parse the source file
	mainFrag, err := par.Parse(lex, ruleFile)
	if err != nil {
		return nil, err
	}

	mod.Frag = mainFrag
	return mod, nil
}
