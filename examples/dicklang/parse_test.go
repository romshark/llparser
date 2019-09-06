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
	require.Equal(t, "B===>", string(mod.Dicks[0].Frag.Src()))
	require.Equal(t, "8==>", string(mod.Dicks[1].Frag.Src()))
	require.Equal(t, "B::>", string(mod.Dicks[2].Frag.Src()))
	require.Equal(t, "<====8", string(mod.Dicks[3].Frag.Src()))
	require.Equal(t, "<::::::3", string(mod.Dicks[4].Frag.Src()))
	require.Equal(t, "8xxxx>", string(mod.Dicks[5].Frag.Src()))
	require.Equal(t, "8xxx=xxx>", string(mod.Dicks[6].Frag.Src()))
	require.Equal(t, "B:x:=:x>", string(mod.Dicks[7].Frag.Src()))
	require.Equal(t, "<:=3", string(mod.Dicks[8].Frag.Src()))
}
