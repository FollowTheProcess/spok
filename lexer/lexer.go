// Package lexer implements spok's semantic lexer.
//
// Spok uses a concurrent, state-function based lexer similar to that described by Rob Pike
// in his talk "Lexical Scanning in Go", based on the implementation of template/text in the go std lib.
//
// The lexer proceeds one utf-8 rune at a time until a particular lexical token is recognised,
// the token is then "emitted" over a channel where it may be consumed by a client e.g. the parser.
// The state of the lexer is maintained between token emits unlike a more conventional switch-based lexer
// that must determine it's current state from scratch in every loop.
//
// This lexer uses "lexFunctions" to pass the state from one loop to an another. For example, if we're currently
// lexing a global variable ident, the next token must be a ':=' so we can go straight there without traversing
// the entire lexical state space first to determine "are we in a global variable definition?".
//
// The lexer 'run' method consumes these "lexFunctions" which return states in a continual loop until nil is returned
// marking the fact that "there is nothing more to lex" at which point the lexer closes the tokens channel, which
// will be picked up by the parser as a signal that the input stream has ended.
package lexer

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/FollowTheProcess/spok/token"
)

const (
	eof    = -1           // Sigil for marking an EOF as a rune
	digits = "0123456789" // Valid numeric digits
)

// lexFn represents the state of the scanner as a function that returns the next state.
type lexFn func(*lexer) lexFn

// lexer is spok's semantic lexer.
type lexer struct {
	tokens chan token.Token // Channel of lexed tokens, received by the parser
	input  string           // The string being scanned
	start  int              // Start position of the current token
	pos    int              // Current position in the input
	line   int              // Current line in the input
	width  int              // Width of the last rune read from input
}

// rest returns the string from the current lexer position to the end of the input.
func (l *lexer) rest() string {
	if l.atEOF() {
		return ""
	}
	return l.input[l.pos:]
}

// all returns the string from the lexer start position to the end of the input.
func (l *lexer) all() string {
	if l.start >= len(l.input) || l.pos > len(l.input) {
		return ""
	}
	return l.input[l.start:l.pos]
}

// current returns the rune the lexer is currently sat on.
func (l *lexer) current() rune {
	if l.atEOF() {
		return eof
	}
	return rune(l.input[l.pos])
}

// atEOL returns whether or not the lexer is currently at the end of a line.
func (l *lexer) atEOL() bool {
	return l.peek() == '\n' || strings.HasPrefix(l.rest(), "\r\n")
}

// atEOF returns whether or not the lexer is currently at the end of a file.
func (l *lexer) atEOF() bool {
	return l.pos >= len(l.input)
}

// skipWhitespace consumes any utf-8 whitespace until something meaningful is hit.
func (l *lexer) skipWhitespace() {
	for {
		r := l.next()
		if !unicode.IsSpace(r) {
			l.backup()  // Go back to the last non-space
			l.discard() // Bring the start position of the lexer up to current
			break
		}

		if r == eof {
			l.emit(token.EOF)
			break
		}
	}
}

// next returns, and consumes, the next rune in the input.
func (l *lexer) next() rune {
	if l.pos > len(l.input) {
		l.width = 0
		return eof
	}
	rune, width := utf8.DecodeRuneInString(l.rest())
	l.width = width
	l.pos += l.width
	if rune == '\n' {
		l.line++
	}
	return rune
}

// peek returns, but does not consume, the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
	if l.width == 1 && l.current() == '\n' {
		l.line--
	}
}

// absorb advances the lexer position over to the end of the given token.
func (l *lexer) absorb(t token.Type) {
	l.pos += len(t.String())
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) {
	for strings.ContainsRune(valid, l.next()) {
	}
	l.backup()
}

// emit passes an item back to the parser via the tokens channel.
func (l *lexer) emit(t token.Type) {
	l.tokens <- token.Token{
		Value: l.all(),
		Type:  t,
		Pos:   l.start,
		Line:  l.line,
	}
	l.start = l.pos
}

// discard brings the lexer's start position up to it's current position,
// discaring everything in between in the process but maintaining the line count.
func (l *lexer) discard() {
	l.line += strings.Count(l.all(), "\n")
	l.start = l.pos
}

// errorf returns an error token and terminates the scan by passing back
// a nil pointer that will be the next state, terminating l.nextToken.
func (l *lexer) errorf(format string, args ...interface{}) lexFn {
	l.tokens <- token.Token{
		Value: fmt.Sprintf(format, args...),
		Type:  token.ERROR,
		Pos:   l.start,
		Line:  l.line,
	}
	return nil
}

