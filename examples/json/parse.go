package main

import (
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

	// FrLiteral Json Boolean Datatype
	FrLiteral
)

// Parse parses a dick-lang file
// This parser was written with the JSON RFC Spec (https://tools.ietf.org/html/rfc7159) as reference
func Parse(fileName string, source []rune) (*ModelJSON, error) {

	// Initialize model
	mod := &ModelJSON{}

	// Define the grammar

	ruleNull := &parser.Rule{
		Designation: "Rule for null",
		Kind:        FrNull,
		Pattern: parser.TermExact{
			Kind:        misc.FrWord,
			Expectation: []rune("null"),
		},
		Action: mod.onJSONDetected,
	}

	// ruleString := &parser.Rule{
	// 	Designation: "string",
	// 	Kind:        FrDick,
	// 	Pattern: parser.Sequence{
	// 		parser.TermExact{Kind: FrLiteral, Expectation: "\""},
	// 		parser.Either{
	// 			parser.TermExact{Kind: FrBalls, Expectation: "8"},
	// 			parser.TermExact{Kind: FrBalls, Expectation: "B"},
	// 		},
	// 		ruleShaft,
	// 		parser.TermExact{Kind: FrLiteral, Expectation: "\""},
	// 	},
	// 	Action: mod.onDickDetected,
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
		Pattern: parser.Either{
			ruleNull,
		},
	}

	// Initialize lexer and parser
	par := parser.NewParser()
	lex := misc.NewLexer(&parser.SourceFile{
		Name: fileName,
		Src:  source,
	})

	// Parse the source file
	mainFrag, err := par.Parse(lex, ruleFile)
	if err != nil {
		return nil, err
	}

	mod.Frag = mainFrag
	return mod, nil
}
