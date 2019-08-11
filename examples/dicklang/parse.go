package main

import (
	"io/ioutil"

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
func Parse(filePath string) (*ModelDicks, error) {
	// Read the source file into memory
	bt, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Initialize model
	mod := &ModelDicks{}

	// Define the grammar
	ruleShaft := &parser.Rule{
		Designation: "shaft",
		Kind:        FrShaft,
		Pattern: parser.OneOrMore{
			Pattern: parser.Either{
				parser.TermExact{Kind: misc.FrSign, Expectation: "="},
				parser.TermExact{Kind: misc.FrSign, Expectation: ":"},
				parser.TermExact{Kind: misc.FrWord, Expectation: "x"},
			},
		},
	}

	ruleDickRight := &parser.Rule{
		Designation: "dick(right)",
		Kind:        FrDick,
		Pattern: parser.Sequence{
			parser.Either{
				parser.TermExact{Kind: FrBalls, Expectation: "8"},
				parser.TermExact{Kind: FrBalls, Expectation: "B"},
			},
			ruleShaft,
			parser.TermExact{Kind: FrHead, Expectation: ">"},
		},
		Action: mod.onDickDetected,
	}

	ruleDickLeft := &parser.Rule{
		Designation: "dick(left)",
		Kind:        FrDick,
		Pattern: parser.Sequence{
			parser.TermExact{Kind: FrHead, Expectation: "<"},
			ruleShaft,
			parser.Either{
				parser.TermExact{Kind: FrBalls, Expectation: "8"},
				parser.TermExact{Kind: FrBalls, Expectation: "3"},
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
		Name: filePath,
		Src:  string(bt),
	})

	// Parse the source file
	mainFrag, err := par.Parse(lex, ruleFile)
	if err != nil {
		return nil, err
	}

	mod.Frag = mainFrag
	return mod, nil
}