// nextToken returns the next token from the input,
// generally called by the parser not the lexing goroutine.
func (l *lexer) nextToken() token.Token {
	return <-l.tokens
}

// lex creates a new lexer for the input string and sets it off
// in a goroutine.
func lex(input string) *lexer {
	l := &lexer{
		tokens: make(chan token.Token),
		input:  input,
		start:  0,
		pos:    0,
		line:   1,
		width:  0,
	}
	go l.run()
	return l
}

// run starts the state machine for the lexer.
func (l *lexer) run() {
	for state := lexStart; state != nil; {
		state = state(l)
	}
	close(l.tokens)
}

// lexStart is the initial state of the lexer.
//
// The only things spok can encounter at the top level are:
// Whitespace: ignored
// Comments: preceded with a '#'
// Global variables
// Task definitions
// EOF
// Anything else is an error.
func lexStart(l *lexer) lexFn {

	l.skipWhitespace()

	switch {
	case strings.HasPrefix(l.rest(), token.HASH.String()):
		return lexHash
	case strings.HasPrefix(l.rest(), token.TASK.String()):
		return lexTaskKeyword
	case unicode.IsLetter(l.peek()):
		return lexIdent
	case l.atEOF():
		l.emit(token.EOF)
		return nil
	default:
		return unexpectedToken
	}
}

// lexHash scans a comment marker '#'.
func lexHash(l *lexer) lexFn {
	l.absorb(token.HASH)
	l.emit(token.HASH)
	return lexComment
}

// lexComment scans a comment text, the '#' has already been encountered.
func lexComment(l *lexer) lexFn {
	for {
		if l.atEOL() || l.atEOF() {
			l.emit(token.COMMENT)
			return lexStart
		}
		l.next()
	}
}

// lexTaskKeyword scans a task definition keyword.
func lexTaskKeyword(l *lexer) lexFn {
	l.absorb(token.TASK)
	l.emit(token.TASK)
	l.skipWhitespace()
	return lexIdent
}

// lexLeftParen scans an opening parenthesis.
func lexLeftParen(l *lexer) lexFn {
	l.absorb(token.LPAREN)
	l.emit(token.LPAREN)
	l.skipWhitespace()
	return lexArgs
}

// lexRightParen scans a closing parenthesis.
func lexRightParen(l *lexer) lexFn {
	l.absorb(token.RPAREN)
	l.emit(token.RPAREN)
	l.skipWhitespace()
	return lexLeftBrace
}

// lexLeftBrace scans an opening curly brace.
func lexLeftBrace(l *lexer) lexFn {
	l.absorb(token.LBRACE)
	l.emit(token.LBRACE)
	l.skipWhitespace()
	return lexTaskBody
}

// lexRightBrace scans a closing curly brace.
func lexRightBrace(l *lexer) lexFn {
	l.absorb(token.RBRACE)
	l.emit(token.RBRACE)
	return lexStart
}

// lexTaskBody scans the body of a task declaration.
func lexTaskBody(l *lexer) lexFn {
	if l.atEOF() {
		return l.errorf("SyntaxError: Unterminated task body (Line %d, Position %d)", l.line, l.pos)
	}
	l.skipWhitespace()

	switch r := l.next(); {
	case r == '}':
		l.backup()
		return lexRightBrace
	case unicode.IsLetter(r):
		// Assumes command starts with a letter, pretty safe for 99.9% of commands
		return lexTaskCommands
	default:
		return unexpectedToken
	}
}

// lexTaskCommands scans line(s) of commands in a task body.
func lexTaskCommands(l *lexer) lexFn {
	// A command can end in a newline or not similar to a line in a go function
	// e.g. this is valid -> task test() { go test ./... }
	// as well as the "normal" go function style spread over a few lines:
	// task test() {
	//	go test ./...
	// }

	for {
		switch r := l.next(); {
		case r == '\n':
			// If there's a newline, might be more commands on the next line
			l.backup()
			l.emit(token.COMMAND)
			l.skipWhitespace()
		case r == '}':
			l.backup()
			// The command may end in a space which we should clean up
			if strings.HasSuffix(l.all(), " ") {
				l.pos--
			}
			if len(l.all()) != 0 {
				// If we actually have a command and not just an empty token
				l.emit(token.COMMAND)
			}
			l.skipWhitespace()
			return lexRightBrace
		case isASCII(r):
			// Potential command text, absorb.
		default:
			return unexpectedToken
		}
	}
}

