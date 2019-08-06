package parser

// Scanner represents a sequence-recording lexing source-code scanner
type Scanner struct {
	Lexer   Lexer
	Records []Fragment
}

// NewScanner creates a new scanner instance
func NewScanner(lexer Lexer) *Scanner {
	if lexer == nil {
		panic("missing lexer during scanner initialization")
	}
	return &Scanner{Lexer: lexer}
}

// Fork creates a new scanner branching off the original one
// without the record history of the original scanner
func (sc *Scanner) Fork() *Scanner {
	return &Scanner{Lexer: sc.Lexer.Fork()}
}

// New creates a new scanner succeeding the original one
// dropping its record history
func (sc *Scanner) New() *Scanner {
	return &Scanner{Lexer: sc.Lexer}
}

// Back moves the scanner back in the recorded history
func (sc *Scanner) Back(steps uint) {
	reci := uint(len(sc.Records))
	if reci < 1 {
		return
	}
	if steps >= uint(len(sc.Records)) {
		sc.Records = nil
		sc.Lexer.Set(NewCursor(sc.Lexer.Position().File))
		return
	}
	sc.Records = sc.Records[:reci-steps]
	last := sc.Records[len(sc.Records)-1].End()
	sc.Lexer.Set(last)
}

// Next advances the scanner by 1 token returning either the read fragment
// or an error if the lexer failed
func (sc *Scanner) Next() (*Token, error) {
	nx, err := sc.Lexer.Next()
	if err != nil {
		return nil, err
	}
	if nx == nil {
		return nil, nil
	}
	sc.Records = append(sc.Records, nx)
	return nx, nil
}

// Append appends a fragment to the records
func (sc *Scanner) Append(fragment Fragment) {
	sc.Records = append(sc.Records, fragment)
}

// Fragment returns a typed composite fragment
func (sc *Scanner) Fragment(kind FragmentKind) Fragment {
	if len(sc.Records) < 1 {
		pos := sc.Lexer.Position()
		return &Construct{
			Token: &Token{
				VBegin: pos,
				VEnd:   pos,
				VKind:  kind,
			},
			VElements: nil,
		}
	}
	begin := sc.Records[0].Begin()
	end := sc.Records[len(sc.Records)-1].End()
	return &Construct{
		Token: &Token{
			VBegin: begin,
			VEnd:   end,
			VKind:  kind,
		},
		VElements: sc.Records,
	}
}

// Set sets the scanner's lexer position and tidies up the record history
func (sc *Scanner) Set(cursor Cursor) {
	sc.Lexer.Set(cursor)
	sc.TidyUp()
}

// TidyUp removes all records after the current position
func (sc *Scanner) TidyUp() int {
	removed := 0
	pos := sc.Lexer.Position()
	for ix := len(sc.Records) - 1; ix >= 0; ix-- {
		rc := sc.Records[ix]
		if rc.Begin().Index < pos.Index {
			break
		}
		removed++
	}

	// Remove the last n records
	sc.Records = sc.Records[:len(sc.Records)-removed]
	return removed
}
