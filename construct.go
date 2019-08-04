package parser

// Construct represents a composite fragment
type Construct struct {
	*Token
	VElements []Fragment
}

// Elements returns the element fragments of the construct fragment
func (ct *Construct) Elements() []Fragment { return ct.VElements }
