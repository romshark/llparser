package parser_test

import (
	"bytes"
	"fmt"
	"testing"

	llp "github.com/romshark/llparser"
	"github.com/stretchr/testify/require"
)

func TestPrintFragment(t *testing.T) {
	test := func(
		t *testing.T,
		options llp.FragPrintOptions,
		expectation string,
	) {
		pr := llp.NewParser()
		src := newSource("abcdef")
		mainFrag, err := pr.Parse(src, &llp.Rule{
			Kind: llp.FragmentKind(100),
			Pattern: llp.Sequence{
				llp.Exact{
					Kind:        llp.FragmentKind(101),
					Expectation: []rune("abc"),
				},
				llp.Exact{
					Kind:        llp.FragmentKind(102),
					Expectation: []rune("def"),
				},
			},
		}, nil)
		require.NoError(t, err)
		require.NotNil(t, mainFrag)

		// Override output target
		bf := &bytes.Buffer{}
		options.Out = bf

		bytesWritten, err := llp.PrintFragment(mainFrag, options)
		require.NoError(t, err)
		require.Equal(t, expectation, bf.String())
		require.Equal(t, len(expectation), bytesWritten)
	}

	t.Run("Default", func(t *testing.T) {
		test(
			t,
			llp.FragPrintOptions{},
			"100 (test.txt: 1:1-1:7 'abcdef') {"+
				" 101 (test.txt: 1:1-1:4 'abc')"+
				" 102 (test.txt: 1:4-1:7 'def') }",
		)
	})

	t.Run("Prefix_Indentation_LineBreak", func(t *testing.T) {
		test(
			t,
			llp.FragPrintOptions{
				Prefix:      []byte("***"),
				Indentation: []byte("--"),
				LineBreak:   []byte("\r\n"),
			},
			"***100 (test.txt: 1:1-1:7 'abcdef') {\r\n"+
				"***--101 (test.txt: 1:1-1:4 'abc')\r\n"+
				"***--102 (test.txt: 1:4-1:7 'def')\r\n"+
				"***}",
		)
	})

	t.Run("CustomHeadFmt", func(t *testing.T) {
		headFmt := func(token *llp.Token) []byte {
			switch int(token.VKind) {
			case 100:
				return []byte("First")
			case 101:
				return []byte("Second")
			}
			return []byte(fmt.Sprintf("T(%d)", int(token.VKind)))
		}
		test(
			t,
			llp.FragPrintOptions{HeadFmt: headFmt},
			"First { Second T(102) }",
		)
	})
}
