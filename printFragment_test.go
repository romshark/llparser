package parser_test

import (
	"bytes"
	"testing"

	parser "github.com/romshark/llparser"
	"github.com/stretchr/testify/require"
)

func TestPrintFragment(t *testing.T) {
	pr := parser.NewParser()
	src := newSource("abcdef")
	mainFrag, err := pr.Parse(src, &parser.Rule{
		Kind: parser.FragmentKind(100),
		Pattern: parser.Sequence{
			parser.Exact{
				Kind:        parser.FragmentKind(101),
				Expectation: []rune("abc"),
			},
			parser.Exact{
				Kind:        parser.FragmentKind(102),
				Expectation: []rune("def"),
			},
		},
	}, nil)
	require.NoError(t, err)
	require.NotNil(t, mainFrag)

	kindTranslator := func(kind parser.FragmentKind) string {
		switch int(kind) {
		case 100:
			return "First"
		case 101:
			return "Second"
		}
		return ""
	}

	bf := &bytes.Buffer{}
	_, err = parser.PrintFragment(mainFrag, bf, "  ", kindTranslator)
	require.NoError(t, err)

	expected := "First (test.txt: 1:1-1:7 'abcdef') {\n" +
		"  Second (test.txt: 1:1-1:4 'abc')\n" +
		"  102 (test.txt: 1:4-1:7 'def')\n}"
	require.Equal(t, expected, bf.String())
}
