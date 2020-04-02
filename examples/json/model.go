package main

import (
	parser "github.com/romshark/llparser"
)

// ModelJSON represents the model of a json source file
type ModelJSON struct {
	Frag parser.Fragment
	JSON []ModelJSONObject
}

// ModelJSONObject represents the model of a dick expression
type ModelJSONObject struct {
	Frag        parser.Fragment
	ShaftLength uint
}

func (mod *ModelJSON) onJSONDetected(frag parser.Fragment) error {
	shaftLength := uint(len(frag.Elements()[0].Elements()))

	// Register the new json object
	mod.JSON = append(mod.JSON, ModelJSONObject{
		Frag:        frag,
		ShaftLength: shaftLength,
	})

	return nil
}
