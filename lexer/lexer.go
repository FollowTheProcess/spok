// Package lexer implements spok's semantic lexer.
//
// Spok uses a concurrent, state-function based lexer similar to that described by Rob Pike
// in his talk "lexical analysis in Go", based on the implementation of template/text in the go std lib
package lexer

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/FollowTheProcess/spok/token"
)

const eof = -1

// lexFn represents the state of the scanner as a function that returns the next state
type lexFn func(*lexer) lexFn

// lexer is spok's semantic lexer
type lexer struct {
	tokens chan token.Token // Channel of lexed tokens, received by the parser
	input  string           // The string being scanned
	start  int              // Start position of the current token
	pos    int              // Current position in the input
	line   int              // Current line in the input
	width  int              // Width of the last rune read from input
}

// rest returns the string from the current lexer position to the end of the input
func (l *lexer) rest() string {
	return l.input[l.pos:]
}

// all returns the string from the lexer start position to the end of the input
func (l *lexer) all() string {
	return l.input[l.start:l.pos]
}

// current returns the rune the lexer is currently sat on i.e. l.input[l.pos]
func (l *lexer) current() rune {
	return rune(l.input[l.pos])
}

// atEOL returns whether or not the lexer is currently at the end of a line
func (l *lexer) atEOL() bool {
	return l.peek() == '\n' || strings.HasPrefix(l.rest(), "\r\n")
}

// atEOF returns whether or not the lexer is currently at the end of a file
func (l *lexer) atEOF() bool {
	return l.pos >= len(l.input)
}

// skipWhitespace consumes any utf-8 whitespace until something meaningful is hit
func (l *lexer) skipWhitespace() {
	for {
		r := l.next()
		if !unicode.IsSpace(r) {
			l.backup()
			break
		}

		if r == eof {
			l.emit(token.EOF)
			break
		}
	}
}

// next returns (and consumes) the next rune in the input
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

// peek returns (but does not consume) the next rune in the input
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next
func (l *lexer) backup() {
	l.pos -= l.width
	if l.width == 1 && l.current() == '\n' {
		l.line--
	}
}

// emit passes an item back to the parser via the tokens channel
func (l *lexer) emit(t token.Type) {
	l.tokens <- token.Token{
		Value: l.all(),
		Type:  t,
		Pos:   l.start,
		Line:  l.line,
	}
	l.start = l.pos
}

// ignore skips over the input before the current lexer position
// effectively just moves start up to current pos
func (l *lexer) ignore() {
	l.line += strings.Count(l.all(), "\n")
	l.start = l.pos
}

// errorf returns an error token and terminates the scan by passing back
// a nil pointer that will be the next state, terminating l.nextToken
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
// generally called by the parser not the lexing goroutine
func (l *lexer) nextToken() token.Token {
	return <-l.tokens
}

// lex creates a new lexer for the input string and sets it off
// in a goroutine
func lex(input string) *lexer {
	l := &lexer{
		tokens: make(chan token.Token),
		input:  input,
		start:  0,
		pos:    0,
		line:   0,
		width:  0,
	}
	go l.run()
	return l
}

// run starts the state machine for the lexer
func (l *lexer) run() {
	for state := lexStart; state != nil; {
		state = state(l)
	}
	close(l.tokens)
}

// lexStart is the initial state of the lexer
func lexStart(l *lexer) lexFn {
	l.skipWhitespace()
	// The only things spok can encounter at the top level are:
	// - Comments, preceded with a '#'
	// - Global variables
	// - Task definitions
	// - EOF
	// Anything else is an error
	switch {
	case strings.HasPrefix(l.rest(), token.HASH.String()):
		return lexHash
	case strings.HasPrefix(l.rest(), token.TASK.String()):
		return lexTask
	case unicode.IsLetter(l.peek()):
		return lexIdent
	case l.atEOF():
		// atEOF means we know there's nothing left (maybe a \n)
		l.ignore()
		l.emit(token.EOF)
		return nil
	default:
		l.errorf("Unexpected token")
		return nil
	}
}

// lexHash scans a comment marker '#'
func lexHash(l *lexer) lexFn {
	l.pos += len(token.HASH.String())
	l.emit(token.HASH)
	return lexComment
}

// lexComment scans a comment text, the '#' has already been encountered
func lexComment(l *lexer) lexFn {
	for {
		if l.atEOL() {
			l.emit(token.COMMENT)
			l.pos++
			return lexStart
		}

		l.next()

		if l.atEOF() {
			return l.errorf("Unexpected EOF")
		}
	}
}

// lexTask scans a task definition keyword
func lexTask(l *lexer) lexFn {
	// TODO: Implement
	return nil
}

// lexIdent scans an identifier e.g. global variable or name of task
func lexIdent(l *lexer) lexFn {
	// TODO: Implement
	return nil
}
