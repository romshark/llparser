package parser

// scanner represents a sequence-recording lexing source-code scanner
type scanner struct {
	Lexer   *lexer
	Records []Fragment
}

// newScanner creates a new scanner instance
func newScanner(lexer *lexer) *scanner {
	if lexer == nil {
		panic("missing lexer during scanner initialization")
	}
	return &scanner{Lexer: lexer}
}

// New creates a new scanner succeeding the original one
// dropping its record history
func (sc *scanner) New() *scanner {
	return &scanner{Lexer: sc.Lexer}
}

// ReadExact advances the scanner by 1 exact token returning either the read
// fragment or nil if the expectation didn't match
func (sc *scanner) ReadExact(
	expectation []rune,
	kind FragmentKind,
) (*Token, bool, error) {
	nx, match, err := sc.Lexer.ReadExact(expectation, kind)
	if err != nil {
		return nil, false, err
	}
	if nx == nil {
		return nil, match, nil
	}
	sc.Records = append(sc.Records, nx)
	return nx, match, nil
}

// ReadUntil advances the scanner by 1 exact token returning either the read
// fragment or nil if the expectation didn't match
func (sc *scanner) ReadUntil(
	fn func(Cursor) uint,
	kind FragmentKind,
) (*Token, error) {
	nx, err := sc.Lexer.ReadUntil(fn, kind)
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
func (sc *scanner) Append(
	pattern Pattern,
	fragment Fragment,
) {
	if fragment == nil {
		return
	}
	if _, ok := pattern.(*Rule); ok {
		sc.Records = append(sc.Records, fragment)
		return
	}
	termPt := pattern.TerminalPattern()
	if termPt != nil {
		if _, ok := termPt.(*Rule); ok {
			sc.Records = append(sc.Records, fragment)
		}
	}
}

// Fragment returns a typed composite fragment
func (sc *scanner) Fragment(kind FragmentKind) Fragment {
	if len(sc.Records) < 1 {
		pos := sc.Lexer.cr
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
func (sc *scanner) Set(cursor Cursor) {
	sc.Lexer.cr = cursor
	sc.TidyUp()
}

// TidyUp removes all records after the current position
func (sc *scanner) TidyUp() int {
	removed := 0
	pos := sc.Lexer.cr
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
