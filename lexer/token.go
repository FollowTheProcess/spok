package lexer

import "fmt"

// tokenType identifies the type of lex tokens
type tokenType int

//go:generate stringer -type=tokenType -linecomment
const (
	tokenError        tokenType = iota // Error
	tokenEOF                           // EOF
	tokenHash                          // #
	tokenComment                       // Comment
	tokenDeclare                       // :=
	tokenTask                          // task
	tokenOpenParen                     // (
	tokenCloseParen                    // )
	tokenOpenBrace                     // {
	tokenCloseBrace                    // }
	tokenOpenBracket                   // [
	tokenCloseBracket                  // ]
	tokenOutput                        // ->
)

// token represents a semantic token or lexeme
type token struct {
	value string    // The value of the token, e.g. "23"
	typ   tokenType // The token type, e.g. tokenString
	pos   int       // The starting position of this item in the input string
	line  int       // The line number at the start of this item
}

// String satisfies the stringer interface so we can pretty print our tokens
func (t token) String() string {
	switch {
	case t.typ == tokenEOF:
		return "EOF"
	case t.typ == tokenError:
		return t.value
	case len(t.value) > 15:
		return fmt.Sprintf("%.15q...", t.value)
	}

	return fmt.Sprintf("%q", t.value)
}
