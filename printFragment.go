package parser

import (
	"fmt"
	"io"
	"os"
	"reflect"
)

var snipBlkStart = []byte(" {")
var snipBlkEnd = []byte("}")
var snipLineBreak = []byte("\n")
var snipSpace = []byte(" ")

// FragPrintOptions defines the AST printing options
type FragPrintOptions struct {
	Out         io.Writer
	Indentation []byte
	Prefix      []byte
	LineBreak   []byte
	Format      func(Fragment) (head, body []byte)
}

// PrintFragment prints the fragment structure recursively
func PrintFragment(
	fragment Fragment,
	options FragPrintOptions,
) (bytesWritten int, err error) {
	if options.Out == nil {
		// Use stdout by default
		options.Out = os.Stdout
	}

	// write returns true if there was an error,
	// otherwise returns false
	write := func(str []byte) bool {
		var bw int
		bw, err = options.Out.Write(str)
		if err != nil {
			// Abort printing
			return true
		}
		bytesWritten += bw
		// Continue printing
		return false
	}

	writePrefix := func() bool {
		// Write the prefix if any
		if len(options.Prefix) > 0 {
			return write(options.Prefix)
		}
		return false
	}

	writeLnBrk := func() bool {
		if len(options.Indentation) < 1 {
			// Write whitespace instead when indentation is disabled
			return write(snipSpace)
		}

		// Write line-break
		if options.LineBreak != nil {
			// Use specified line-breaks
			if write(options.LineBreak) {
				return true
			}
		} else {
			// Fallback to the default line-breaks
			if write(snipLineBreak) {
				return true
			}
		}

		return writePrefix()
	}

	printIndent := func(ind uint) bool {
		// Write the indentation
		if len(options.Indentation) > 0 {
			for ix := uint(0); ix < ind; ix++ {
				if write(options.Indentation) {
					return true
				}
			}
		}
		return false
	}

	var printFrag func(ind uint, frag Fragment) bool
	printFrag = func(ind uint, frag Fragment) bool {
		if printIndent(ind) {
			return true
		}

		// Write the actual line
		switch frag := frag.(type) {
		case *Construct:
			// Print construct recursively
			var head, body []byte
			if options.Format != nil {
				// Use specified stringification method
				head, body = options.Format(frag)
			}
			if head == nil {
				// Fallback to the default stringification method
				head = []byte(frag.String())
			}
			if write(head) {
				return true
			}
			if body != nil {
				if write(body) {
					return true
				}
				break
			}
			if write(snipBlkStart) {
				return true
			}
			if writeLnBrk() {
				return true
			}
			for _, subFrag := range frag.VElements {
				if printFrag(ind+1, subFrag) {
					return true
				}
				if writeLnBrk() {
					return true
				}
			}
			if printIndent(ind) {
				return true
			}
			return write(snipBlkEnd)
		case *Token:
			// Print leave fragment
			var head []byte
			if options.Format != nil {
				// Use specified stringification method
				head, _ = options.Format(frag)
			}
			if head == nil {
				// Fallback to the default stringification method
				head = []byte(frag.String())
			}
			if write(head) {
				return true
			}
		default:
			panic(fmt.Errorf(
				"unsupported fragment type: %s",
				reflect.TypeOf(frag),
			))
		}
		return false
	}

	if writePrefix() {
		return
	}
	printFrag(0, fragment)
	return
}
