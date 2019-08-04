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
		return nil
	}
	if len(sc.Records) == 1 {
		record := sc.Records[0]
		return &Token{
			VBegin: record.Begin(),
			VEnd:   record.End(),
			VKind:  kind,
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
