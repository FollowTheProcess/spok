package parser

import (
	"testing"

	"github.com/FollowTheProcess/spok/ast"
	"github.com/FollowTheProcess/spok/token"
	"github.com/google/go-cmp/cmp"
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

func TestEOF(t *testing.T) {
	p := &Parser{
		lexer:     &testLexer{stream: []token.Token{tEOF}},
		buffer:    [3]token.Token{},
		peekCount: 0,
	}

	tree, err := p.Parse()
	if err != nil {
		t.Fatalf("Parser returned an error: %s", err)
	}

	if !tree.IsEmpty() {
		t.Errorf("Expected an empty AST")
	}
}

func TestExpect(t *testing.T) {
	p := &Parser{
		lexer:     &testLexer{stream: []token.Token{newToken(token.STRING, "hello"), tEOF}},
		buffer:    [3]token.Token{},
		peekCount: 0,
	}

	p.expect(token.IDENT)

	if len(p.errors) != 1 {
		t.Errorf("Wrong number of errors: got %d, wanted %d", len(p.errors), 1)
	}

	want := `Unexpected token: got "hello", expected IDENT`
	err := p.popError()
	if err.Error() != want {
		t.Errorf("Wrong error message: got %s, wanted %s", err.Error(), want)
	}
}

func TestParseError(t *testing.T) {
	stream := []token.Token{
		newToken(token.IDENT, "TEST"),
		tDeclare,
		newToken(token.ERROR, "SyntaxError: Unexpected token '2' (Line 1, Position 8)"),
		tEOF,
	}
	p := &Parser{
		lexer:     &testLexer{stream: stream},
		buffer:    [3]token.Token{},
		peekCount: 0,
	}

	_, err := p.Parse()
	if err == nil {
		t.Fatalf("Expected an error but got nil")
	}

	want := "SyntaxError: Unexpected token '2' (Line 1, Position 8)"
	if err.Error() != want {
		t.Errorf("Wrong error message: got %s, wanted %s", err.Error(), want)
	}
}

func TestParseComment(t *testing.T) {
	commentStream := []token.Token{tHash, newToken(token.COMMENT, " A comment"), tEOF}
	p := &Parser{
		lexer:     &testLexer{stream: commentStream},
		buffer:    [3]token.Token{},
		peekCount: 0,
	}

	p.next() // #

	comment := p.parseComment()

	if comment.Text != " A comment" {
		t.Errorf("Wrong comment text: got %s, wanted %s", comment.Text, " A comment")
	}

}

func TestParseIdent(t *testing.T) {
	identStream := []token.Token{
		newToken(token.IDENT, "GLOBAL"),
		tEOF,
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

func TestParseFunction(t *testing.T) {
	tests := []struct {
		name   string
		stream []token.Token
		want   ast.Function
	}{
		{
			name: "exec",
			stream: []token.Token{
				newToken(token.IDENT, "exec"),
				tLParen,
				newToken(token.STRING, "git rev-parse HEAD"),
				tRParen,
				tEOF,
			},
			want: ast.Function{
				Name: ast.Ident{
					Name:     "exec",
					NodeType: ast.NodeIdent,
				},
				Arguments: []ast.Node{
					ast.String{
						Text:     "git rev-parse HEAD",
						NodeType: ast.NodeString,
					},
				},
				NodeType: ast.NodeFunction,
			},
		},
		{
			name: "join",
			stream: []token.Token{
				newToken(token.IDENT, "join"),
				tLParen,
				newToken(token.IDENT, "ROOT"),
				newToken(token.STRING, "docs"),
				tRParen,
				tEOF,
			},
			want: ast.Function{
				Name: ast.Ident{
					Name:     "join",
					NodeType: ast.NodeIdent,
				},
				Arguments: []ast.Node{
					ast.Ident{
						Name:     "ROOT",
						NodeType: ast.NodeIdent,
					},
					ast.String{
						Text:     "docs",
						NodeType: ast.NodeString,
					},
				},
				NodeType: ast.NodeFunction,
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
			fn := p.parseFunction(p.next())

			if diff := cmp.Diff(tt.want, fn); diff != "" {
				t.Errorf("Function mismatch (-want +assign):\n%s", diff)
			}
		})
	}
}

func TestParseAssign(t *testing.T) {
	tests := []struct {
		name   string
		stream []token.Token
		want   ast.Assign
	}{
		{
			name: "string rhs",
			stream: []token.Token{
				newToken(token.IDENT, "GLOBAL"),
				tDeclare,
				newToken(token.STRING, "hello"),
				tEOF,
			},
			want: ast.Assign{
				Name:     ast.Ident{Name: "GLOBAL", NodeType: ast.NodeIdent},
				Value:    ast.String{Text: "hello", NodeType: ast.NodeString},
				NodeType: ast.NodeAssign,
			},
		},
		{
			name: "ident rhs",
			stream: []token.Token{
				newToken(token.IDENT, "GLOBAL"),
				tDeclare,
				newToken(token.STRING, "VARIABLE"),
				tEOF,
			},
			want: ast.Assign{
				Name:     ast.Ident{Name: "GLOBAL", NodeType: ast.NodeIdent},
				Value:    ast.String{Text: "VARIABLE", NodeType: ast.NodeString},
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
				tEOF,
			},
			want: ast.Assign{
				Name: ast.Ident{Name: "GIT_COMMIT", NodeType: ast.NodeIdent},
				Value: ast.Function{
					Name: ast.Ident{
						Name:     "exec",
						NodeType: ast.NodeIdent,
					},
					Arguments: []ast.Node{
						ast.String{
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

			if diff := cmp.Diff(tt.want, assign); diff != "" {
				t.Errorf("Assign mismatch (-want +assign):\n%s", diff)
			}
		})
	}

}

func TestParseTask(t *testing.T) {
	tests := []struct {
		name   string
		stream []token.Token
		want   ast.Task
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
				tEOF,
			},
			want: ast.Task{
				Name: ast.Ident{
					Name:     "test",
					NodeType: ast.NodeIdent,
				},
				Docstring:    ast.Comment{NodeType: ast.NodeComment},
				Dependencies: []ast.Node{},
				Outputs:      []ast.Node{},
				Commands: []ast.Command{
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
				tEOF,
			},
			want: ast.Task{
				Name: ast.Ident{
					Name:     "test",
					NodeType: ast.NodeIdent,
				},
				Docstring: ast.Comment{
					Text:     " This one has a docstring",
					NodeType: ast.NodeComment,
				},
				Dependencies: []ast.Node{},
				Outputs:      []ast.Node{},
				Commands: []ast.Command{
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
				tEOF,
			},
			want: ast.Task{
				Name: ast.Ident{
					Name:     "build",
					NodeType: ast.NodeIdent,
				},
				Docstring: ast.Comment{NodeType: ast.NodeComment},
				Dependencies: []ast.Node{
					ast.String{
						Text:     "file.go",
						NodeType: ast.NodeString,
					},
				},
				Outputs: []ast.Node{},
				Commands: []ast.Command{
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
				tEOF,
			},
			want: ast.Task{
				Name: ast.Ident{
					Name:     "build",
					NodeType: ast.NodeIdent,
				},
				Docstring: ast.Comment{NodeType: ast.NodeComment},
				Dependencies: []ast.Node{
					ast.Ident{
						Name:     "VARIABLE",
						NodeType: ast.NodeIdent,
					},
				},
				Outputs: []ast.Node{},
				Commands: []ast.Command{
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
				tEOF,
			},
			want: ast.Task{
				Name: ast.Ident{
					Name:     "build",
					NodeType: ast.NodeIdent,
				},
				Docstring: ast.Comment{NodeType: ast.NodeComment},
				Dependencies: []ast.Node{
					ast.String{
						Text:     "file.go",
						NodeType: ast.NodeString,
					},
					ast.String{
						Text:     "file2.go",
						NodeType: ast.NodeString,
					},
				},
				Outputs: []ast.Node{},
				Commands: []ast.Command{
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
				tEOF,
			},
			want: ast.Task{
				Name: ast.Ident{
					Name:     "build",
					NodeType: ast.NodeIdent,
				},
				Docstring: ast.Comment{NodeType: ast.NodeComment},
				Dependencies: []ast.Node{
					ast.Ident{
						Name:     "VARIABLE",
						NodeType: ast.NodeIdent,
					},
					ast.Ident{
						Name:     "VARIABLE2",
						NodeType: ast.NodeIdent,
					},
				},
				Outputs: []ast.Node{},
				Commands: []ast.Command{
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
				tEOF,
			},
			want: ast.Task{
				Name: ast.Ident{
					Name:     "build",
					NodeType: ast.NodeIdent,
				},
				Docstring: ast.Comment{NodeType: ast.NodeComment},
				Dependencies: []ast.Node{
					ast.String{
						Text:     "file.go",
						NodeType: ast.NodeString,
					},
					ast.Ident{
						Name:     "FILE",
						NodeType: ast.NodeIdent,
					},
				},
				Outputs: []ast.Node{},
				Commands: []ast.Command{
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
				tEOF,
			},
			want: ast.Task{
				Name: ast.Ident{
					Name:     "build",
					NodeType: ast.NodeIdent,
				},
				Docstring: ast.Comment{NodeType: ast.NodeComment},
				Dependencies: []ast.Node{
					ast.String{
						Text:     "**/*.go",
						NodeType: ast.NodeString,
					},
				},
				Outputs: []ast.Node{
					ast.String{
						Text:     "./bin/main",
						NodeType: ast.NodeString,
					},
				},
				Commands: []ast.Command{
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
				tEOF,
			},
			want: ast.Task{
				Name: ast.Ident{
					Name:     "build",
					NodeType: ast.NodeIdent,
				},
				Docstring: ast.Comment{NodeType: ast.NodeComment},
				Dependencies: []ast.Node{
					ast.String{
						Text:     "**/*.go",
						NodeType: ast.NodeString,
					},
				},
				Outputs: []ast.Node{
					ast.Ident{
						Name:     "BIN",
						NodeType: ast.NodeIdent,
					},
				},
				Commands: []ast.Command{
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
				tEOF,
			},
			want: ast.Task{
				Name: ast.Ident{
					Name:     "build",
					NodeType: ast.NodeIdent,
				},
				Docstring: ast.Comment{NodeType: ast.NodeComment},
				Dependencies: []ast.Node{
					ast.String{
						Text:     "**/*.go",
						NodeType: ast.NodeString,
					},
				},
				Outputs: []ast.Node{
					ast.String{
						Text:     "output1",
						NodeType: ast.NodeString,
					},
					ast.String{
						Text:     "output2",
						NodeType: ast.NodeString,
					},
					ast.String{
						Text:     "output3",
						NodeType: ast.NodeString,
					},
				},
				Commands: []ast.Command{
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
				tEOF,
			},
			want: ast.Task{
				Name: ast.Ident{
					Name:     "build",
					NodeType: ast.NodeIdent,
				},
				Docstring: ast.Comment{NodeType: ast.NodeComment},
				Dependencies: []ast.Node{
					ast.String{
						Text:     "**/*.go",
						NodeType: ast.NodeString,
					},
				},
				Outputs: []ast.Node{
					ast.Ident{
						Name:     "BIN",
						NodeType: ast.NodeIdent,
					},
					ast.Ident{
						Name:     "BIN2",
						NodeType: ast.NodeIdent,
					},
					ast.Ident{
						Name:     "BIN3",
						NodeType: ast.NodeIdent,
					},
				},
				Commands: []ast.Command{
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
				tEOF,
			},
			want: ast.Task{
				Name: ast.Ident{
					Name:     "build",
					NodeType: ast.NodeIdent,
				},
				Docstring: ast.Comment{NodeType: ast.NodeComment},
				Dependencies: []ast.Node{
					ast.String{
						Text:     "**/*.go",
						NodeType: ast.NodeString,
					},
				},
				Outputs: []ast.Node{
					ast.String{
						Text:     "output1",
						NodeType: ast.NodeString,
					},
					ast.Ident{
						Name:     "SOMETHING",
						NodeType: ast.NodeIdent,
					},
				},
				Commands: []ast.Command{
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
				tEOF,
			},
			want: ast.Task{
				Name: ast.Ident{
					Name:     "complex",
					NodeType: ast.NodeIdent,
				},
				Docstring: ast.Comment{
					Text:     " A complex task with every component",
					NodeType: ast.NodeComment,
				},
				Dependencies: []ast.Node{
					ast.String{
						Text:     "**/*.go",
						NodeType: ast.NodeString,
					},
					ast.Ident{
						Name:     "fmt",
						NodeType: ast.NodeIdent,
					},
				},
				Outputs: []ast.Node{
					ast.String{
						Text:     "./bin/main",
						NodeType: ast.NodeString,
					},
					ast.Ident{
						Name:     "SOMETHINGELSE",
						NodeType: ast.NodeIdent,
					},
				},
				Commands: []ast.Command{
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

			// Simulate us passing in the ast.Comment if it was encountered by the
			// top level parse method.
			var comment ast.Comment
			if tt.stream[0].Is(token.HASH) {
				p.next() // #
				comment = ast.Comment{
					Text:     p.next().Value, // Comment
					NodeType: ast.NodeComment,
				}
			}
			p.next() // task keyword
			task := p.parseTask(comment)

			if diff := cmp.Diff(tt.want, task); diff != "" {
				t.Errorf("Task mismatch (-want +task):\n%s", diff)
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
	newToken(token.IDENT, "GIT_COMMIT"),
	tDeclare,
	newToken(token.IDENT, "exec"),
	tLParen,
	newToken(token.STRING, "git rev-parse HEAD"),
	tRParen,
	tHash,
	newToken(token.COMMENT, " Run the project unit tests"),
	tTask,
	newToken(token.IDENT, "test"),
	tLParen,
	newToken(token.IDENT, "fmt"),
	tRParen,
	tLBrace,
	newToken(token.COMMAND, "go test -race ./..."),
	tRBrace,
	tHash,
	newToken(token.COMMENT, " Format the project source"),
	tTask,
	newToken(token.IDENT, "fmt"),
	tLParen,
	newToken(token.STRING, "**/*.go"),
	tRParen,
	tLBrace,
	newToken(token.COMMAND, "go fmt ./..."),
	tRBrace,
	tHash,
	newToken(token.COMMENT, " Do many things"),
	tTask,
	newToken(token.IDENT, "many"),
	tLParen,
	tRParen,
	tLBrace,
	newToken(token.COMMAND, "line 1"),
	newToken(token.COMMAND, "line 2"),
	newToken(token.COMMAND, "line 3"),
	newToken(token.COMMAND, "line 4"),
	tRBrace,
	tHash,
	newToken(token.COMMENT, " Compile the project"),
	tTask,
	newToken(token.IDENT, "build"),
	tLParen,
	newToken(token.STRING, "**/*.go"),
	tRParen,
	tOutput,
	newToken(token.STRING, "./bin/main"),
	tLBrace,
	newToken(token.COMMAND, `go build -ldflags="-X main.version=test -X main.commit=7cb0ec5609efb5fe0"`),
	tRBrace,
	tHash,
	newToken(token.COMMENT, " Show the global variables"),
	tTask,
	newToken(token.IDENT, "show"),
	tLParen,
	tRParen,
	tLBrace,
	newToken(token.COMMAND, "echo GLOBAL"),
	tRBrace,
	tHash,
	newToken(token.COMMENT, " Generate multiple outputs"),
	tTask,
	newToken(token.IDENT, "moar_things"),
	tLParen,
	tRParen,
	tOutput,
	tLParen,
	newToken(token.STRING, "output1.go"),
	newToken(token.STRING, "output2.go"),
	tRParen,
	tLBrace,
	newToken(token.COMMAND, "do some stuff here"),
	tRBrace,
	tTask,
	newToken(token.IDENT, "no_comment"),
	tLParen,
	tRParen,
	tLBrace,
	newToken(token.COMMAND, `echo "this task has no docstring"`),
	tRBrace,
	tHash,
	newToken(token.COMMENT, " Generate output from a variable"),
	tTask,
	newToken(token.IDENT, "makedocs"),
	tLParen,
	tRParen,
	tOutput,
	newToken(token.IDENT, "DOCS"),
	tLBrace,
	newToken(token.COMMAND, `echo "making docs"`),
	tRBrace,
	tHash,
	newToken(token.COMMENT, " Generate multiple outputs in variables"),
	tTask,
	newToken(token.IDENT, "makestuff"),
	tLParen,
	tRParen,
	tOutput,
	tLParen,
	newToken(token.IDENT, "DOCS"),
	newToken(token.IDENT, "BUILD"),
	tRParen,
	tLBrace,
	newToken(token.COMMAND, `echo "doing things"`),
	tRBrace,
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

	if tree.IsEmpty() {
		t.Fatal("Parser produced an empty AST")
	}

	want := ast.Tree{
		Nodes: []ast.Node{
			ast.Comment{
				Text:     " This is a top level comment",
				NodeType: ast.NodeComment,
			},
			ast.Comment{
				Text:     " This variable is presumably important later",
				NodeType: ast.NodeComment,
			},
			ast.Assign{
				Name: ast.Ident{
					Name:     "GLOBAL",
					NodeType: ast.NodeIdent,
				},
				Value: ast.String{
					Text:     "very important stuff here",
					NodeType: ast.NodeString,
				},
				NodeType: ast.NodeAssign,
			},
			ast.Assign{
				Value: ast.Function{
					Name: ast.Ident{
						Name:     "exec",
						NodeType: ast.NodeIdent,
					},
					Arguments: []ast.Node{
						ast.String{
							Text:     "git rev-parse HEAD",
							NodeType: ast.NodeString,
						},
					}, NodeType: ast.NodeFunction,
				},
				Name: ast.Ident{
					Name:     "GIT_COMMIT",
					NodeType: ast.NodeIdent,
				},
				NodeType: ast.NodeAssign,
			},
			ast.Task{
				Name: ast.Ident{
					Name:     "test",
					NodeType: ast.NodeIdent,
				},
				Docstring: ast.Comment{
					Text:     " Run the project unit tests",
					NodeType: ast.NodeComment,
				},
				Dependencies: []ast.Node{
					ast.Ident{
						Name:     "fmt",
						NodeType: ast.NodeIdent,
					},
				},
				Outputs: []ast.Node{},
				Commands: []ast.Command{
					{
						Command:  "go test -race ./...",
						NodeType: ast.NodeCommand,
					},
				},
				NodeType: ast.NodeTask,
			},
			ast.Task{
				Name: ast.Ident{
					Name:     "fmt",
					NodeType: ast.NodeIdent,
				},
				Docstring: ast.Comment{
					Text:     " Format the project source",
					NodeType: ast.NodeComment,
				},
				Dependencies: []ast.Node{
					ast.String{
						Text:     "**/*.go",
						NodeType: ast.NodeString,
					},
				},
				Outputs: []ast.Node{},
				Commands: []ast.Command{
					{
						Command:  "go fmt ./...",
						NodeType: ast.NodeCommand,
					},
				},
				NodeType: ast.NodeTask,
			},
			ast.Task{
				Name: ast.Ident{
					Name:     "many",
					NodeType: ast.NodeIdent,
				},
				Docstring: ast.Comment{
					Text:     " Do many things",
					NodeType: ast.NodeComment,
				},
				Dependencies: []ast.Node{},
				Outputs:      []ast.Node{},
				Commands: []ast.Command{
					{
						Command:  "line 1",
						NodeType: ast.NodeCommand,
					},
					{
						Command:  "line 2",
						NodeType: ast.NodeCommand,
					},
					{
						Command:  "line 3",
						NodeType: ast.NodeCommand,
					},
					{
						Command:  "line 4",
						NodeType: ast.NodeCommand,
					},
				},
				NodeType: ast.NodeTask,
			},
			ast.Task{
				Name: ast.Ident{
					Name:     "build",
					NodeType: ast.NodeIdent,
				},
				Docstring: ast.Comment{
					Text:     " Compile the project",
					NodeType: ast.NodeComment,
				},
				Dependencies: []ast.Node{
					ast.String{
						Text:     "**/*.go",
						NodeType: ast.NodeString,
					},
				},
				Outputs: []ast.Node{
					ast.String{
						Text:     "./bin/main",
						NodeType: ast.NodeString,
					},
				},
				Commands: []ast.Command{
					{
						Command:  `go build -ldflags="-X main.version=test -X main.commit=7cb0ec5609efb5fe0"`,
						NodeType: ast.NodeCommand,
					},
				},
				NodeType: ast.NodeTask,
			},
			ast.Task{
				Name: ast.Ident{
					Name:     "show",
					NodeType: ast.NodeIdent,
				},
				Docstring: ast.Comment{
					Text:     " Show the global variables",
					NodeType: ast.NodeComment,
				},
				Dependencies: []ast.Node{},
				Outputs:      []ast.Node{},
				Commands: []ast.Command{
					{
						Command:  "echo GLOBAL",
						NodeType: ast.NodeCommand,
					},
				},
				NodeType: ast.NodeTask,
			},
			ast.Task{
				Name: ast.Ident{
					Name:     "moar_things",
					NodeType: ast.NodeIdent,
				},
				Docstring: ast.Comment{
					Text:     " Generate multiple outputs",
					NodeType: ast.NodeComment,
				},
				Dependencies: []ast.Node{},
				Outputs: []ast.Node{
					ast.String{
						Text:     "output1.go",
						NodeType: ast.NodeString,
					},
					ast.String{
						Text:     "output2.go",
						NodeType: ast.NodeString,
					},
				},
				Commands: []ast.Command{
					{
						Command:  "do some stuff here",
						NodeType: ast.NodeCommand,
					},
				},
				NodeType: ast.NodeTask,
			},
			ast.Task{
				Name: ast.Ident{
					Name:     "no_comment",
					NodeType: ast.NodeIdent,
				},
				Docstring:    ast.Comment{},
				Dependencies: []ast.Node{},
				Outputs:      []ast.Node{},
				Commands: []ast.Command{
					{
						Command:  `echo "this task has no docstring"`,
						NodeType: ast.NodeCommand,
					},
				},
				NodeType: ast.NodeTask,
			},
			ast.Task{
				Name: ast.Ident{
					Name:     "makedocs",
					NodeType: ast.NodeIdent,
				},
				Docstring: ast.Comment{
					Text:     " Generate output from a variable",
					NodeType: ast.NodeComment,
				},
				Dependencies: []ast.Node{},
				Outputs: []ast.Node{
					ast.Ident{
						Name:     "DOCS",
						NodeType: ast.NodeIdent,
					},
				},
				Commands: []ast.Command{
					{
						Command:  `echo "making docs"`,
						NodeType: ast.NodeCommand,
					},
				},
				NodeType: ast.NodeTask,
			},
			ast.Task{
				Name: ast.Ident{
					Name:     "makestuff",
					NodeType: ast.NodeIdent,
				},
				Docstring: ast.Comment{
					Text:     " Generate multiple outputs in variables",
					NodeType: ast.NodeComment,
				},
				Dependencies: []ast.Node{},
				Outputs: []ast.Node{
					ast.Ident{
						Name:     "DOCS",
						NodeType: ast.NodeIdent,
					},
					ast.Ident{
						Name:     "BUILD",
						NodeType: ast.NodeIdent,
					},
				},
				Commands: []ast.Command{
					{
						Command:  `echo "doing things"`,
						NodeType: ast.NodeCommand,
					},
				},
				NodeType: ast.NodeTask,
			},
		},
	}

	if diff := cmp.Diff(want, tree); diff != "" {
		t.Errorf("AST mismatch (-want +tree):\n%s", diff)
	}
}

func BenchmarkParseFullSpokfile(b *testing.B) {
	p := &Parser{
		lexer: &testLexer{
			stream: fullSpokfileStream,
		},
		buffer:    [3]token.Token{},
		peekCount: 0,
	}

	b.ResetTimer()
	_, err := p.Parse()
	if err != nil {
		b.Fatalf("Parser produced an error: %v", err)
	}
}