// lexIdent scans an identifier e.g. global variable or name of task.
func lexIdent(l *lexer) lexFn {
	// Read until we get an invalid ident rune
	for {
		r := l.next()
		if !isValidIdent(r) {
			l.backup()
			break
		}
	}
	l.emit(token.IDENT)
	l.skipWhitespace()

	switch {
	case l.peek() == '(':
		// We have arguments i.e. a task
		return lexLeftParen
	case strings.HasPrefix(l.rest(), token.DECLARE.String()):
		// We have a global variable declaration
		return lexDeclare
	case l.atEOL(), l.atEOF():
		// We've just lexed an ident on the RHS of a declaration
		return lexStart
	case l.peek() == ')':
		// It's an ident used in a task argument
		return lexRightParen
	case l.peek() == ',':
		// It's an ident in a list of task arguments
		return lexArgs
	default:
		// Whatever it is shouldn't be here
		return unexpectedToken
	}
}

// lexArgs scans an argument declaration i.e. task dependencies or builtin function args.
func lexArgs(l *lexer) lexFn {
	// Arguments can only be strings (file dependencies) or names of other tasks
	l.skipWhitespace()

	switch r := l.next(); {
	case r == ')':
		// No task dependency
		l.backup()
		return lexRightParen
	case r == '"':
		return lexString
	case isValidIdent(r):
		return lexIdent
	case r == ',':
		// We have a list of arguments, lex the next one
		return lexArgs
	default:
		return l.errorf("SyntaxError: Invalid character used in task dependency [%s] (Line %d, Position %d). Only strings and declared variables may be used.", string(l.current()), l.line, l.pos)
	}
}

// lexDeclare scans a declaration operation in a global variable.
func lexDeclare(l *lexer) lexFn {
	l.skipWhitespace()
	l.absorb(token.DECLARE)
	l.emit(token.DECLARE)
	l.skipWhitespace()

	switch r := l.next(); {
	case r == '"':
		// We have a quoted string e.g. "hello"
		return lexString
	case isValidIdent(r):
		// We have something unquoted, i.e. another ident
		return lexIdent
	case unicode.IsDigit(r):
		// We have a number e.g. 27
		return lexInteger
	default:
		// Anything else is disallowed
		return unexpectedToken
	}

}

// lexString scans a quoted string, the opening quote is already known to exist.
func lexString(l *lexer) lexFn {
	for {
		r := l.next()
		if r == '"' {
			break
		}

		if l.atEOF() || l.atEOL() {
			return l.errorf("SyntaxError: Unterminated string literal (Line %d, Position %d)", l.line, l.pos)
		}
	}

	l.emit(token.STRING)
	if l.atEOF() || l.atEOL() {
		// If this is the end, it must have been a global variable assignment
		return lexStart
	}
	// Else we must be handling a task argument
	return lexArgs
}

// lexInteger scans a decimal integer.
func lexInteger(l *lexer) lexFn {
	l.acceptRun(digits)

	// Next thing cannot be anything other than EOL or EOF
	// if so, we have a bad integer e.g. 2756g
	if !l.atEOF() && !l.atEOL() {
		return l.errorf("SyntaxError: Invalid integer literal (Line %d, Position %d)", l.line, l.pos)
	}

	l.emit(token.INTEGER)
	return lexStart
}

// isValidIdent reports whether a rune is valid in an identifier.
func isValidIdent(r rune) bool {
	return unicode.IsLetter(r) || r == '_'
}

// isASCII reports whether or not the rune is a valid ASCII character.
func isASCII(r rune) bool {
	if r > unicode.MaxASCII {
		return false
	}
	return true
}

// unexpectedToken emits an error token with details about the offending input from the lexer.
func unexpectedToken(l *lexer) lexFn {
	var message string
	char := l.current()

	switch {
	case unicode.IsGraphic(char):
		message = fmt.Sprintf("SyntaxError: Unexpected token '%s' (Line %d, Position %d)", string(char), l.line, l.pos)
	default:
		message = fmt.Sprintf("SyntaxError: Unexpected token '%U' (Line %d, Position %d)", char, l.line, l.pos)
	}

	return l.errorf(message)
}
