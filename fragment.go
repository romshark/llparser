package main

import "parser/parser"

// FragmentKind represents the kind of a fragment
type FragmentKind = parser.FragmentKind

const (
	_ FragmentKind = iota

	// FrTkSpace represents either a whitespace, a tab or a line-break
	FrTkSpace

	// FrTkLatinAlphanum represents a sequence of alphanumeric latin letters
	FrTkLatinAlphanum

	// FrTkSymLeftCurlyBracket represents the '{' symbol
	FrTkSymLeftCurlyBracket

	// FrTkSymRightCurlyBracket represents the '}' symbol
	FrTkSymRightCurlyBracket

	// FrTkSymLeftParenthesis represents the '(' symbol
	FrTkSymLeftParenthesis

	// FrTkSymRightParenthesis represents the ')' symbol
	FrTkSymRightParenthesis

	// FrTkSymEq represents the '=' symbol
	FrTkSymEq

	// FrTkIdnType represents a type identifier token
	FrTkIdnType

	// FrSchemaFile represents a schema file
	FrSchemaFile

	// FrDeclTypeAlias represents an alias type declaration
	FrDeclTypeAlias

	// FrDeclTraitImpl represents a trait implementation declaration
	FrDeclTraitImpl

	// FrSchemaFileHeader represents a schema file header
	FrSchemaFileHeader
)

// FragmentKindDesignation stringifies a fragment kind
func FragmentKindDesignation(fk FragmentKind) string {
	switch fk {
	case FrTkSpace:
		return "T_Space"
	case FrTkLatinAlphanum:
		return "T_LatinAlphanum"
	case FrTkSymLeftCurlyBracket:
		return "T_Sym_LeftCurlyBracket"
	case FrTkSymRightCurlyBracket:
		return "T_Sym_RightCurlyBracket"
	case FrTkSymLeftParenthesis:
		return "T_Sym_LeftParenthesis"
	case FrTkSymRightParenthesis:
		return "T_Sym_RightParenthesis"
	case FrTkSymEq:
		return "T_Sym_Eq"
	case FrTkIdnType:
		return "T_Idn_Type"
	case FrSchemaFile:
		return "C_SchemaFile"
	case FrDeclTypeAlias:
		return "C_DeclTypeAlias"
	case FrDeclTraitImpl:
		return "C_DeclTraitImpl"
	case FrSchemaFileHeader:
		return "C_SchemaFileHeader"
	}
	return "<unknown>"
}
