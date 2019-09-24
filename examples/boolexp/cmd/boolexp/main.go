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

	prs := parser.NewParser()
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
				HeadFmt: func(tk *llp.Token) []byte {
					return []byte(parser.FragKindString(tk.VKind))
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
