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
// marking the fact that either "there is nothing more to lex" or "we've hit an error" at which point the lexer closes
// the tokens channel, which will be picked up by the parser as a signal that the input stream has ended.
//
// In lexing/parsing, the error checking complexity is always kept somewhere. Spok has made the choice that the lexer
// should do much of the syntax error handling as it has the most direct access to the raw input as well as the positions,
// characters etc. The approach of stateful "lexFunctions" helps enable this as every lexing function "knows where it is"
// in the language, improving the quality of the error messages. The lexer handling most of the error complexity has helped
// to keep the parser very simple which I think is a good trade off and the test cases for the parser already far outweigh
// that of the lexer.
package lexer

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/FollowTheProcess/spok/token"
)

// Tokeniser represents anything capable of producing a token.Token
// when asked to by it's NextToken method, this includes our actual Lexer
// defined below but can be readily stubbed out for testing e.g. the parser.
type Tokeniser interface {
	// NextToken yields a single token from the input stream.
	NextToken() token.Token
}

// lexFn represents the state of the scanner as a function that returns the next state.
type lexFn func(*Lexer) lexFn

// Lexer is spok's semantic Lexer.
type Lexer struct {
	tokens    chan token.Token // Channel of lexed tokens, received by the parser
	input     string           // The string being scanned
	start     int              // Start position of the current token
	pos       int              // Current position in the input
	line      int              // Current line in the input
	startLine int              // The line on which the current token started
	width     int              // Width of the last rune read from input
}

// rest returns the string from the current lexer position to the end of the input.
func (l *Lexer) rest() string {
	if l.atEOF() {
		return ""
	}
	return l.input[l.pos:]
}

// all returns the string from the lexer start position to it's current position.
func (l *Lexer) all() string {
	if l.start >= len(l.input) || l.pos > len(l.input) {
		return ""
	}
	return l.input[l.start:l.pos]
}

// current returns the rune the lexer is currently sat on.
func (l *Lexer) current() rune {
	return rune(l.input[l.pos])
}

// atEOL returns whether or not the lexer is currently at the end of a line.
func (l *Lexer) atEOL() bool {
	return l.peek() == '\n' || strings.HasPrefix(l.rest(), "\r\n")
}

// atEOF returns whether or not the lexer is currently at the end of a file.
func (l *Lexer) atEOF() bool {
	return l.pos >= len(l.input)
}

// skipWhitespace consumes any utf-8 whitespace until something meaningful is hit.
func (l *Lexer) skipWhitespace() {
	for {
		r := l.next()
		if !unicode.IsSpace(r) {
			l.backup()  // Go back to the last non-space
			l.discard() // Bring the start position of the lexer up to current
			break
		}
	}
}

// next returns, and consumes, the next rune in the input.
func (l *Lexer) next() rune {
	rune, width := utf8.DecodeRuneInString(l.rest())
	l.width = width
	l.pos += l.width
	if rune == '\n' {
		l.line++
	}
	return rune
}

// peek returns, but does not consume, the next rune in the input.
func (l *Lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *Lexer) backup() {
	l.pos -= l.width
	if l.width == 1 && l.current() == '\n' {
		l.line--
	}
}

// absorb advances the lexer position over the given token.
func (l *Lexer) absorb(t token.Type) {
	l.pos += len(t.String())
}

// emit passes an item back to the parser via the tokens channel.
func (l *Lexer) emit(t token.Type) {
	l.tokens <- token.Token{
		Value: l.all(),
		Type:  t,
		Pos:   l.start,
		Line:  l.startLine,
	}
	l.start = l.pos
	l.startLine = l.line
}

// discard brings the lexer's start position up to it's current position,
// discaring everything in between in the process but maintaining the line count.
func (l *Lexer) discard() {
	l.start = l.pos
	l.startLine = l.line
}

// error emits an error token and terminates the scan by passing back
// a nil pointer that will be the next state, terminating l.run().
func (l *Lexer) error(err error) lexFn {
	l.tokens <- token.Token{
		Value: err.Error(),
		Type:  token.ERROR,
		Pos:   l.start,
		Line:  l.startLine,
	}
	return nil
}

// getLine is called when erroring and gets the entire current line of context from the
// input to show with the error.
func (l *Lexer) getLine() string {
	var lines []string
	for _, line := range strings.Split(l.input, "\n") {
		lines = append(lines, strings.TrimSpace(line))
	}
	return lines[l.line-1]
}

// NextToken returns the next token from the input,
// generally called by the parser not the lexing goroutine.
func (l *Lexer) NextToken() token.Token {
	return <-l.tokens
}

// New creates a new lexer for the input string and sets it off in a goroutine.
func New(input string) *Lexer {
	l := &Lexer{
		tokens:    make(chan token.Token),
		input:     input,
		start:     0,
		pos:       0,
		line:      1,
		startLine: 1,
		width:     0,
	}
	go l.run()
	return l
}

