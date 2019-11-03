package parser

type recursionRegister map[*Rule]uint

// Reset resets all register counters to 0
func (rr recursionRegister) Reset() {
	for rl := range rr {
		rr[rl] = uint(0)
	}
}

// findRules finds all rules in a pattern recursively
func findRules(pattern Pattern, reg recursionRegister) {
	if reg == nil {
		reg = recursionRegister{}
	}
	switch pt := pattern.(type) {
	case *Rule:
		if pt == nil {
			return
		}
		if _, ok := reg[pt]; ok {
			return
		}
		reg[pt] = 0
		findRules(pt.Pattern, reg)
	case Sequence:
		for _, pt := range pt {
			findRules(pt, reg)
		}
	case Either:
		for _, pt := range pt {
			findRules(pt, reg)
		}
	case Not:
		findRules(pt.Pattern, reg)
	case *Repeated:
		if pt == nil {
			return
		}
		findRules(pt.Pattern, reg)
	}
}
