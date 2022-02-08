package parser

import (
	"reflect"
	"testing"

	"github.com/FollowTheProcess/spok/ast"
	"github.com/FollowTheProcess/spok/token"
)

// testLexer is an object that implements the lexer.Tokeniser interface
// so we can generate a stream of tokens without textual input
// decoupling the lexer and the parser, the parser should not have to
// care where the token stream comes from, it just needs to know how to convert them to ast nodes.
// This also means that if we break the actual lexer during development, the parser tests won't also break.
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
	tHash    = newToken(token.HASH, "#")
	tDeclare = newToken(token.DECLARE, ":=")
	tLParen  = newToken(token.LPAREN, "(")
	tRParen  = newToken(token.RPAREN, ")")
	// tTask    = newToken(token.TASK, "task")
	// tLBrace  = newToken(token.LBRACE, "{")
	// tRBrace  = newToken(token.RBRACE, "}")
	// tOutput  = newToken(token.OUTPUT, "->").
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
		buffer:    [3]token.Token{},
		peekCount: 0,
	}

	if err := p.expect(token.HASH); err != nil {
		t.Fatal(err)
	}
	comment := p.parseComment(p.next())

	if comment.Text != " A comment" {
		t.Errorf("Wrong comment text: got %s, wanted %s", comment.Text, " A comment")
	}

}

func TestParseIdent(t *testing.T) {
	identStream := []token.Token{
		newToken(token.IDENT, "GLOBAL"),
	}
	p := &Parser{
		lexer:     &testLexer{stream: identStream},
		buffer:    [3]token.Token{},
		peekCount: 0,
	}

	ident := p.parseIdent(p.next())

	if ident.Name != "GLOBAL" {
		t.Errorf("Wrong ident name: got %s, wanted %s", ident.Name, "GLOBAL")
	}

}

func TestParseAssign(t *testing.T) {
	tests := []struct {
		want   *ast.Assign
		name   string
		stream []token.Token
	}{
		{
			name: "string rhs",
			stream: []token.Token{
				newToken(token.IDENT, "GLOBAL"),
				tDeclare,
				newToken(token.STRING, "hello"),
			},
			want: &ast.Assign{
				Name:     &ast.Ident{Name: "GLOBAL", NodeType: ast.NodeIdent},
				Value:    &ast.String{Text: "hello", NodeType: ast.NodeString},
				NodeType: ast.NodeAssign,
			},
		},
		{
			name: "ident rhs",
			stream: []token.Token{
				newToken(token.IDENT, "GLOBAL"),
				tDeclare,
				newToken(token.STRING, "VARIABLE"),
			},
			want: &ast.Assign{
				Name:     &ast.Ident{Name: "GLOBAL", NodeType: ast.NodeIdent},
				Value:    &ast.String{Text: "VARIABLE", NodeType: ast.NodeString},
				NodeType: ast.NodeAssign,
			},
		},
		{
			name: "function rhs",
			stream: []token.Token{
				newToken(token.IDENT, "GIT_COMMIT"),
				tDeclare,
				newToken(token.IDENT, "exec"),
				tLParen,
				newToken(token.STRING, "git rev-parse HEAD"),
				tRParen,
			},
			want: &ast.Assign{
				Name: &ast.Ident{Name: "GIT_COMMIT", NodeType: ast.NodeIdent},
				Value: &ast.Function{
					Name: &ast.Ident{
						Name:     "exec",
						NodeType: ast.NodeIdent,
					},
					Arguments: []ast.Node{
						&ast.String{
							Text:     "git rev-parse HEAD",
							NodeType: ast.NodeString,
						},
					},
					NodeType: ast.NodeFunction,
				},
				NodeType: ast.NodeAssign,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{
				lexer:     &testLexer{stream: tt.stream},
				buffer:    [3]token.Token{},
				peekCount: 0,
			}

			ident := p.next()
			if err := p.expect(token.DECLARE); err != nil {
				t.Fatal(err)
			}
			assign := p.parseAssign(ident)

			if !reflect.DeepEqual(assign, tt.want) {
				t.Errorf("got %v, wanted %v", assign, tt.want)
			}
		})
	}

}
