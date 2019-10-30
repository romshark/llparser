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
) (tk *Token, match bool, err error) {
	tk, match, err = sc.Lexer.ReadExact(expectation, kind)
	if err != nil || tk == nil {
		return
	}
	sc.Records = append(sc.Records, tk)
	return
}

// ReadUntil advances the scanner by 1 exact token returning either the read
// fragment or nil if the expectation didn't match
func (sc *scanner) ReadUntil(
	fn func(index uint, cursor Cursor) bool,
	kind FragmentKind,
) (tk *Token, err error) {
	tk, err = sc.Lexer.ReadUntil(fn, kind)
	if err != nil || tk == nil {
		return
	}
	sc.Records = append(sc.Records, tk)
	return
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
	if termPt := pattern.TerminalPattern(); termPt != nil {
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
func (sc *scanner) TidyUp() (removed int) {
	pos := sc.Lexer.cr
	for ix := len(sc.Records) - 1; ix >= 0; ix-- {
		if sc.Records[ix].Begin().Index < pos.Index {
			break
		}
		removed++
	}

	// Remove the last n records
	sc.Records = sc.Records[:len(sc.Records)-removed]
	return
}