// run starts the state machine for the lexer.
// when the next state is nil (EOF or Error), the loop is broken and the tokens channel is closed.
// There's a nice little go trick here:
// Closing the channel means that, if it reads from the channel, the parser will receive the zero value
// without blocking, and because our channel is a channel of token.Token, which has an underlying
// type of int, the zero value is 0. And in the enumerated constants defined in token.go, 0 is
// mapped to EOF. This means our parser will always have an EOF as the last token.
func (l *Lexer) run() {
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
func lexStart(l *Lexer) lexFn {
	l.skipWhitespace()

	switch {
	case strings.HasPrefix(l.rest(), token.HASH.String()):
		return lexHash
	case strings.HasPrefix(l.rest(), token.TASK.String()):
		return lexTaskKeyword
	case isValidIdent(l.peek()):
		return lexIdent
	case l.atEOF():
		l.emit(token.EOF)
		return nil
	default:
		return unexpectedToken
	}
}

// lexHash scans a comment marker '#'.
func lexHash(l *Lexer) lexFn {
	l.absorb(token.HASH)
	l.emit(token.HASH)
	return lexComment
}

// lexComment scans a comment text, the '#' has already been encountered.
func lexComment(l *Lexer) lexFn {
	for {
		if l.atEOL() || l.atEOF() {
			l.emit(token.COMMENT)
			return lexStart
		}
		l.next()
	}
}

// lexTaskKeyword scans a task definition keyword.
func lexTaskKeyword(l *Lexer) lexFn {
	l.absorb(token.TASK)
	l.emit(token.TASK)
	l.skipWhitespace()
	return lexTaskName
}

// lexLeftParen scans an opening parenthesis.
func lexLeftParen(l *Lexer) lexFn {
	l.absorb(token.LPAREN)
	l.emit(token.LPAREN)
	l.skipWhitespace()
	return lexArgs
}

// lexRightParen scans a closing parenthesis.
func lexRightParen(l *Lexer) lexFn {
	l.absorb(token.RPAREN)
	l.emit(token.RPAREN)
	l.skipWhitespace()

	switch r := l.peek(); {
	case r == '{':
		// Next thing up is a task body
		return lexLeftBrace
	case strings.HasPrefix(l.rest(), token.OUTPUT.String()):
		// Task output declaration
		return lexOutputOperator
	case l.atEOL(), l.atEOF(), isValidIdent(r):
		// Just lexed a global variable function call and we're either
		// at EOL, EOF, or there's another row of global variable declarations below
		return lexStart
	case r == '#':
		// If a global function call precedes a commented task
		return lexHash
	case r == '"', r == '(':
		// This is when someone forgets a '->' when declaring task outputs
		return l.error(syntaxError{
			message: fmt.Sprintf("Unexpected token '%s'. Task output missing the '->' operator?", string(r)),
			context: l.getLine(),
			line:    l.line,
			pos:     l.pos,
		})
	default:
		return unexpectedToken
	}
}

// lexOutputOperator scans a task output operator.
func lexOutputOperator(l *Lexer) lexFn {
	l.absorb(token.OUTPUT)
	l.emit(token.OUTPUT)
	l.skipWhitespace()

	switch r := l.next(); {
	case r == '"':
		// Single task output
		return lexString
	case r == '(':
		// List of task outputs, nice little hack here because the rules
		// are the same as task dependencies
		l.backup()
		return lexLeftParen
	case isValidIdent(r):
		return lexIdent
	case r == '{':
		// Error: declared task has an output but didn't specify it
		l.backup()
		return l.error(syntaxError{
			message: "Task declared dependency but none found",
			context: l.getLine(),
			line:    l.line,
			pos:     l.pos,
		})
	case unicode.IsPunct(r):
		// This is normally a filepath-like string (e.g. "file.go", or "./bin/main") with no opening quote
		return l.error(syntaxError{
			message: fmt.Sprintf("Unexpected punctuation in ident '%s'. String missing opening quote?", string(r)),
			context: l.getLine(),
			line:    l.line,
			pos:     l.pos,
		})
	default:
		l.backup()
		return unexpectedToken
	}
}

// lexLeftBrace scans an opening curly brace.
func lexLeftBrace(l *Lexer) lexFn {
	l.absorb(token.LBRACE)
	l.emit(token.LBRACE)
	l.skipWhitespace()
	return lexTaskBody
}

// lexRightBrace scans a closing curly brace.
func lexRightBrace(l *Lexer) lexFn {
	l.absorb(token.RBRACE)
	l.emit(token.RBRACE)
	return lexStart
}

// lexTaskBody scans the body of a task declaration.
func lexTaskBody(l *Lexer) lexFn {
	if l.atEOF() {
		return l.error(syntaxError{
			message: "Unterminated task body",
			context: l.getLine(),
			line:    l.line,
			pos:     l.pos,
		})
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
func lexTaskCommands(l *Lexer) lexFn {
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
		case strings.HasPrefix(l.rest(), token.LINTERP.String()):
			// We've hit an opening interpolation, ignore this here it just becomes
			// part of the command text
			l.absorb(token.LINTERP)
		case strings.HasPrefix(l.rest(), token.RINTERP.String()):
			// We've hit a closing interpolation, ignore this here it just becomes
			// part of the command text
			l.absorb(token.RINTERP)
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
		case l.atEOF(), r == '#':
			l.error(syntaxError{
				message: "Unterminated task body",
				context: l.getLine(),
				line:    l.line,
				pos:     l.pos,
			})
		case isASCII(r):
			// Potential command text, absorb.
		default:
			l.backup()
			return unexpectedToken
		}
	}
}

// lexTaskName scans an identifier in the specific context of the name of a task.
func lexTaskName(l *Lexer) lexFn {
	// Read until we hit an invalid ident rune
	for {
		r := l.next()
		if !isValidIdent(r) {
			l.backup()
			break
		}
	}
	l.emit(token.IDENT)
	l.skipWhitespace()

	// We know we're currently in a task name, so if the next thing
	// is not '(', it's a syntax error
	if l.peek() != '(' {
		return l.error(syntaxError{
			message: "Task missing parentheses, expected '('",
			context: l.getLine(),
			line:    l.line,
			pos:     l.pos,
		})
	}

	// We now know the next thing is a '(' and that's all we have to care about here
	return lexLeftParen
}

// lexIdent scans an identifier e.g. global variable or name of task.
func lexIdent(l *Lexer) lexFn {
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
		return lexComma
	case l.peek() == '{':
		// Just lexed an ident used in a task output
		return lexLeftBrace
	case unicode.IsPunct(l.peek()):
		// This is normally a filepath-like string like "file.go" but without the opening quote
		// or it's a missing comma in a series of args
		return l.error(syntaxError{
			message: "String literal missing opening quote or missing comma in variadic arguments",
			context: l.getLine(),
			line:    l.line,
			pos:     l.pos,
		})
	case isValidIdent(l.peek()):
		// Most likely a comment with a forgotten starting hash
		// i.e. # This is a comment becomes 'This' 'is' and the second ident
		// is what we pick up here
		return l.error(syntaxError{
			message: fmt.Sprintf("Unexpected token '%s'. Comment without a '#' or string literal missing opening quote?", string(l.peek())),
			context: l.getLine(),
			line:    l.line,
			pos:     l.pos,
		})
	default:
		// Whatever it is shouldn't be here
		return unexpectedToken
	}
}

