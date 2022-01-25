// Package lexer implements spok's semantic lexing
//
// Spok uses a concurrent, state-function based lexer similar to that described by Rob Pike
// in his talk "Lexical Scanning in Go" at GoogleFOSSSydney which was based on the lexer
// used in the go package text/template
package lexer

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// stateFunc represents the state of the lexer as a function that returns the next state
type stateFunc func(*lexer) stateFunc

type lexer struct {
	tokens chan token // Channel of lexed tokens, received by the parser
	input  string     // String being lexed
	pos    int        // Current position in the input
	start  int        // Start position of the current token
	width  int        // Width of the last rune read from input
	line   int        // Line number of the current token
}

// skipWhitespace absorbs whitespace until the lexer hits anything meaningful
func (l *lexer) skipWhitespace() {
	for {
		r := l.next()
		if !unicode.IsSpace(r) {
			l.backup()
			break
		}

		// We've reached the end of the file
		if r == rune(tokenEOF) {
			l.emit(tokenEOF)
			break
		}
	}
}

//

// next consumes and returns the next rune in the input
func (l *lexer) next() rune {
	if l.pos >= len(l.input) {
		l.width = 0
		return rune(tokenEOF)
	}
	rune, width := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = width
	l.pos += l.width
	if rune == '\n' {
		l.line++
	}
	return rune
}

// peek returns but does not consume the next rune in the input
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next
func (l *lexer) backup() {
	l.pos -= l.width
	// Correct newline count.
	if l.width == 1 && l.input[l.pos] == '\n' {
		l.line--
	}
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (l *lexer) errorf(format string, args ...interface{}) stateFunc {
	l.tokens <- token{
		typ:   tokenError,
		pos:   l.start,
		value: fmt.Sprintf(format, args...),
		line:  l.line,
	}
	return nil
}

// emit passes a token to the parser via the tokens channel
func (l *lexer) emit(t tokenType) {
	l.tokens <- token{
		typ:   t,
		pos:   l.start,
		value: l.input[l.start:l.pos],
		line:  l.line,
	}
	l.start = l.pos
}

// run runs the state machine for the lexer.
func (l *lexer) run() {
	for state := lexStart; state != nil; {
		state = state(l)
	}
	close(l.tokens)
}

// nextToken pulls the next token from the input, designed to be called
// from the parser not the lexing goroutine
func (l *lexer) nextToken() token {
	return <-l.tokens
}

// lex creates a new scanner for the input string.
func lex(input string) *lexer {
	l := &lexer{
		input:  input,
		tokens: make(chan token),
		line:   1,
	}
	go l.run()
	return l
}

// startState consumes whitespace until it hits anything that could be meaningful
// for spok the only things it can hit are comments, global variables and task definitions
func lexStart(l *lexer) stateFunc {
	l.skipWhitespace()

	switch {
	case strings.HasPrefix(l.input[l.pos:], tokenHash.String()):
		return lexHash
	case strings.HasPrefix(l.input[l.pos:], tokenTask.String()):
		return lexTask
	default:
		// Must be a global variable
		return lexIdent
	}

}

// lexHash lexes a spok comment statement, the hash is already known to be present
func lexHash(l *lexer) stateFunc {
	l.pos += len(tokenHash.String())
	l.emit(tokenHash)
	return lexComment
}

// lexTask lexes a spok task definition, the task keyword is already known to be present
func lexTask(l *lexer) stateFunc {
	// TODO: Implement
	return nil
}

// lexIdent lexes a spok identifier
func lexIdent(l *lexer) stateFunc {
	// TODO: Implement
	return nil
}

// lexComment lexes a spok comment, the hash marker is already known to be present
func lexComment(l *lexer) stateFunc {
	l.pos += len(tokenHash.String())

	for {
		r := l.next()
		fmt.Printf("Current rune: %s\n", string(r))
		if r == '\n' {
			fmt.Println("Found newline")
			l.backup()
			break
		}
		// We've reached the end of the file unexpectedly
		if r == rune(tokenEOF) {
			l.emit(tokenEOF)
			return l.errorf("EOF")
		}
	}

	l.emit(tokenComment)
	return lexStart
}
