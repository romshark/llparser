package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	llp "github.com/romshark/llparser"
	"github.com/romshark/llparser/examples/boolexp/parser"
)

var flagSrc = flag.String("e", "", "boolean expression")
var flagPrintAST = flag.Bool("ast", false, "print AST")
var flagPrintParseTree = flag.Bool("ptree", false, "print parse-tree")

func main() {
	flag.Parse()

	prs, err := parser.NewParser()
	if err != nil {
		log.Fatalf("fatal: %s", err)
	}

	ast, err := prs.Parse("example.boolexp", []rune(*flagSrc))
	if err != nil {
		log.Fatalf("ERR: %s", err)
	}

	if *flagPrintAST {
		// Print the abstract syntax tree
		fmt.Println("")
		fmt.Println("AST:")
		if _, err := ast.Print(parser.ASTPrintOptions{
			Out:         os.Stdout,
			Indentation: []byte(" "),
			Prefix:      []byte(" "),
		}); err != nil {
			log.Fatalf("ERR: %s", err)
		}
		fmt.Println("")
		fmt.Println("")
	}

	if *flagPrintParseTree {
		// Print the parse tree
		fmt.Println("")
		fmt.Println("PARSE TREE:")
		if _, err := llp.PrintFragment(
			ast.Root.Fragment,
			llp.FragPrintOptions{
				Out:         os.Stdout,
				Indentation: []byte(" "),
				Prefix:      []byte(" "),
				Format: func(frag llp.Fragment) (head, body []byte) {
					head = []byte(parser.FragKindString(frag.Kind()))
					return
				},
			},
		); err != nil {
			log.Fatalf("ERR: %s", err)
		}
		fmt.Println("")
		fmt.Println("")
	}

	fmt.Printf("%s = %t", *flagSrc, ast.Root.Val())
}
