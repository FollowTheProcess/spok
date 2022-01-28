// Package token declares a number of constants that represent lexical tokens in spok
// as well as basic operations on those tokens e.g. printing
package token

import "fmt"

// Type is the set of lexical tokens in spok.
type Type int

//go:generate stringer -type=Type -linecomment -output=token_string.go
const (
	ERROR    Type = iota // ERROR
	EOF                  // EOF
	COMMENT              // COMMENT
	HASH                 // #
	LPAREN               // (
	RPAREN               // )
	LBRACE               // {
	RBRACE               // }
	LBRACKET             // [
	RBRACKET             // ]
	QUOTE                // "
	COMMA                // ,
	TASK                 // task
	STRING               // STRING
	COMMAND              // COMMAND
	INTEGER              // INTEGER
	OUTPUT               // ->
	IDENT                // IDENT
	DECLARE              // :=
)

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
	case len(t.Value) > 15:
		return fmt.Sprintf("%.15q...", t.Value)
	}
	return fmt.Sprintf("%q", t.Value)
}

// Is returns whether or not the current token is of a certain type.
func (t Token) Is(typ Type) bool {
	return t.Type == typ
}
