package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindRules(t *testing.T) {
	f := &Rule{Designation: "f"}
	e := &Rule{
		Designation: "e",
		Pattern:     &Repeated{Pattern: f},
	}
	f.Pattern = Not{Pattern: e}
	d := &Rule{
		Designation: "d",
		Pattern:     &Exact{},
	}
	c := &Rule{
		Designation: "c",
		Pattern:     &Lexed{},
	}
	b := &Rule{
		Designation: "b",
		Pattern:     Either{d, e},
	}
	a := &Rule{
		Designation: "a",
		Pattern:     Sequence{b, c},
	}
	main := &Rule{
		Designation: "main",
		Pattern:     a,
	}

	reg := recursionRegister{}
	findRules(main, reg)

	require.Len(t, reg, 7)
	require.Contains(t, reg, main)
	require.Contains(t, reg, a)
	require.Contains(t, reg, b)
	require.Contains(t, reg, c)
	require.Contains(t, reg, d)
	require.Contains(t, reg, e)
	require.Contains(t, reg, f)
}
