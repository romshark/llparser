package main

import (
	"fmt"

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
	shaftLength := uint(len(frag.Elements()[1].Elements()))

	// Check dick length
	if shaftLength < 2 {
		return fmt.Errorf(
			"sorry, Improper JSON format (%d/2)",
			shaftLength,
		)
	}

	// Register the newly json structure
	mod.JSON = append(mod.JSON, ModelJSONObject{
		Frag:        frag,
		ShaftLength: shaftLength,
	})

	return nil
}
