// Package lexer implements spok's semantic lexer.
//
// Spok uses a concurrent, state-function based lexer similar to that described by Rob Pike
// in his talk "lexical analysis in Go", based on the implementation of template/text in the go std lib
package lexer

import (
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

// skipWhitespace consumes any utf-8 whitespace until something meaningful is hit
func (l *lexer) skipWhitespace() {
	for {
		r := l.next()

		if !unicode.IsSpace(r) {
			l.backup()
			break
		}

		if r == rune(token.EOF) {
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