// lexArgs scans an argument declaration i.e. task dependencies, builtin function args
// or lists of task outputs.
func lexArgs(l *Lexer) lexFn {
	// Arguments can only be strings (filenames or globs) or idents
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
		// We have a list of arguments, lex the comma
		l.backup()
		return lexComma
	case r == '{':
		// The string was a task output
		l.backup()
		return lexLeftBrace
	default:
		return l.error(syntaxError{
			message: "Invalid character used in task dependency/output",
			context: l.getLine(),
			line:    l.line,
			pos:     l.pos,
		})
	}
}

// lexComma scans a comma token.
func lexComma(l *Lexer) lexFn {
	l.absorb(token.COMMA)
	l.emit(token.COMMA)
	l.skipWhitespace()

	switch r := l.next(); {
	case r == '"':
		// Quoted string argument
		return lexString
	case isValidIdent(r):
		// Ident arg
		return lexIdent
	case r == ')':
		// Allow a trailing comma
		l.backup()
		return lexRightParen
	default:
		// Anything else is disallowed
		l.backup()
		return unexpectedToken
	}
}

// lexDeclare scans a declaration operation in a global variable.
func lexDeclare(l *Lexer) lexFn {
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
	default:
		// Anything else is disallowed
		l.backup()
		return unexpectedToken
	}
}

// lexString scans a quoted string, the opening quote is already known to exist,
// the emitted string token will always contain the quotes i.e. the token value
// in go-ish syntax will be `"hello"`, not simply "hello".
func lexString(l *Lexer) lexFn {
	for {
		r := l.next()
		if r == '"' {
			break
		}

		if l.atEOF() || l.atEOL() {
			l.backup()
			return l.error(syntaxError{
				message: fmt.Sprintf("String literal missing closing quote: %s", l.all()),
				context: l.getLine(),
				line:    l.line,
				pos:     l.pos,
			})
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

// isValidIdent reports whether a rune is valid in an identifier.
func isValidIdent(r rune) bool {
	return unicode.IsLetter(r) || r == '_'
}

// isASCII reports whether or not the rune is a valid ASCII character.
func isASCII(r rune) bool {
	return r <= unicode.MaxASCII
}

// unexpectedToken emits an error token with details about the offending input from the lexer.
func unexpectedToken(l *Lexer) lexFn {
	return l.error(syntaxError{
		message: fmt.Sprintf("Unexpected token '%s'", string(l.current())),
		context: l.getLine(),
		line:    l.line,
		pos:     l.pos,
	})
}
