package parser

import (
	"fmt"
	"io"
	"reflect"
)

// PrintFragment prints the fragment structure recursively
// to the given output stream
func PrintFragment(
	fragment Fragment,
	out io.Writer,
	indent string,
) (bytesWritten int, err error) {
	indb := []byte(indent)

	print := func(ind uint, str string) bool {
		for ix := uint(0); ix < ind; ix++ {
			btw, err := out.Write(indb)
			if err != nil {
				return true
			}
			bytesWritten += btw
		}
		btw, err := out.Write([]byte(str))
		if err != nil {
			return true
		}
		bytesWritten += btw
		return false
	}

	var printFrag func(uint, Fragment) bool
	printFrag = func(ind uint, frag Fragment) bool {
		switch fr := frag.(type) {
		case *Construct:
			if print(ind, fmt.Sprintf("%s {\n", fr.Token)) {
				return true
			}
			for _, elem := range fr.VElements {
				if printFrag(ind+1, elem) {
					return true
				}
				if print(ind, "\n") {
					return true
				}
			}
			if print(ind, "}") {
				return true
			}
		case *Token:
			if print(ind, fr.String()) {
				return true
			}
		default:
			panic(fmt.Errorf(
				"unsupported fragment type %s",
				reflect.TypeOf(frag),
			))
		}
		return false
	}

	printFrag(0, fragment)
	return
}
