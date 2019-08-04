package main

import "parser/parser"

var rTypeIdn = &parser.Rule{
	Designation: "type identifier",
	Kind:        FrTkIdnType,
	Pattern:     parser.Checked(rCapLatAlphanum),
}

var rDeclTypeAlias = &parser.Rule{
	Designation: "alias type declaration",
	Kind:        FrDeclTypeAlias,
	Pattern: parser.Sequence{
		rTypeIdn,
		parser.Optional{Pattern: parser.Term(FrTkSpace)},
		parser.Term(FrTkSymEq),
		parser.Optional{Pattern: parser.Term(FrTkSpace)},
		rTypeIdn,
	},
	Action: onDeclTypeAlias,
}

var rDeclTraitImpl = &parser.Rule{
	Designation: "trait implementation declaration",
	Kind:        FrDeclTraitImpl,
	Pattern: parser.Sequence{
		rTypeIdn,
		parser.Term(FrTkSpace),
		parser.TermExact("implements"),
		parser.Term(FrTkSpace),
		rTypeIdn,
	},
	Action: onDeclTraitImpl,
}

var rSchemaFileHeader = &parser.Rule{
	Designation: "schema file header",
	Kind:        FrSchemaFileHeader,
	Pattern: parser.Sequence{
		parser.TermExact("schema"),
		parser.Term(FrTkSpace),
		parser.Checked(rLowLatAlphanum),
	},
}

var rSchemaFile = &parser.Rule{
	Designation: "schema file",
	Kind:        FrSchemaFile,
	Pattern: parser.Sequence{
		rSchemaFileHeader,
		parser.Term(FrTkSpace),
		parser.OneOrMore{
			Pattern: parser.Either{
				rDeclTypeAlias,
				rDeclTraitImpl,
			},
		},
	},
}

func rCapLatAlphanum(str string) bool {
	if len(str) < 1 {
		return false
	}
	if !isLatinUpperCase(str[0]) {
		return false
	}
	for i := 1; i < len(str); i++ {
		if !isLatinAlphanum(str[i]) {
			return false
		}
	}
	return true
}

func rLowLatAlphanum(str string) bool {
	if len(str) < 1 {
		return false
	}
	if !isLatinUpperCase(str[0]) {
		return false
	}
	for i := 1; i < len(str); i++ {
		if !isLatinAlphanum(str[i]) {
			return false
		}
	}
	return true
}
