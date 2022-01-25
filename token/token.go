// Package token declares a number of constants that represent lexical tokens in spok
// as well as basic operations on those tokens e.g. printing
package token

// Token is the set of lexical tokens in spok
type Token int

//go:generate stringer -type=Token -linecomment
const (
	ERROR    Token = iota // ERROR
	EOF                   // EOF
	COMMENT               // COMMENT
	LPAREN                // (
	RPAREN                // )
	LBRACE                // {
	RBRACE                // }
	LBRACKET              // [
	RBRACKET              // ]
	QUOTE                 // "
	TASK                  // TASK
	STRING                // STRING
	OUTPUT                // ->
	IDENT                 // IDENT
)
