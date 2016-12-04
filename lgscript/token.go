package lgscript

// Token represents a lexical token.
type Token int

const (
	// - Special tokens

	// ILLEGAL : ILLEGAL token
	ILLEGAL Token = iota
	// EOF : End of file token
	EOF
	// WS  : White space token
	WS

	// TEXT : text token
	TEXT

	// - Keywords

	// ADDRESS : address token
	ADDRESS // *
	// BRACKET : command token
	BRACKET // [
	// BUTTONS : question point token
	BUTTONS // @buttons
	// BRANCH : branch point token
	BRANCH // -
	// END : branch end point token
	END // @end
	// WAIT : wait token
	WAIT
)
