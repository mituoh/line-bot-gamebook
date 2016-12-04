package lgscript

import (
	"bufio"
	"bytes"
	"io"
)

// Scanner represents a lexical scanner.
type Scanner struct {
	r *bufio.Reader
}

// NewScanner returns a new instance of Scanner.
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{r: bufio.NewReader(r)}
}

// Scan returns the next token and literal value.
func (s *Scanner) Scan() (tok Token, lit string) {
	// Read the next rune.
	ch := s.read()

	// If we see whitespace then consume all contiguous whitespace.
	// If we see a letter then consume as an text or reserved word.
	if ch == eof {
		return EOF, ""
	} else if isWhitespace(ch) {
		s.unread()
		return s.scanWhitespace()
	} else if isAddress(ch) {
		s.unread()
		return s.scanAddress()
	} else if isBracket(ch) {
		s.unread()
		return s.scanBracket()
	} else if isCommand(ch) {
		s.unread()
		return s.scanCommand()
	} else if isBranch(ch) {
		s.unread()
		return s.scanBranch()
	}

	if ch == ']' {
		return s.scanWhitespace()
	}

	s.unread()
	return s.scanText()
}

// scanWhitespace consumes the current rune and all contiguous whitespace.
func (s *Scanner) scanWhitespace() (tok Token, lit string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent whitespace character into the buffer.
	// Non-whitespace characters and EOF will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isWhitespace(ch) {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}

	return WS, buf.String()
}

func (s *Scanner) scanAddress() (tok Token, lit string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent text character into the buffer.
	// Non-text characters and EOF will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof || isEOL(ch) {
			break
		} else if !isLetter(ch) && !isDigit(ch) {
			s.unread()
			break
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}

	// Otherwise return as a regular text.
	return ADDRESS, buf.String()
}

func (s *Scanner) scanBracket() (tok Token, lit string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent text character into the buffer.
	// Non-text characters and EOF will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isLetter(ch) && !isDigit(ch) {
			s.unread()
			break
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}

	// Otherwise return as a regular text.
	return BRACKET, buf.String()
}

func (s *Scanner) scanCommand() (tok Token, lit string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent text character into the buffer.
	// Non-text characters and EOF will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isLetter(ch) && !isDigit(ch) {
			s.unread()
			break
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}

	// If the string matches a keyword then return that keyword.
	switch buf.String() {
	case "@buttons":
		return BUTTONS, buf.String()
	case "@end":
		return END, buf.String()
	case "@wait":
		return WAIT, buf.String()
	}

	return WS, buf.String()
}

func (s *Scanner) scanBranch() (tok Token, lit string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent text character into the buffer.
	// Non-text characters and EOF will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isLetter(ch) && !isDigit(ch) {
			s.unread()
			break
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}

	// Otherwise return as a regular text.
	return BRANCH, buf.String()
}

// scanText consumes the current rune and all contiguous text runes.
func (s *Scanner) scanText() (tok Token, lit string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent text character into the buffer.
	// Non-text characters and EOF or EOL will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof || isEOL(ch) {
			break
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}

	// Otherwise return as a regular text.
	return TEXT, buf.String()
}

// read reads the next rune from the bufferred reader.
// Returns the rune(0) if an error occurs (or io.EOF is returned).
func (s *Scanner) read() rune {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}
	return ch
}

// unread places the previously read rune back on the reader.
func (s *Scanner) unread() { _ = s.r.UnreadRune() }

// isWhitespace returns true if the rune is a space, tab or new line.
func isWhitespace(ch rune) bool { return ch == ' ' || ch == '\t' || ch == '\n' }

// isEOL returns true if the rune is a end of line.
func isEOL(ch rune) bool { return ch == '\n' }

// isAddress returns true if the rune is a address.
func isAddress(ch rune) bool { return ch == '*' }

// isBracket returns true if the rune is a goto branch.
func isBracket(ch rune) bool { return ch == '[' }

// isCommand returns true if the rune is a command.
func isCommand(ch rune) bool { return ch == '@' }

// isBranch returns true if the rune is a brunch.
func isBranch(ch rune) bool { return ch == '-' }

// isLetter returns true if the rune is a letter.
func isLetter(ch rune) bool { return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') }

// isDigit returns true if the rune is a digit.
func isDigit(ch rune) bool { return (ch >= '0' && ch <= '9') }

// eof represents a marker rune for the end of the reader.
var eof = rune(0)
