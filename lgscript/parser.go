package lgscript

import "io"

// Script represents a LGScript.
type Script struct {
	Text   string
	Action struct {
		Token    Token
		Branches [4]Branch
		Wait     string
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
			script.Action.Token = tok
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
			for i := 0; i < 4; i++ {
				_, l := p.scanIgnoreWhitespace() // "-" or "@end"
				println(l)
				if l == "-" {
					p.parseBranch(&script.Action.Branches[i].Address,
						&script.Action.Branches[i].Text)
				} else {
					p.unscan()
					break
				}
			}

			// END
			t, _ := p.scanIgnoreWhitespace()
			if t == END {
				// println("END")
				scripts = append(scripts, *script)
			}
			if pf {
				return scripts, nil
			}
		case WAIT:
			script := &Script{}
			_, lit := p.scanIgnoreWhitespace()
			script.Action.Token = tok
			script.Action.Wait = lit
			scripts = append(scripts, *script)
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

func (p *Parser) parseBranch(a *string, t *string) {
	p.scanIgnoreWhitespace() // "["
	_, lit := p.scanIgnoreWhitespace()
	*a = lit
	// println("\tBRANCH-GOTO-ADDRESS : " + lit)
	_, litTEXT := p.scanIgnoreWhitespace()
	*t = litTEXT
	// println("\t\tTEXT : " + litTEXT)
}

// scan returns the next token from the underlying scanner.
// If a token has been unscanned then read that instead.
func (p *Parser) scan() (tok Token, lit string) {
	// If we have a token on the buffer, then return it.
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.tok, p.buf.lit
	}

	// Otherwise read the next token from the scanner.
	tok, lit = p.s.Scan()

	// Save it to the buffer in case we unscan later.
	p.buf.tok, p.buf.lit = tok, lit

	return
}

// scanIgnoreWhitespace scans the next non-whitespace token.
func (p *Parser) scanIgnoreWhitespace() (tok Token, lit string) {
	tok, lit = p.scan()
	if tok == WS {
		tok, lit = p.scan()
	}
	return
}

// unscan pushes the previously read token back onto the buffer.
func (p *Parser) unscan() { p.buf.n = 1 }
