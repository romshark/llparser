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
	Pattern: llparser.Exact{
		Kind:        101,
		Expectation: []rune("string"),
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
	Pattern: llparser.Exact{
		Kind:        101,
		Expectation: []rune("string"),
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
	llparser.Exact{Expectation: "="},
	llparser.Optional{Pattern: rule}, // potential recursion
}
```

### Terminals

#### Pattern: Exact
`Exact` expects a particular sequence of characters to be lexed:

```go
Pattern: llparser.Exact{
	Kind:        SomeKindConstant,
	Expectation: []rune("some string"),
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
		Pattern: llparser.Exact{
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

### The Parse-Tree

A parse-tree defines the serialized representation of the parsed input stream and consists of `Fragment` interfaces represented by the main fragment returned by `llparser.Parse`. A fragment is a typed chunk of the source code pointing to a start and end position in the source file, defining the *kind* of the chunk and referring to its child-fragments.

### Error-Handling

Normally, when the parser fails to match the provided grammar it returns an
`ErrUnexpectedToken` error which is rather generic and doesn't reflect the actual
mistake with a comprehensive error message. To improve the quality of the returned
error messages an error-rule can be provided which the parser tries to match at
the position of an unexpected token. If the error-rule is matched successfully
the error returned by the matched `Action` callback is returned. If the error-rule
doesn't match then the default `ErrUnexpectedToken` error is returned as usual.

```go
grammar := &parser.Rule{
	Pattern: parser.Sequence{
		parser.Exact{Expectation: []rune("foo")},
		parser.Exact{Expectation: []rune("...")},
	},
}

errRule := &parser.Rule{
	Pattern: parser.Either{
		parser.OneOrMore{
			Pattern: parser.Exact{Expectation: []rune(";")},
		},
		parser.OneOrMore{
			Pattern: parser.Exact{Expectation: []rune(".")},
		},
	},
	Action: func(fr parser.Fragment) error {
		// Return a convenient error message instead of a generic one
		return fmt.Errorf("expected 3 dots, got %d", len(fr.Src()))
	},
}

mainFrag, err := pr.Parse(src, grammar, errRule)
if err != nil {
	log.Fatal("Parser error: ", err)
}
```