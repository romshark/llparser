package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParser(t *testing.T) {
	// Change the source code however you like
	var src = `
	B===> 8==>  B::>
		<====8 <::::::3
			8xxxx> 8xxx=xxx>
		B:x:=:x>
	 <:=3
	`
	mod, err := Parse("sample.dicklang", []rune(src))
	require.NoError(t, err)
	require.Len(t, mod.Dicks, 9)
	require.Equal(t, mod.Dicks[0].Frag.Src(), []rune("B===>"))
	require.Equal(t, mod.Dicks[1].Frag.Src(), []rune("8==>"))
	require.Equal(t, mod.Dicks[2].Frag.Src(), []rune("B::>"))
	require.Equal(t, mod.Dicks[3].Frag.Src(), []rune("<====8"))
	require.Equal(t, mod.Dicks[4].Frag.Src(), []rune("<::::::3"))
	require.Equal(t, mod.Dicks[5].Frag.Src(), []rune("8xxxx>"))
	require.Equal(t, mod.Dicks[6].Frag.Src(), []rune("8xxx=xxx>"))
	require.Equal(t, mod.Dicks[7].Frag.Src(), []rune("B:x:=:x>"))
	require.Equal(t, mod.Dicks[8].Frag.Src(), []rune("<:=3"))
}
