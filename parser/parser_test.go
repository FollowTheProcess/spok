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
	tTask    = newToken(token.TASK, "task")
	tLBrace  = newToken(token.LBRACE, "{")
	tRBrace  = newToken(token.RBRACE, "}")
	tOutput  = newToken(token.OUTPUT, "->")
	tEOF     = newToken(token.EOF, "")
)

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

	comment := p.parseComment()

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
			assign := p.parseAssign(ident)

			if !reflect.DeepEqual(assign, tt.want) {
				t.Errorf("got %v, wanted %v", assign, tt.want)
			}
		})
	}

}

func TestParseTask(t *testing.T) {
	tests := []struct {
		want   *ast.Task
		name   string
		stream []token.Token
	}{
		{
			name: "basic",
			stream: []token.Token{
				tTask,
				newToken(token.IDENT, "test"),
				tLParen,
				tRParen,
				tLBrace,
				newToken(token.COMMAND, "go test ./..."),
				tRBrace,
			},
			want: &ast.Task{
				Name: &ast.Ident{
					Name:     "test",
					NodeType: ast.NodeIdent,
				},
				Docstring:    &ast.Comment{NodeType: ast.NodeComment},
				Dependencies: []ast.Node{},
				Outputs:      []ast.Node{},
				Commands: []*ast.Command{
					{
						Command:  "go test ./...",
						NodeType: ast.NodeCommand,
					},
				},
				NodeType: ast.NodeTask,
			},
		},
		{
			name: "basic with docstring",
			stream: []token.Token{
				tHash,
				newToken(token.COMMENT, " This one has a docstring"),
				tTask,
				newToken(token.IDENT, "test"),
				tLParen,
				tRParen,
				tLBrace,
				newToken(token.COMMAND, "go test ./..."),
				tRBrace,
			},
			want: &ast.Task{
				Name: &ast.Ident{
					Name:     "test",
					NodeType: ast.NodeIdent,
				},
				Docstring: &ast.Comment{
					Text:     " This one has a docstring",
					NodeType: ast.NodeComment,
				},
				Dependencies: []ast.Node{},
				Outputs:      []ast.Node{},
				Commands: []*ast.Command{
					{
						Command:  "go test ./...",
						NodeType: ast.NodeCommand,
					},
				},
				NodeType: ast.NodeTask,
			},
		},
		{
			name: "string dependency",
			stream: []token.Token{
				tTask,
				newToken(token.IDENT, "build"),
				tLParen,
				newToken(token.STRING, "file.go"),
				tRParen,
				tLBrace,
				newToken(token.COMMAND, "go build"),
				tRBrace,
			},
			want: &ast.Task{
				Name: &ast.Ident{
					Name:     "build",
					NodeType: ast.NodeIdent,
				},
				Docstring: &ast.Comment{NodeType: ast.NodeComment},
				Dependencies: []ast.Node{
					&ast.String{
						Text:     "file.go",
						NodeType: ast.NodeString,
					},
				},
				Outputs: []ast.Node{},
				Commands: []*ast.Command{
					{
						Command:  "go build",
						NodeType: ast.NodeCommand,
					},
				},
				NodeType: ast.NodeTask,
			},
		},
		{
			name: "ident dependency",
			stream: []token.Token{
				tTask,
				newToken(token.IDENT, "build"),
				tLParen,
				newToken(token.IDENT, "VARIABLE"),
				tRParen,
				tLBrace,
				newToken(token.COMMAND, "go build"),
				tRBrace,
			},
			want: &ast.Task{
				Name: &ast.Ident{
					Name:     "build",
					NodeType: ast.NodeIdent,
				},
				Docstring: &ast.Comment{NodeType: ast.NodeComment},
				Dependencies: []ast.Node{
					&ast.Ident{
						Name:     "VARIABLE",
						NodeType: ast.NodeIdent,
					},
				},
				Outputs: []ast.Node{},
				Commands: []*ast.Command{
					{
						Command:  "go build",
						NodeType: ast.NodeCommand,
					},
				},
				NodeType: ast.NodeTask,
			},
		},
		{
			name: "multi string dependency",
			stream: []token.Token{
				tTask,
				newToken(token.IDENT, "build"),
				tLParen,
				newToken(token.STRING, "file.go"),
				newToken(token.STRING, "file2.go"),
				tRParen,
				tLBrace,
				newToken(token.COMMAND, "go build"),
				tRBrace,
			},
			want: &ast.Task{
				Name: &ast.Ident{
					Name:     "build",
					NodeType: ast.NodeIdent,
				},
				Docstring: &ast.Comment{NodeType: ast.NodeComment},
				Dependencies: []ast.Node{
					&ast.String{
						Text:     "file.go",
						NodeType: ast.NodeString,
					},
					&ast.String{
						Text:     "file2.go",
						NodeType: ast.NodeString,
					},
				},
				Outputs: []ast.Node{},
				Commands: []*ast.Command{
					{
						Command:  "go build",
						NodeType: ast.NodeCommand,
					},
				},
				NodeType: ast.NodeTask,
			},
		},
		{
			name: "multi ident dependency",
			stream: []token.Token{
				tTask,
				newToken(token.IDENT, "build"),
				tLParen,
				newToken(token.IDENT, "VARIABLE"),
				newToken(token.IDENT, "VARIABLE2"),
				tRParen,
				tLBrace,
				newToken(token.COMMAND, "go build"),
				tRBrace,
			},
			want: &ast.Task{
				Name: &ast.Ident{
					Name:     "build",
					NodeType: ast.NodeIdent,
				},
				Docstring: &ast.Comment{NodeType: ast.NodeComment},
				Dependencies: []ast.Node{
					&ast.Ident{
						Name:     "VARIABLE",
						NodeType: ast.NodeIdent,
					},
					&ast.Ident{
						Name:     "VARIABLE2",
						NodeType: ast.NodeIdent,
					},
				},
				Outputs: []ast.Node{},
				Commands: []*ast.Command{
					{
						Command:  "go build",
						NodeType: ast.NodeCommand,
					},
				},
				NodeType: ast.NodeTask,
			},
		},
		{
			name: "string and ident dependency",
			stream: []token.Token{
				tTask,
				newToken(token.IDENT, "build"),
				tLParen,
				newToken(token.STRING, "file.go"),
				newToken(token.IDENT, "FILE"),
				tRParen,
				tLBrace,
				newToken(token.COMMAND, "go build"),
				tRBrace,
			},
			want: &ast.Task{
				Name: &ast.Ident{
					Name:     "build",
					NodeType: ast.NodeIdent,
				},
				Docstring: &ast.Comment{NodeType: ast.NodeComment},
				Dependencies: []ast.Node{
					&ast.String{
						Text:     "file.go",
						NodeType: ast.NodeString,
					},
					&ast.Ident{
						Name:     "FILE",
						NodeType: ast.NodeIdent,
					},
				},
				Outputs: []ast.Node{},
				Commands: []*ast.Command{
					{
						Command:  "go build",
						NodeType: ast.NodeCommand,
					},
				},
				NodeType: ast.NodeTask,
			},
		},
		{
			name: "string output",
			stream: []token.Token{
				tTask,
				newToken(token.IDENT, "build"),
				tLParen,
				newToken(token.STRING, "**/*.go"),
				tRParen,
				tOutput,
				newToken(token.STRING, "./bin/main"),
				tLBrace,
				newToken(token.COMMAND, "go build"),
				tRBrace,
			},
			want: &ast.Task{
				Name: &ast.Ident{
					Name:     "build",
					NodeType: ast.NodeIdent,
				},
				Docstring: &ast.Comment{NodeType: ast.NodeComment},
				Dependencies: []ast.Node{
					&ast.String{
						Text:     "**/*.go",
						NodeType: ast.NodeString,
					},
				},
				Outputs: []ast.Node{
					&ast.String{
						Text:     "./bin/main",
						NodeType: ast.NodeString,
					},
				},
				Commands: []*ast.Command{
					{
						Command:  "go build",
						NodeType: ast.NodeCommand,
					},
				},
				NodeType: ast.NodeTask,
			},
		},
		{
			name: "ident output",
			stream: []token.Token{
				tTask,
				newToken(token.IDENT, "build"),
				tLParen,
				newToken(token.STRING, "**/*.go"),
				tRParen,
				tOutput,
				newToken(token.IDENT, "BIN"),
				tLBrace,
				newToken(token.COMMAND, "go build"),
				tRBrace,
			},
			want: &ast.Task{
				Name: &ast.Ident{
					Name:     "build",
					NodeType: ast.NodeIdent,
				},
				Docstring: &ast.Comment{NodeType: ast.NodeComment},
				Dependencies: []ast.Node{
					&ast.String{
						Text:     "**/*.go",
						NodeType: ast.NodeString,
					},
				},
				Outputs: []ast.Node{
					&ast.Ident{
						Name:     "BIN",
						NodeType: ast.NodeIdent,
					},
				},
				Commands: []*ast.Command{
					{
						Command:  "go build",
						NodeType: ast.NodeCommand,
					},
				},
				NodeType: ast.NodeTask,
			},
		},
		{
			name: "multi string output",
			stream: []token.Token{
				tTask,
				newToken(token.IDENT, "build"),
				tLParen,
				newToken(token.STRING, "**/*.go"),
				tRParen,
				tOutput,
				tLParen,
				newToken(token.STRING, "output1"),
				newToken(token.STRING, "output2"),
				newToken(token.STRING, "output3"),
				tRParen,
				tLBrace,
				newToken(token.COMMAND, "go build"),
				tRBrace,
			},
			want: &ast.Task{
				Name: &ast.Ident{
					Name:     "build",
					NodeType: ast.NodeIdent,
				},
				Docstring: &ast.Comment{NodeType: ast.NodeComment},
				Dependencies: []ast.Node{
					&ast.String{
						Text:     "**/*.go",
						NodeType: ast.NodeString,
					},
				},
				Outputs: []ast.Node{
					&ast.String{
						Text:     "output1",
						NodeType: ast.NodeString,
					},
					&ast.String{
						Text:     "output2",
						NodeType: ast.NodeString,
					},
					&ast.String{
						Text:     "output3",
						NodeType: ast.NodeString,
					},
				},
				Commands: []*ast.Command{
					{
						Command:  "go build",
						NodeType: ast.NodeCommand,
					},
				},
				NodeType: ast.NodeTask,
			},
		},
		{
			name: "multi ident output",
			stream: []token.Token{
				tTask,
				newToken(token.IDENT, "build"),
				tLParen,
				newToken(token.STRING, "**/*.go"),
				tRParen,
				tOutput,
				tLParen,
				newToken(token.IDENT, "BIN"),
				newToken(token.IDENT, "BIN2"),
				newToken(token.IDENT, "BIN3"),
				tRParen,
				tLBrace,
				newToken(token.COMMAND, "go build"),
				tRBrace,
			},
			want: &ast.Task{
				Name: &ast.Ident{
					Name:     "build",
					NodeType: ast.NodeIdent,
				},
				Docstring: &ast.Comment{NodeType: ast.NodeComment},
				Dependencies: []ast.Node{
					&ast.String{
						Text:     "**/*.go",
						NodeType: ast.NodeString,
					},
				},
				Outputs: []ast.Node{
					&ast.Ident{
						Name:     "BIN",
						NodeType: ast.NodeIdent,
					},
					&ast.Ident{
						Name:     "BIN2",
						NodeType: ast.NodeIdent,
					},
					&ast.Ident{
						Name:     "BIN3",
						NodeType: ast.NodeIdent,
					},
				},
				Commands: []*ast.Command{
					{
						Command:  "go build",
						NodeType: ast.NodeCommand,
					},
				},
				NodeType: ast.NodeTask,
			},
		},
		{
			name: "multi both output",
			stream: []token.Token{
				tTask,
				newToken(token.IDENT, "build"),
				tLParen,
				newToken(token.STRING, "**/*.go"),
				tRParen,
				tOutput,
				tLParen,
				newToken(token.STRING, "output1"),
				newToken(token.IDENT, "SOMETHING"),
				tRParen,
				tLBrace,
				newToken(token.COMMAND, "go build"),
				tRBrace,
			},
			want: &ast.Task{
				Name: &ast.Ident{
					Name:     "build",
					NodeType: ast.NodeIdent,
				},
				Docstring: &ast.Comment{NodeType: ast.NodeComment},
				Dependencies: []ast.Node{
					&ast.String{
						Text:     "**/*.go",
						NodeType: ast.NodeString,
					},
				},
				Outputs: []ast.Node{
					&ast.String{
						Text:     "output1",
						NodeType: ast.NodeString,
					},
					&ast.Ident{
						Name:     "SOMETHING",
						NodeType: ast.NodeIdent,
					},
				},
				Commands: []*ast.Command{
					{
						Command:  "go build",
						NodeType: ast.NodeCommand,
					},
				},
				NodeType: ast.NodeTask,
			},
		},
		{
			name: "complex",
			stream: []token.Token{
				tHash,
				newToken(token.COMMENT, " A complex task with every component"),
				tTask,
				newToken(token.IDENT, "complex"),
				tLParen,
				newToken(token.STRING, "**/*.go"),
				newToken(token.IDENT, "fmt"),
				tRParen,
				tOutput,
				tLParen,
				newToken(token.STRING, "./bin/main"),
				newToken(token.IDENT, "SOMETHINGELSE"),
				tRParen,
				tLBrace,
				newToken(token.COMMAND, "go fmt ./..."),
				newToken(token.COMMAND, "go test -race ./..."),
				newToken(token.COMMAND, `go build -ldflags="-X github.com/FollowTheProcess/spok/cli/cmd.version=dev"`),
				tRBrace,
			},
			want: &ast.Task{
				Name: &ast.Ident{
					Name:     "complex",
					NodeType: ast.NodeIdent,
				},
				Docstring: &ast.Comment{
					Text:     " A complex task with every component",
					NodeType: ast.NodeComment,
				},
				Dependencies: []ast.Node{
					&ast.String{
						Text:     "**/*.go",
						NodeType: ast.NodeString,
					},
					&ast.Ident{
						Name:     "fmt",
						NodeType: ast.NodeIdent,
					},
				},
				Outputs: []ast.Node{
					&ast.String{
						Text:     "./bin/main",
						NodeType: ast.NodeString,
					},
					&ast.Ident{
						Name:     "SOMETHINGELSE",
						NodeType: ast.NodeIdent,
					},
				},
				Commands: []*ast.Command{
					{
						Command:  "go fmt ./...",
						NodeType: ast.NodeCommand,
					},
					{
						Command:  "go test -race ./...",
						NodeType: ast.NodeCommand,
					},
					{
						Command:  `go build -ldflags="-X github.com/FollowTheProcess/spok/cli/cmd.version=dev"`,
						NodeType: ast.NodeCommand,
					},
				},
				NodeType: ast.NodeTask,
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

			// Simulate us passing in the *ast.Comment if it was encountered by the
			// top level parse method.
			var comment *ast.Comment
			if tt.stream[0].Is(token.HASH) {
				p.next() // #
				comment = &ast.Comment{
					Text:     p.next().Value, // Comment
					NodeType: ast.NodeComment,
				}
			}
			p.next() // task keyword
			task := p.parseTask(comment)

			if !reflect.DeepEqual(task, tt.want) {
				t.Errorf("got %v, wanted %v", task, tt.want)
			}
		})
	}
}

