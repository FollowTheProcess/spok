package lexer

type tokenType int

//go:generate stringer -type=tokenType -linecomment
const (
	tokenError        tokenType = iota // Error
	tokenEOF                           // EOF
	tokenIDENT                         // IDENT
	tokenInt                           // Integer
	tokenString                        // String
	tokenQuote                         // "
	tokenColon                         // :
	tokenEquals                        // =
	tokenTask                          // task
	tokenOpenParen                     // (
	tokenCloseParen                    // )
	tokenOpenBrace                     // {
	tokenCloseBrace                    // }
	tokenOpenBracket                   // [
	tokenCloseBracket                  // ]
	tokenArrowStem                     // -
	tokenArrowHead                     // >

)

//
type token struct {
	value string    // Value, e.g. "23"
	typ   tokenType // Type, e.g. tokenString
}
