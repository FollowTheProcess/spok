package parser

import (
	"testing"

	"github.com/FollowTheProcess/spok/ast"
	"github.com/FollowTheProcess/spok/token"
)

// testLexer is an object that implements the lexer.Tokeniser interface
// so we can generate a stream of tokens without textual input
// separating the concerns of the lexer and the parser, the latter
// should not have to care where the token stream comes from, it just needs
// to know how to convert them to ast nodes. This also means that if we break
// the actual lexer during development, the parser tests won't also break.
type testLexer struct {
	stream []token.Token
}

func (l *testLexer) NextToken() token.Token {
	if len(l.stream) == 0 {
		// We don't have to manually add EOF now
		l.stream = append(l.stream, newToken(token.EOF, ""))
	}

	// Grab the first in the stream, "consume" it from the stream
	// and return it
	tok := l.stream[0]
	l.stream = l.stream[1:]
	return tok
}

func newToken(typ token.Type, value string) token.Token {
	return token.Token{
		Value: value,
		Type:  typ,
	}
}

var (
	tHash = newToken(token.HASH, "#")
)

func TestEOF(t *testing.T) {
	p := &Parser{
		lexer:  &testLexer{},
		buffer: [3]token.Token{},
	}

	tree, err := p.Parse()
	if err != nil {
		t.Fatalf("Parser returned an error: %v", err)
	}

	if tree == nil {
		t.Fatalf("Parser returned a nil AST")
	}

	if !tree.IsEmpty() {
		t.Fatalf("Wrong number of ast nodes, got %d, wanted %d", len(tree.Nodes), 0)
	}
}

func TestParseComment(t *testing.T) {
	commentStream := []token.Token{tHash, newToken(token.COMMENT, " A comment")}
	p := &Parser{
		lexer:     &testLexer{stream: commentStream},
		text:      "",
		buffer:    [3]token.Token{},
		peekCount: 0,
	}

	tree, err := p.Parse()
	if err != nil {
		t.Fatalf("Parser returned an error token: %v", err)
	}

	if tree == nil {
		t.Fatalf("Parser returned a nil AST")
	}

	if len(tree.Nodes) != 1 {
		t.Fatalf("Wrong number of ast nodes, got %d, wanted %d", len(tree.Nodes), 1)
	}

	node := tree.Nodes[0]
	comment, ok := node.(ast.CommentNode)
	if !ok {
		t.Fatalf("Node was not a comment node, got %T", node)
	}

	if comment.Text != " A comment" {
		t.Errorf("Wrong comment text: got %s, wanted %s", comment.Text, " A comment")
	}

}
