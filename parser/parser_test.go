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
	// tTask    = newToken(token.TASK, "task")
	// tLParen  = newToken(token.LPAREN, "(")
	// tRParen  = newToken(token.RPAREN, ")")
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
	comment, ok := node.(*ast.CommentNode)
	if !ok {
		t.Fatalf("Node was not a comment node, got %T", node)
	}

	if comment.Text != " A comment" {
		t.Errorf("Wrong comment text: got %s, wanted %s", comment.Text, " A comment")
	}

}

func TestParseAssign(t *testing.T) {
	assignStream := []token.Token{
		newToken(token.IDENT, "GLOBAL"),
		tDeclare,
		newToken(token.STRING, "hello"),
	}
	p := &Parser{
		lexer:     &testLexer{stream: assignStream},
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
	assign, ok := node.(*ast.AssignNode)
	if !ok {
		t.Fatalf("Node was not an assign node, got %T", node)
	}

	want := &ast.AssignNode{
		Name:     &ast.IdentNode{Name: "GLOBAL", NodeType: ast.NodeIdent},
		Value:    &ast.StringNode{Text: "hello", NodeType: ast.NodeString},
		NodeType: ast.NodeAssign,
	}

	if !reflect.DeepEqual(assign, want) {
		t.Errorf("got %#v, wanted %#v", assign, want)
	}
}

// This is the same token stream as the result of the lexer integration test
// designed to function as an integration test for the parser too, this way
// we know the lexer produces these tokens, and the parser is capable of
// converting them into the correct ast nodes, thus our parsing system
// as a whole works on an entire spokfile.
var fullSpokfileStream = []token.Token{
	tHash,
	newToken(token.COMMENT, " This is a top level comment"),
	tHash,
	newToken(token.COMMENT, " This variable is presumably important later"),
	newToken(token.IDENT, "GLOBAL"),
	tDeclare,
	newToken(token.STRING, `"very important stuff here"`),
	// TODO: Uncomment these as new parsing methods are added
	// newToken(token.IDENT, "GIT_COMMIT"),
	// tDeclare,
	// newToken(token.IDENT, "exec"),
	// tLParen,
	// newToken(token.STRING, `"git rev-parse HEAD"`),
	// tRParen,
	// tHash,
	// newToken(token.COMMENT, " Run the project unit tests"),
	// tTask,
	// newToken(token.IDENT, "test"),
	// tLParen,
	// newToken(token.IDENT, "fmt"),
	// tRParen,
	// tLBrace,
	// newToken(token.COMMAND, "go test -race ./..."),
	// tRBrace,
	// tHash,
	// newToken(token.COMMENT, " Format the project source"),
	// tTask,
	// newToken(token.IDENT, "fmt"),
	// tLParen,
	// newToken(token.STRING, `"**/*.go"`),
	// tRParen,
	// tLBrace,
	// newToken(token.COMMAND, "go fmt ./..."),
	// tRBrace,
	// tHash,
	// newToken(token.COMMENT, " Do many things"),
	// tTask,
	// newToken(token.IDENT, "many"),
	// tLParen,
	// tRParen,
	// tLBrace,
	// newToken(token.COMMAND, "line 1"),
	// newToken(token.COMMAND, "line 2"),
	// newToken(token.COMMAND, "line 3"),
	// newToken(token.COMMAND, "line 4"),
	// tRBrace,
	// tHash,
	// newToken(token.COMMENT, " Compile the project"),
	// tTask,
	// newToken(token.IDENT, "build"),
	// tLParen,
	// newToken(token.STRING, `"**/*.go"`),
	// tRParen,
	// tOutput,
	// newToken(token.STRING, `"./bin/main"`),
	// tLBrace,
	// newToken(token.COMMAND, `go build -ldflags="-X main.version=test -X main.commit=7cb0ec5609efb5fe0"`),
	// tRBrace,
	// tHash,
	// newToken(token.COMMENT, " Show the global variables"),
	// tTask,
	// newToken(token.IDENT, "show"),
	// tLParen,
	// tRParen,
	// tLBrace,
	// newToken(token.COMMAND, "echo GLOBAL"),
	// tRBrace,
	// tHash,
	// newToken(token.COMMENT, " Generate multiple outputs"),
	// tTask,
	// newToken(token.IDENT, "moar_things"),
	// tLParen,
	// tRParen,
	// tOutput,
	// tLParen,
	// newToken(token.STRING, `"output1.go"`),
	// newToken(token.STRING, `"output2.go"`),
	// tRParen,
	// tLBrace,
	// newToken(token.COMMAND, "do some stuff here"),
	// tRBrace,
	// tTask,
	// newToken(token.IDENT, "no_comment"),
	// tLParen,
	// tRParen,
	// tLBrace,
	// newToken(token.COMMAND, `echo "this task has no docstring"`),
	// tRBrace,
	// tHash,
	// newToken(token.COMMENT, " Generate output from a variable"),
	// tTask,
	// newToken(token.IDENT, "makedocs"),
	// tLParen,
	// tRParen,
	// tOutput,
	// newToken(token.IDENT, "DOCS"),
	// tLBrace,
	// newToken(token.COMMAND, `echo "making docs"`),
	// tRBrace,
	// tHash,
	// newToken(token.COMMENT, " Generate multiple outputs in variables"),
	// tTask,
	// newToken(token.IDENT, "makestuff"),
	// tLParen,
	// tRParen,
	// tOutput,
	// tLParen,
	// newToken(token.IDENT, "DOCS"),
	// newToken(token.IDENT, "BUILD"),
	// tRParen,
	// tLBrace,
	// newToken(token.COMMAND, `echo "doing things"`),
	// tRBrace,
	// tEOF,
}

func TestParserIntegration(t *testing.T) {
	p := &Parser{
		lexer:     &testLexer{stream: fullSpokfileStream},
		buffer:    [3]token.Token{},
		peekCount: 0,
	}

	// The actual parsed AST
	tree, err := p.Parse()
	if err != nil {
		t.Fatalf("Parser returned an error token: %v", err)
	}

	if tree == nil {
		t.Fatalf("Parser returned a nil AST")
	}

	// The AST we want to end up with
	want := &ast.Tree{
		Nodes: []ast.Node{
			&ast.CommentNode{
				Text:     " This is a top level comment",
				NodeType: ast.NodeComment,
			},
			&ast.CommentNode{
				Text:     " This variable is presumably important later",
				NodeType: ast.NodeComment,
			},
			&ast.AssignNode{
				Name: &ast.IdentNode{
					Name:     "GLOBAL",
					NodeType: ast.NodeIdent,
				},
				Value: &ast.StringNode{
					Text:     `"very important stuff here"`,
					NodeType: ast.NodeString,
				},
				NodeType: ast.NodeAssign,
			},
		},
	}

	if len(tree.Nodes) != len(want.Nodes) {
		t.Errorf("wrong number of ast nodes: got %d, wanted %d", len(tree.Nodes), len(want.Nodes))
	}

	if !reflect.DeepEqual(tree, want) {
		t.Errorf("got %v, wanted %v", tree, want)
	}
}

// BenchmarkParseFullSpokfile determines the performance of parsing the integration spokfile above.
func BenchmarkParseFullSpokfile(b *testing.B) {
	p := &Parser{
		lexer:     &testLexer{stream: fullSpokfileStream},
		buffer:    [3]token.Token{},
		peekCount: 0,
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := p.Parse()
		if err != nil {
			b.Fatalf("Parser returned an error token: %v", err)
		}
	}

}
