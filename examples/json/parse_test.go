// This is an example for the github.com/romshark/llparser library
package main

import (
	"fmt"
	"log"
	"os"
	"testing"

	parser "github.com/romshark/llparser"
	_ "github.com/romshark/llparser/misc"
)

// Set printParseTree to true to print the parse-tree
// instead of the model
var printParseTree = false

func TestParser(t *testing.T) {

	// Change the source code however you like
	var src = `
	"demo"
	`
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
	l := len(mod.JSON)
	if l != 9 {
		t.Errorf("Number of Dicks parsed was incorrect , got: %d, want: %d.", l, 1)
	}

	// Print all parsed dicks
	fmt.Printf("%d JSON parsed:\n", l)
	// for ix, dick := range mod.Dicks {
	// 	fmt.Printf(" %d: %s\n", ix, dick.Frag.Src())
	// }
}