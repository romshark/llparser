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

// New creates a new scanner succeeding the original one
// dropping its record history
func (sc *Scanner) New() *Scanner {
	return &Scanner{Lexer: sc.Lexer}
}

// Read advances the scanner by 1 token returning either the read fragment
// or an error if the lexer failed
func (sc *Scanner) Read() (*Token, error) {
	nx, err := sc.Lexer.Read()
	if err != nil {
		return nil, err
	}
	if nx == nil {
		return nil, nil
	}
	sc.Records = append(sc.Records, nx)
	return nx, nil
}

// ReadExact advances the scanner by 1 exact token returning either the read
// fragment or nil if the expectation didn't match
func (sc *Scanner) ReadExact(
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
func (sc *Scanner) ReadUntil(
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
func (sc *Scanner) Append(
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
