package lexer

import "testing"

type lexTest struct {
	name   string
	input  string
	tokens []token
}

func makeToken(typ tokenType, val string) token {
	return token{typ: typ, value: val}
}

var lexTests = []lexTest{
	{
		name:   "comment",
		input:  "# A comment\n",
		tokens: []token{makeToken(tokenHash, "#"), makeToken(tokenComment, "A comment")},
	},
}

func equal(t1, t2 []token, checkPos bool) bool {
	if len(t1) != len(t2) {
		return false
	}
	for k := range t1 {
		if t1[k].typ != t2[k].typ {
			return false
		}
		if t1[k].value != t2[k].value {
			return false
		}
		if checkPos && t1[k].pos != t2[k].pos {
			return false
		}
		if checkPos && t1[k].line != t2[k].line {
			return false
		}
	}
	return true
}

// collect gathers the emitted items into a slice.
func collect(t *lexTest) (tokens []token) {
	l := lex(t.input)
	for {
		tok := l.nextToken()
		tokens = append(tokens, tok)
		if tok.typ == tokenEOF || tok.typ == tokenError {
			break
		}
	}
	return
}

func TestLex(t *testing.T) {
	for _, test := range lexTests {
		items := collect(&test)
		if !equal(items, test.tokens, false) {
			t.Errorf("%s: got\n\t%+v\nexpected\n\t%+v", test.name, items, test.tokens)
		}
	}
}
