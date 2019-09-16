package main

import (
	parser "github.com/romshark/llparser"
)

// ModelDicks represents the model of a dicks source file
type ModelDicks struct {
	Frag  parser.Fragment
	Dicks []ModelDick
}

// ModelDick represents the model of a dick expression
type ModelDick struct {
	Frag        parser.Fragment
	ShaftLength uint
}

func (mod *ModelDicks) onDickDetected(frag parser.Fragment) error {
	// Register the newly parsed dick
	mod.Dicks = append(mod.Dicks, ModelDick{
		Frag:        frag,
		ShaftLength: uint(len(frag.Elements()[1].Elements())),
	})

	return nil
}
