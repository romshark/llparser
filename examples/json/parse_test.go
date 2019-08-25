// This is an example for the github.com/romshark/llparser library
package main

import (
	"fmt"
	"log"
	"os"

	parser "github.com/romshark/llparser"
	"github.com/romshark/llparser/misc"
)

/********************************\
	Source Code

	feel free to edit!
\********************************/

// Set printParseTree to true to print the parse-tree
// instead of the model
var printParseTree = false

// Change the source code however you like
var src = `
B===> 8==>  B::>
	<====8 <::::::3
		8xxxx> 8xxx=xxx>
	B:x:=:x>
 <:=3
`

/********************************\
	Model
\********************************/

// ModelDicks represents the model of a dicks source file
type ModelDicks struct {
	Frag  parser.Fragment
	Dicks []ModelDick
}

// ModelDick represents the model of a dick expression
type ModelDick struct {
	Frag        parser.Fragment
	ShaftLength uint
}

func (mod *ModelDicks) onDickDetected(frag parser.Fragment) error {
	shaftLength := uint(len(frag.Elements()[1].Elements()))

	// Check dick length
	if shaftLength < 2 {
		return fmt.Errorf(
			"sorry, but that dick's too small (%d/2)",
			shaftLength,
		)
	}

	// Register the newly parsed dick
	mod.Dicks = append(mod.Dicks, ModelDick{
		Frag:        frag,
		ShaftLength: shaftLength,
	})

	return nil
}

/********************************\
	Parser
\********************************/

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
func Parse(source string) (*ModelDicks, error) {
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
		Name: "playground",
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

func main() {
	mod, err := Parse(src)
	if err != nil {
		log.Fatal("ERR: ", err)
	}

	if printParseTree {
		// Print the parse-tree only
		_, err := parser.PrintFragment(mod.Frag, os.Stdout, "  ")
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	// Print all parsed dicks
	fmt.Printf("%d dicks parsed:\n", len(mod.Dicks))
	for ix, dick := range mod.Dicks {
		fmt.Printf(" %d: %s\n", ix, dick.Frag.Src())
	}
}
