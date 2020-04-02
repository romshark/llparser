package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParser(t *testing.T) {
	var src = `
    null
		`
	mod, err := Parse("test.json", []rune(strings.TrimSpace(src)))
	require.NoError(t, err)
	require.Len(t, mod.JSON, 1)
	require.Equal(t, "null", string(mod.JSON[0].Frag.Src()))
}
