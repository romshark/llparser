package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParser(t *testing.T) {
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
	require.Equal(t, []rune("B===>"), mod.Dicks[0].Frag.Src())
	require.Equal(t, []rune("8==>"), mod.Dicks[1].Frag.Src())
	require.Equal(t, []rune("B::>"), mod.Dicks[2].Frag.Src())
	require.Equal(t, []rune("<====8"), mod.Dicks[3].Frag.Src())
	require.Equal(t, []rune("<::::::3"), mod.Dicks[4].Frag.Src())
	require.Equal(t, []rune("8xxxx>"), mod.Dicks[5].Frag.Src())
	require.Equal(t, []rune("8xxx=xxx>"), mod.Dicks[6].Frag.Src())
	require.Equal(t, []rune("B:x:=:x>"), mod.Dicks[7].Frag.Src())
	require.Equal(t, []rune("<:=3"), mod.Dicks[8].Frag.Src())
}
