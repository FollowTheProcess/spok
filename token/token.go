// Package token declares a number of constants that represent lexical tokens in spok
// as well as basic operations on those tokens e.g. printing
package token

import "fmt"

// Type is the set of lexical tokens in spok.
type Type int

// Note: EOF is the zero value such that when the parser reads from a closed channel
// the read value will be token.EOF.
//
//go:generate stringer -type=Type -linecomment -output=token_string.go
const (
	EOF     Type = iota // EOF
	ERROR               // ERROR
	COMMENT             // COMMENT
	HASH                // #
	LPAREN              // (
	RPAREN              // )
	LBRACE              // {
	RBRACE              // }
	QUOTE               // "
	COMMA               // ,
	TASK                // task
	STRING              // STRING
	COMMAND             // COMMAND
	OUTPUT              // ->
	IDENT               // IDENT
	DECLARE             // :=
	LINTERP             // {{
	RINTERP             // }}
)

const displayLength = 15

// Token represents a spok lexical token.
type Token struct {
	Value string // Value, e.g. "("
	Type  Type   // Type, e.g. LPAREN
	Pos   int    // Starting position of this token in the input string
	Line  int    // Line number at the start of this token
}

// String satisfies the stringer interface and allows us to pretty print the tokens.
func (t Token) String() string {
	switch {
	case t.Type == EOF:
		return "EOF"
	case t.Type == ERROR:
		return t.Value
	case len(t.Value) > displayLength:
		return fmt.Sprintf("%.15q...", t.Value)
	}
	return fmt.Sprintf("%q", t.Value)
}

// Is returns whether or not the current token is of a certain type.
func (t Token) Is(typ Type) bool {
	return t.Type == typ
}
