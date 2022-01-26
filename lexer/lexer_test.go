package lexer

import (
	"testing"

	"github.com/FollowTheProcess/spok/token"
)

type lexTest struct {
	name   string
	input  string
	tokens []token.Token
}

func newToken(typ token.Type, value string) token.Token {
	return token.Token{
		Value: value,
		Type:  typ,
	}
}

var (
	tEOF  = newToken(token.EOF, "")
	tHash = newToken(token.HASH, "#")
)

var lexTests = []lexTest{
	{
		name:   "empty",
		input:  "",
		tokens: []token.Token{tEOF},
	},
	{
		name:   "hash",
		input:  "#\n",
		tokens: []token.Token{tHash, newToken(token.COMMENT, ""), tEOF},
	},
	{
		name:   "comment",
		input:  "# A comment\n",
		tokens: []token.Token{tHash, newToken(token.COMMENT, " A comment"), tEOF},
	},
	{
		name:   "whitespace",
		input:  "      \t\n\t\t\n\n\n   ",
		tokens: []token.Token{tEOF},
	},
}

// collect gathers the emitted tokens into a slice for comparison
func collect(t *lexTest) (tokens []token.Token) {
	l := lex(t.input)
	for {
		tok := l.nextToken()
		tokens = append(tokens, tok)
		if tok.Type == token.EOF || tok.Type == token.ERROR {
			break
		}
	}
	return
}

// equal compares to slices of tokens for equality
func equal(t1, t2 []token.Token) bool {
	if len(t1) != len(t2) {
		return false
	}
	for i := range t1 {
		if t1[i].Type != t2[i].Type {
			return false
		}
		if t1[i].Value != t2[i].Value {
			return false
		}
	}
	return true
}

func TestLexer(t *testing.T) {
	for _, test := range lexTests {
		t.Run(test.name, func(t *testing.T) {
			tokens := collect(&test)
			if !equal(tokens, test.tokens) {
				t.Errorf("%s: got\n\t%#v\nexpected\n\t%#v", test.name, tokens, test.tokens)
			}
		})
	}
}
