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

	// FrName represents a json key
	FrName

	// FrString Json String Datatype
	FrString

	// FrNumber Json Number Datatype; both decimal and floats
	FrNumber

	// FrArray Json Array Datatype
	FrArray

	// FrObject Json Object Datatype
	FrObject

	// FrNull Json Null Datatype
	FrNull

	// FrBoolean Json Boolean Datatype
	FrBoolean
)

// Parse parses a dick-lang file
// This parser was written with the JSON RFC Spec (https://tools.ietf.org/html/rfc7159) as reference
func Parse(filePath string) (*ModelJSON, error) {
	// Read the source file into memory
	bt, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Initialize model
	mod := &ModelJSON{}

	// Define the grammar

	ruleNull := &parser.Rule{
		Designation: "string",
		Kind:        FrString,
		Pattern: parser.TermExact{
			Kind: misc.FrWord, Expectation: "null",
		},
	}

	// ruleString := &parser.Rule{
	// 	Designation: "string",
	// 	Kind:        FrString,
	// 	Pattern: parser.TermExact{
	// 		Pattern: parser.Either{
	// 			parser.TermExact{Kind: misc.FrSign, Expectation: "="},
	// 			parser.TermExact{Kind: misc.FrSign, Expectation: ":"},
	// 			parser.TermExact{Kind: misc.FrWord, Expectation: "x"},
	// 		},
	// 	},
	// }

	// ruleNumber := &parser.Rule{
	// 	Designation: "string",
	// 	Kind:        FrString,
	// 	Pattern: parser.TermExact{
	// 		Pattern: parser.Either{
	// 			parser.TermExact{Kind: misc.FrSign, Expectation: "="},
	// 			parser.TermExact{Kind: misc.FrSign, Expectation: ":"},
	// 			parser.TermExact{Kind: misc.FrWord, Expectation: "x"},
	// 		},
	// 	},
	// }

	// ruleBoolean := &parser.Rule{
	// 	Designation: "boolean",
	// 	Kind:        FrBoolean,
	// 	Pattern: parser.TermExact{
	// 		Pattern: parser.Either{
	// 			parser.TermExact{Kind: misc.FrSign, Expectation: "true"},
	// 			parser.TermExact{Kind: misc.FrSign, Expectation: "false"},
	// 		},
	// 	},
	// }

	// ruleDickRight := &parser.Rule{
	// 	Designation: "dick(right)",
	// 	Kind:        FrDick,
	// 	Pattern: parser.Sequence{
	// 		parser.Either{
	// 			parser.TermExact{Kind: FrBalls, Expectation: "8"},
	// 			parser.TermExact{Kind: FrBalls, Expectation: "B"},
	// 		},
	// 		ruleShaft,
	// 		parser.TermExact{Kind: FrHead, Expectation: ">"},
	// 	},
	// 	Action: mod.onDickDetected,
	// }

	// ruleDickLeft := &parser.Rule{
	// 	Designation: "dick(left)",
	// 	Kind:        FrDick,
	// 	Pattern: parser.Sequence{
	// 		parser.TermExact{Kind: FrHead, Expectation: "<"},
	// 		ruleShaft,
	// 		parser.Either{
	// 			parser.TermExact{Kind: FrBalls, Expectation: "8"},
	// 			parser.TermExact{Kind: FrBalls, Expectation: "3"},
	// 		},
	// 	},
	// 	Action: mod.onDickDetected,
	// }

	ruleFile := &parser.Rule{
		Designation: "file",
		Pattern: parser.Sequence{
			parser.Optional{Pattern: parser.Term(misc.FrSpace)},
			parser.ZeroOrMore{
				Pattern: parser.Sequence{
					parser.Either{
						ruleNull,
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
