package parser

import llp "github.com/romshark/llparser"

// ModelDicks represents the model of a dicks source file
type ModelDicks struct {
	Frag  llp.Fragment
	Dicks []ModelDick
}

// ModelDick represents the model of a dick expression
type ModelDick struct {
	Frag        llp.Fragment
	ShaftLength uint
}

func (mod *ModelDicks) onDickDetected(frag llp.Fragment) error {
	// Register the newly parsed dick
	mod.Dicks = append(mod.Dicks, ModelDick{
		Frag:        frag,
		ShaftLength: uint(len(frag.Elements()[1].Elements())),
	})

	return nil
}
