package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/romshark/llparser/examples/boolexp/parser"
)

var flagSrc = flag.String("e", "", "boolean expression")
var flagPrintAST = flag.Bool("ast", false, "print AST only")

func main() {
	flag.Parse()

	prs := parser.NewParser()
	ast, err := prs.Parse("example.boolexp", []rune(*flagSrc))
	if err != nil {
		log.Fatalf("ERR: %s", err)
	}

	if *flagPrintAST {
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

	fmt.Printf("%s = %t", *flagSrc, ast.Root.Val())
}
