<a href="https://travis-ci.org/romshark/llparser">
	<img src="https://travis-ci.org/romshark/llparser.svg?branch=master" alt="Travis CI: build status">
</a>
<a href='https://coveralls.io/github/romshark/llparser'>
	<img src='https://coveralls.io/repos/github/romshark/llparser/badge.svg' alt='Coverage Status' />
</a>
<a href="https://goreportcard.com/report/github.com/romshark/llparser">
	<img src="https://goreportcard.com/badge/github.com/romshark/llparser" alt="GoReportCard">
</a>
<a href="https://godoc.org/github.com/romshark/llparser">
	<img src="https://godoc.org/github.com/romshark/llparser?status.svg" alt="GoDoc">
</a>

<h2>
	<span>A Universal Dynamic <a href="https://en.wikipedia.org/wiki/LL_parser">LL(*)</a> Parser</span>
	<br>
	<sub>written in and for the <a href="https://golang.org/">Go programming language</a>.</sub>
</h2>

[romshark/llparser](https://github.com/romshark/llparser) is a dynamic [recursive-descent top-down](https://en.wikipedia.org/wiki/Recursive_descent_parser) parser which parses any given input stream by trying to recursively match the root rule.
It's universal in that it supports any [LL(*) grammar](https://en.wikipedia.org/wiki/LL_grammar).

This library allows building parsers in Go with relatively good error messages and flexible, even dynamic LL(*) grammars which may mutate at runtime. It parses the input stream into a typed [parse-tree](https://en.wikipedia.org/wiki/Parse_tree) and allows action hooks to be executed when a particular rule is matched.

## Getting Started

### Rules

A grammar always begins with a root rule. A rule is a [non-terminal symbol](https://en.wikipedia.org/wiki/Terminal_and_nonterminal_symbols#Nonterminal_symbols). Non-terminals are nodes of the parse-tree that consist of other non-terminals or terminals while terminals are leaf-nodes. A rule consists of a `Designation`, a `Pattern`, a `Kind` and an `Action`:
```go
mainRule := &llparser.Rule{
	Designation: "name of the rule",
	Kind: 100,
	Pattern: llparser.TermExact{
		Kind:        101,
		Expectation: "string",
	},
	Action: func(f llparser..Fragment) error {
		log.Print("the rule was successfuly matched!")
		return nil
	},
}
```
- `Designation` defines the optional logical name of the rule and is used for debugging and error reporting purposes.
- `Kind` defines the type identifier of the rule. If this field isn't set then zero (untyped) is used by default.
- `Pattern` defines the expected pattern of the rule. This field is required.
- `Action` defines the optional callback which is executed when this rule is matched. The action callback may return an error which will make the parser stop and fail immediately.

Rules can be nested:
```go
ruleTwo := &llparser.Rule{
	Pattern: llparser.TermExact{
		Kind:        101,
		Expectation: "string",
	},
}

ruleOne := &llparser.Rule{
	Pattern: ruleTwo,
}
```

Rules can also recurse:

```go
rule := &llparser.Rule{Kind: 1}
rule.Pattern = llparser.Sequence{
	llparser.TermExact{Expectation: "="},
	llparser.Optional{Pattern: rule}, // potential recursion
}
```

### Terminals

#### Pattern: Term
`Term` expects a particular fragment kind to be lexed:

```go
Pattern: llparser.Term(SomeKindConstant),
```

#### Pattern: TermExact
`TermExact` expects a particular sequence of characters to be lexed:

```go
Pattern: llparser.TermExact{
	Kind:        SomeKindConstant,
	Expectation: "some string",
},
```

#### Pattern: Checked
`Checked` expects the lexed fragment to pass an arbitrary user-defined validation function:

```go
Pattern: llparser.Checked{
	Designation: "some checked terminal",
	Fn:          func(str string) bool { return len(str) > 5 },
},
```

#### Pattern: Lexed
`Lexed` tries to lex an arbitrary sequence of characters according to `Fn`:

```go
Pattern: llparser.Lexed{
	Designation: "some lexed terminal",
	Kind:        SomeKindConstant,
	Fn: func(crs llparser.Cursor) uint {
		if crs.File.Src[crs.Index] == '|' {
			return 0
		}
		return 1
	},
},
```
`Fn` returns either `0` for ending the sequence, `1` for advancing for 1 rune or any positive integer _n_ to advance for _n_ runes.

### Combinators

#### Pattern: Optional
`Optional` tries to match a specific pattern but doesn't expect it to be matched:

```go
Pattern: llparser.Optional{
	Pattern: somePattern,
},
```

#### Pattern: Sequence
`Sequence` expects a specific sequence of patterns:

```go
Pattern: llparser.Sequence{
	somePattern,
	llparser.Term(SomeKindConstant),
	llparser.Optional{
		Pattern: llparser.TermExact{
			Kind:        SomeKindConstant,
			Expectation: "foo",
		},
	},
},
```

#### Pattern: ZeroOrMore
`ZeroOrMore` tries to match one or many instances of a specific pattern but doesn't expect it to be matched:

```go
Pattern: llparser.ZeroOrMore{
	Pattern: somePattern,
},
```

#### Pattern: OneOrMore
`OneOrMore` expects at least one or many instances of a specific pattern:

```go
Pattern: llparser.OneOrMore{
	Pattern: somePattern,
},
```

#### Pattern: Either
`Either` expects either of the given patterns selecting the first match:

```go
Pattern: llparser.Either{
	somePattern,
	anotherPattern,
},
```
### Lexer

The lexer is an abstract part of the parser which [tokenizes](https://en.wikipedia.org/wiki/Lexical_analysis#Tokenization) the input stream:

```go
type Lexer interface {
	Read() (*Token, error)
	ReadExact(
		expectation string,
		kind FragmentKind,
	) (
		token *Token,
		matched bool,
		err error,
	)
	Position() Cursor
	Set(Cursor)
}
```

A [default lexer implementation](https://github.com/romshark/llparser/tree/master/misc) is available out-of-the-box but sometimes implementing your own parser makes more sense. Some examples for when a custom lexer might be useful:
- Sometimes you don't care how many white-spaces, tabs and line-breaks there are between the patterns you care about and thus it doesn't make any sense to make each individual space character a terminal leaf-node, instead the lexer would read a sequence of whitespaces, tabs and line-breaks as a single typed terminal node (in fact, this is the behavior of the default lexer implementation linked aboved, it will treat these kinds of sequences as `misc.FrSpace`) reducing the complexity of the resulting parse-tree.
- If you want to disallow certain kinds of runes in the source code you can make the custom lexer implementation return an `ErrUnexpectedToken` error when approaching one.

### The Parse-Tree

A parse-tree defines the serialized representation of the parsed input stream and consists of `Fragment` interfaces represented by the main fragment returned by `llparser.Parse`. A fragment is a typed chunk of the source code pointing to a start and end position in the source file, defining the *kind* of the chunk and referring to its child-fragments.
