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
	})
	require.NoError(t, err)
	require.NotNil(t, mainFrag)

	bf := &bytes.Buffer{}
	_, err = parser.PrintFragment(mainFrag, bf, "  ")
	require.NoError(t, err)

	expected := "100(test.txt: 1:1-1:7 'abcdef') {\n" +
		"  101(test.txt: 1:1-1:4 'abc')\n" +
		"  102(test.txt: 1:4-1:7 'def')\n}"
	require.Equal(t, expected, bf.String())
}