// fullSpokfile stream is the same stream of tokens as used as the target output in the lexer integration test
// here used as input for the parser integration test. If the parser can handle all this, we know our parsing
// system as a whole is capable of parsing every construct in a spokfile from plain text to AST.
var fullSpokfileStream = []token.Token{
	tHash,
	newToken(token.COMMENT, " This is a top level comment"),
	tHash,
	newToken(token.COMMENT, " This variable is presumably important later"),
	newToken(token.IDENT, "GLOBAL"),
	tDeclare,
	newToken(token.STRING, "very important stuff here"),
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
	tEOF,
}

func TestParserIntegration(t *testing.T) {
	p := &Parser{
		lexer: &testLexer{
			stream: fullSpokfileStream,
		},
		buffer:    [3]token.Token{},
		peekCount: 0,
	}

	tree, err := p.Parse()
	if err != nil {
		t.Fatalf("Parser produced an error: %v", err)
	}

	if tree == nil {
		t.Fatal("Parser produced a nil AST")
	}

	if tree.IsEmpty() {
		t.Fatal("Parser produced an empty AST")
	}

	want := &ast.Tree{
		Nodes: []ast.Node{
			&ast.Comment{
				Text:     " This is a top level comment",
				NodeType: ast.NodeComment,
			},
			&ast.Comment{
				Text:     " This variable is presumably important later",
				NodeType: ast.NodeComment,
			},
			&ast.Assign{
				Name: &ast.Ident{
					Name:     "GLOBAL",
					NodeType: ast.NodeIdent,
				},
				Value: &ast.String{
					Text:     "very important stuff here",
					NodeType: ast.NodeString,
				},
				NodeType: ast.NodeAssign,
			},
		},
	}

	if !reflect.DeepEqual(tree, want) {
		t.Errorf("got %s, wanted %s", tree.String(), want.String())
	}
}
