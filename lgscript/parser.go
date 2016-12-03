package lgscript

import "io"

// Script represents a LGScript.
type Script struct {
	Text   string
	Action struct {
		Token   Token
		Branch1 Branch
		Branch2 Branch
	}
}

// Branch in Script struct
type Branch struct {
	Address string
	Text    string
}

// Parser represents a parser.
type Parser struct {
	s   *Scanner
	buf struct {
		tok Token  // last read token
		lit string // last read literal
		n   int    // buffer size (max=1)
	}
}

// NewParser returns a new instance of Parser.
func NewParser(r io.Reader) *Parser {
	return &Parser{s: NewScanner(r)}
}

// Parse parses a LGScript statement.
func (p *Parser) Parse(a string) ([]Script, error) {
	scripts := []Script{}
	oldScripts := []Script{}
	// pf is ProcessingFlag
	pf := false

	for {
		tok, lit := p.scanIgnoreWhitespace()
		switch tok {
		case TEXT:
			// println("TEXT : " + lit)
			script := &Script{Text: lit}
			scripts = append(scripts, *script)
		case ADDRESS:
			// println("ADDRESS : " + lit)
			// Search address
			if a == lit {
				// println("match address")
				// Initialize scripts
				scripts = oldScripts
				pf = true
			}
		case BRACKET:
			_, lit := p.scanIgnoreWhitespace()
			// println("GOTO-ADDRESS : " + lit)
			if pf {
				pf = false
				a = lit
				oldScripts = scripts
			}
		case BUTTONS:
			script := &Script{}
			// println("BUTTONS : " + lit)
			script.Action.Token = tok
			_, litTEXT := p.scanIgnoreWhitespace()
			script.Text = litTEXT
			// println("\tTEXT : " + litTEXT)

			// BRANCH
			// First branch
			p.scanIgnoreWhitespace() // "-"
			p.scanIgnoreWhitespace() // "["
			_, lit = p.scanIgnoreWhitespace()
			// println("\tBRANCH-GOTO-ADDRESS : " + lit)
			script.Action.Branch1.Address = lit
			_, litTEXT = p.scanIgnoreWhitespace()
			// println("\t\tTEXT : " + litTEXT)
			script.Action.Branch1.Text = litTEXT

			// Second branch
			p.scanIgnoreWhitespace() // "-"
			p.scanIgnoreWhitespace() // "["
			_, lit = p.scanIgnoreWhitespace()
			// println("\tBRANCH-GOTO-ADDRESS : " + lit)
			script.Action.Branch2.Address = lit
			_, litTEXT = p.scanIgnoreWhitespace()
			// println("\t\tTEXT : " + litTEXT)
			script.Action.Branch2.Text = litTEXT

			// END
			t, _ := p.scanIgnoreWhitespace()
			if t == END {
				// println("END")
				scripts = append(scripts, *script)
			}
			if pf {
				return scripts, nil
			}
		case END:
			// println("END")
			if pf {
				return scripts, nil
			}
		case EOF:
			// println("EOF")
			return scripts, nil
		}
	}
}

// scanIgnoreWhitespace scans the next non-whitespace token.
func (p *Parser) scanIgnoreWhitespace() (tok Token, lit string) {
	tok, lit = p.s.Scan()
	if tok == WS {
		tok, lit = p.s.Scan()
	}
	return
}
