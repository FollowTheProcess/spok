package parser //nolint: testpackage // We need access to all the private parse methods

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.followtheprocess.codes/spok/ast"
	"go.followtheprocess.codes/spok/token"
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
	// If the stream is empty, return an EOF to emulate the real lexer
	// closing the token.Token channel which will cause the parser to read
	// an EOF (channel zero value)
	if len(l.stream) == 0 {
		return tEOF
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
	tComma   = newToken(token.COMMA, ",")
	tTask    = newToken(token.TASK, "task")
	tLBrace  = newToken(token.LBRACE, "{")
	tRBrace  = newToken(token.RBRACE, "}")
	tOutput  = newToken(token.OUTPUT, "->")
	tEOF     = newToken(token.EOF, "")
)

func TestEOF(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
	p := &Parser{
		lexer:     &testLexer{stream: []token.Token{newToken(token.STRING, `"hello"`), tEOF}},
		buffer:    [3]token.Token{},
		peekCount: 0,
	}

	err := p.expect(token.IDENT)
	if err == nil {
		t.Fatal("Expected an expect error, got nil")
	}

	// No line or position info because it's our fake lexer but this is where it would go
	want := "Illegal Token: [STRING] \"hello\" (Line 0). Expected 'IDENT'\n\n0 |\t"
	if err.Error() != want {
		t.Errorf("Wrong error message: got %#v, wanted %#v", err.Error(), want)
	}
}

func TestParseComment(t *testing.T) {
	t.Parallel()
	commentStream := []token.Token{tHash, newToken(token.COMMENT, " A comment"), tEOF}
	p := &Parser{
		lexer:  &testLexer{stream: commentStream},
		buffer: [3]token.Token{},
	}

	p.next() // #

	comment := p.parseComment()

	if comment.Text != " A comment" {
		t.Errorf("Wrong comment text: got %s, wanted %s", comment.Text, " A comment")
	}
}

func TestParseIdent(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
	tests := []struct {
		name    string
		stream  []token.Token
		want    ast.Function
		wantErr bool
	}{
		{
			name: "exec",
			stream: []token.Token{
				newToken(token.IDENT, "exec"),
				tLParen,
				newToken(token.STRING, `"git rev-parse HEAD"`),
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
			wantErr: false,
		},
		{
			name: "join",
			stream: []token.Token{
				newToken(token.IDENT, "join"),
				tLParen,
				newToken(token.IDENT, "ROOT"),
				tComma,
				newToken(token.STRING, `"docs"`),
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
			wantErr: false,
		},
		{
			name: "missing LParen",
			stream: []token.Token{
				newToken(token.IDENT, "join"),
				// Should be an LParen here
				newToken(token.IDENT, "ROOT"),
				tComma,
				newToken(token.STRING, `"docs"`),
				tRParen,
				tEOF,
			},
			want:    ast.Function{},
			wantErr: true,
		},
		{
			name: "lexer error token",
			stream: []token.Token{
				newToken(token.IDENT, "join"),
				tLParen,
				newToken(token.IDENT, "ROOT"),
				tComma,
				newToken(token.STRING, `"docs"`),
				newToken(token.ERROR, "beep boop"),
				tRParen,
				tEOF,
			},
			want:    ast.Function{},
			wantErr: true,
		},
		{
			name: "illegal token",
			stream: []token.Token{
				newToken(token.IDENT, "join"),
				tLParen,
				newToken(token.IDENT, "ROOT"),
				tComma,
				newToken(token.STRING, `"docs"`),
				newToken(token.TASK, "I dont belong here"),
				tRParen,
				tEOF,
			},
			want:    ast.Function{},
			wantErr: true,
		},
		{
			name: "missing RParen",
			stream: []token.Token{
				newToken(token.IDENT, "join"),
				tLParen,
				newToken(token.IDENT, "ROOT"),
				tComma,
				newToken(token.STRING, `"docs"`),
			},
			want:    ast.Function{},
			wantErr: true,
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
			fn, err := p.parseFunction(ident)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseFunction() err = %v, wantErr = %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, fn); diff != "" {
				t.Errorf("Function mismatch (-want +assign):\n%s", diff)
			}
		})
	}
}

func TestParseAssign(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		stream  []token.Token
		want    ast.Assign
		wantErr bool
	}{
		{
			name: "string rhs",
			stream: []token.Token{
				newToken(token.IDENT, "GLOBAL"),
				tDeclare,
				newToken(token.STRING, `"hello"`),
				tEOF,
			},
			want: ast.Assign{
				Name:     ast.Ident{Name: "GLOBAL", NodeType: ast.NodeIdent},
				Value:    ast.String{Text: "hello", NodeType: ast.NodeString},
				NodeType: ast.NodeAssign,
			},
			wantErr: false,
		},
		{
			name: "ident rhs",
			stream: []token.Token{
				newToken(token.IDENT, "GLOBAL"),
				tDeclare,
				newToken(token.IDENT, "VARIABLE"),
				tEOF,
			},
			want: ast.Assign{
				Name:     ast.Ident{Name: "GLOBAL", NodeType: ast.NodeIdent},
				Value:    ast.Ident{Name: "VARIABLE", NodeType: ast.NodeIdent},
				NodeType: ast.NodeAssign,
			},
			wantErr: false,
		},
		{
			name: "function rhs",
			stream: []token.Token{
				newToken(token.IDENT, "GIT_COMMIT"),
				tDeclare,
				newToken(token.IDENT, "exec"),
				tLParen,
				newToken(token.STRING, `"git rev-parse HEAD"`),
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
			wantErr: false,
		},
		{
			name: "bad function rhs",
			stream: []token.Token{
				newToken(token.IDENT, "GIT_COMMIT"),
				tDeclare,
				newToken(token.IDENT, "exec"),
				tLParen,
				newToken(token.ERROR, "beep boop"),
				tRParen,
				tEOF,
			},
			want:    ast.Assign{},
			wantErr: true,
		},
		{
			name: "illegal token",
			stream: []token.Token{
				newToken(token.IDENT, "GIT_COMMIT"),
				tDeclare,
				newToken(token.TASK, "I'm not allowed"),
				tRParen,
				tEOF,
			},
			want:    ast.Assign{},
			wantErr: true,
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
			assign, err := p.parseAssign(ident)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseAssign() err = %v, wanted %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, assign); diff != "" {
				t.Errorf("Assign mismatch (-want +assign):\n%s", diff)
			}
		})
	}
}

func TestParseTask(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		stream  []token.Token
		want    ast.Task
		wantErr bool
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
				newToken(token.STRING, `"file.go"`),
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
				newToken(token.STRING, `"file.go"`),
				tComma,
				newToken(token.STRING, `"file2.go"`),
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
				tComma,
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
				newToken(token.STRING, `"file.go"`),
				tComma,
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
				newToken(token.STRING, `"**/*.go"`),
				tRParen,
				tOutput,
				newToken(token.STRING, `"./bin/main"`),
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
				newToken(token.STRING, `"**/*.go"`),
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
				newToken(token.STRING, `"**/*.go"`),
				tRParen,
				tOutput,
				tLParen,
				newToken(token.STRING, `"output1"`),
				tComma,
				newToken(token.STRING, `"output2"`),
				tComma,
				newToken(token.STRING, `"output3"`),
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
				newToken(token.STRING, `"**/*.go"`),
				tRParen,
				tOutput,
				tLParen,
				newToken(token.IDENT, "BIN"),
				tComma,
				newToken(token.IDENT, "BIN2"),
				tComma,
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
				newToken(token.STRING, `"**/*.go"`),
				tRParen,
				tOutput,
				tLParen,
				newToken(token.STRING, `"output1"`),
				tComma,
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
				newToken(token.STRING, `"**/*.go"`),
				tComma,
				newToken(token.IDENT, "fmt"),
				tRParen,
				tOutput,
				tLParen,
				newToken(token.STRING, `"./bin/main"`),
				tComma,
				newToken(token.IDENT, "SOMETHINGELSE"),
				tRParen,
				tLBrace,
				newToken(token.COMMAND, "go fmt ./..."),
				newToken(token.COMMAND, "go test -race ./..."),
				newToken(token.COMMAND, `go build -ldflags="-X go.followtheprocess.codes/spok/cli/cmd.version=dev"`),
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
						Command:  `go build -ldflags="-X go.followtheprocess.codes/spok/cli/cmd.version=dev"`,
						NodeType: ast.NodeCommand,
					},
				},
				NodeType: ast.NodeTask,
			},
		},
		{
			name: "illegal token in dependencies",
			stream: []token.Token{
				tTask,
				newToken(token.IDENT, "build"),
				tLParen,
				newToken(token.STRING, `"file.go"`),
				tComma,
				newToken(token.HASH, "#"), // This isn't allowed
				tRParen,
				tLBrace,
				newToken(token.COMMAND, "go build"),
				tRBrace,
				tEOF,
			},
			want:    ast.Task{},
			wantErr: true,
		},
		{
			name: "illegal token in output",
			stream: []token.Token{
				tTask,
				newToken(token.IDENT, "build"),
				tLParen,
				newToken(token.STRING, `"file.go"`),
				tRParen,
				tOutput,
				newToken(token.HASH, "#"), // This isn't allowed
				tLBrace,
				newToken(token.COMMAND, "go build"),
				tRBrace,
				tEOF,
			},
			want:    ast.Task{},
			wantErr: true,
		},
		{
			name: "lexer error in output",
			stream: []token.Token{
				tTask,
				newToken(token.IDENT, "build"),
				tLParen,
				newToken(token.STRING, `"file.go"`),
				tRParen,
				tOutput,
				newToken(token.ERROR, "beep boop"),
				tLBrace,
				newToken(token.COMMAND, "go build"),
				tRBrace,
				tEOF,
			},
			want:    ast.Task{},
			wantErr: true,
		},
		{
			name: "illegal token in multi output",
			stream: []token.Token{
				tTask,
				newToken(token.IDENT, "build"),
				tLParen,
				newToken(token.STRING, `"file.go"`),
				tRParen,
				tOutput,
				tLParen,
				newToken(token.STRING, `"file.go"`),
				tComma,
				newToken(token.HASH, "#"), // This isn't allowed
				tRParen,
				tLBrace,
				newToken(token.COMMAND, "go build"),
				tRBrace,
				tEOF,
			},
			want:    ast.Task{},
			wantErr: true,
		},
		{
			name: "lexer error in multi output",
			stream: []token.Token{
				tTask,
				newToken(token.IDENT, "build"),
				tLParen,
				newToken(token.STRING, `"file.go"`),
				tRParen,
				tOutput,
				tLParen,
				newToken(token.STRING, `"file.go"`),
				newToken(token.ERROR, "beep boop"),
				tRParen,
				tLBrace,
				newToken(token.COMMAND, "go build"),
				tRBrace,
				tEOF,
			},
			want:    ast.Task{},
			wantErr: true,
		},
		{
			name: "missing LParen",
			stream: []token.Token{
				tTask,
				newToken(token.IDENT, "build"),
				// Should be an LParen here
				newToken(token.STRING, `"file.go"`),
				tRParen,
				tOutput,
				tLParen,
				newToken(token.STRING, `"file.go"`),
				newToken(token.ERROR, "beep boop"),
				tRParen,
				tLBrace,
				newToken(token.COMMAND, "go build"),
				tRBrace,
				tEOF,
			},
			want:    ast.Task{},
			wantErr: true,
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
			task, err := p.parseTask(comment)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTask() err = %v, wanted %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, task); diff != "" {
				t.Errorf("Task mismatch (-want +task):\n%s", diff)
			}
		})
	}
}

func TestParserErrorHandling(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		message string
		stream  []token.Token
	}{
		{
			name: "error token from lexer",
			stream: []token.Token{
				newToken(token.ERROR, "beep boop"),
			},
			message: "beep boop",
		},
		{
			name:    "bad input chars",
			stream:  []token.Token{newToken(token.ERROR, "SyntaxError: Unexpected token '*' (Line 1, Position 0)")},
			message: "SyntaxError: Unexpected token '*' (Line 1, Position 0)",
		},
		{
			name:    "global variable unterminated string",
			message: "SyntaxError: Unterminated string literal (Line 1, Position 14)",
			stream: []token.Token{
				newToken(token.IDENT, "TEST"),
				tDeclare,
				newToken(token.ERROR, "SyntaxError: Unterminated string literal (Line 1, Position 14)"),
			},
		},
		{
			name:    "task bad char before body",
			message: "SyntaxError: Unexpected token '^' (Line 1, Position 12)",
			stream: []token.Token{
				tTask,
				newToken(token.IDENT, "test"),
				tLParen,
				tRParen,
				newToken(token.ERROR, "SyntaxError: Unexpected token '^' (Line 1, Position 12)"),
			},
		},
		{
			name:    "task bad char before body with comment",
			message: "SyntaxError: Unexpected token '^' (Line 1, Position 12)",
			stream: []token.Token{
				tHash,
				newToken(token.COMMENT, " Hello"),
				tTask,
				newToken(token.IDENT, "test"),
				tLParen,
				tRParen,
				newToken(token.ERROR, "SyntaxError: Unexpected token '^' (Line 1, Position 12)"),
			},
		},
		{
			name:    "task invalid chars end of body",
			message: "SyntaxError: Unexpected token 'U+000A' (Line 7, Position 52)",
			stream: []token.Token{
				tTask,
				newToken(token.IDENT, "test"),
				tLParen,
				tRParen,
				tLBrace,
				newToken(token.COMMAND, "go test ./..."),
				newToken(token.COMMAND, "go build ."),
				newToken(token.ERROR, "SyntaxError: Unexpected token 'U+000A' (Line 7, Position 52)"),
			},
		},
		{
			name:    "task unterminated body",
			message: "SyntaxError: Unterminated task body (Line 1, Position 13)",
			stream: []token.Token{
				tTask,
				newToken(token.IDENT, "test"),
				tLParen,
				tRParen,
				tLBrace,
				newToken(token.ERROR, "SyntaxError: Unterminated task body (Line 1, Position 13)"),
			},
		},
		{
			name:    "task invalid char args",
			message: "SyntaxError: Invalid character used in task dependency/output [2] (Line 1, Position 11). Only strings and declared variables may be used.",
			stream: []token.Token{
				tTask,
				newToken(token.IDENT, "test"),
				tLParen,
				newToken(token.ERROR, "SyntaxError: Invalid character used in task dependency/output [2] (Line 1, Position 11). Only strings and declared variables may be used."),
			},
		},
		{
			name:    "task no curlies",
			message: "Illegal Token: [EOF]  (Line 0). Expected '{'\n\n0 |\t",
			stream: []token.Token{
				tTask,
				newToken(token.IDENT, "test"),
				tLParen,
				newToken(token.STRING, `"file.go"`),
				tRParen,
				tEOF,
			},
		},
		{
			name:    "task missing output",
			message: "SyntaxError: Task declared dependency but none found (Line 1, Position 25)",
			stream: []token.Token{
				tTask,
				newToken(token.IDENT, "test"),
				tLParen,
				newToken(token.STRING, `"input.go"`),
				tRParen,
				tOutput,
				newToken(token.ERROR, "SyntaxError: Task declared dependency but none found (Line 1, Position 25)"),
			},
		},
		{
			name:    "task bad token after output",
			message: "SyntaxError: Unexpected token '^' (Line 1, Position 26)",
			stream: []token.Token{
				tTask,
				newToken(token.IDENT, "test"),
				tLParen,
				newToken(token.STRING, `"input.go"`),
				tRParen,
				tOutput,
				newToken(token.ERROR, "SyntaxError: Unexpected token '^' (Line 1, Position 26)"),
			},
		},
		{
			name:    "task missing closing output paren",
			message: "Illegal Token: \"{\" (Line 0). Expected one of ['STRING', 'IDENT', ',']\n\n0 |\t",
			stream: []token.Token{
				tTask,
				newToken(token.IDENT, "test"),
				tLParen,
				newToken(token.STRING, `"input.go"`),
				tRParen,
				tOutput,
				tLParen,
				newToken(token.STRING, `"output1.go`),
				tComma,
				newToken(token.STRING, `"output2.go"`),
				// Should be a closing RParen here, but there isn't
				tLBrace,
			},
		},
		{
			name: "parser expect error",
			stream: []token.Token{
				newToken(token.IDENT, "TEST"),
				// parseAssign will call expect on a ':=' here
				newToken(token.IDENT, "OOPS"),
			},
			message: "Illegal Token: [IDENT] OOPS (Line 0). Expected ':='\n\n0 |\t",
		},
		{
			name:    "parser unexpected top level token",
			stream:  []token.Token{newToken(token.STRING, `"Unexpected"`)},
			message: "Illegal Token: \"Unexpected\" (Line 0). Expected one of ['#', 'IDENT', 'task']\n\n0 |\t",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{
				lexer:  &testLexer{stream: tt.stream},
				buffer: [3]token.Token{},
			}

			_, err := p.Parse()
			if err == nil {
				t.Fatal("Expected error but got nil")
			}

			if err.Error() != tt.message {
				t.Errorf("Wrong error message: got %#v, wanted %#v", err.Error(), tt.message)
			}
		})
	}
}

// TestParseFullSpokfile tests the parser against a stream of tokens
// indicative of a fully populated, syntactically valid spokfile.
func TestParseFullSpokfile(t *testing.T) {
	t.Parallel()
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

	if diff := cmp.Diff(fullSpokfileAST, tree); diff != "" {
		t.Errorf("AST mismatch (-want +tree):\n%s", diff)
	}
}

// BenchmarkParseFullSpokfile determines the parser's performance on a stream
// of tokens indicative of a fully populated, syntactically valid spokfile.
func BenchmarkParseFullSpokfile(b *testing.B) {
	p := &Parser{
		lexer: &testLexer{
			stream: fullSpokfileStream,
		},
		buffer:    [3]token.Token{},
		peekCount: 0,
	}

	for b.Loop() {
		_, err := p.Parse()
		if err != nil {
			b.Fatalf("Parser produced an error: %v", err)
		}
	}
}

//
// INTEGRATION TESTS START HERE
//
// There be larger, integration or coupled lexer-parser tests below. If the real lexer is changed
// none of the above tests will break as they stub out the lexer for our testLexer
// the tests and benchmarks below make use of the real lexer and will break if that lexer breaks.
//
// The tests below will only run if SPOK_INTEGRATION_TEST is set, making it easy to run only isolated
// tests while developing to limit potentially distracting failing integration test output until ready.
//
// The benchmarks below will run when invoking go test with the -bench flag, there is no concept of unit
// or integration benchmarks here.
//

// A more or less complete, syntactically valid, spokfile with all the allowed constructs to act as
// an integration test and benchmark.
// Keep in sync with it's counterpart in lexer_test.go.
const fullSpokfile = `
# This is a top level comment

# This variable is presumably important later
GLOBAL := "very important stuff here"

GIT_COMMIT := exec("git rev-parse HEAD")

# Run the project unit tests
task test(fmt) {
	go test -race ./...
}

# Format the project source
task fmt("**/*.go") {
	go fmt ./...
}

# Do many things
task many() {
	line 1
	line 2
	line 3
	line 4
}

# Compile the project
task build("**/*.go") -> "./bin/main" {
	go build -ldflags="-X main.version=test -X main.commit=7cb0ec5609efb5fe0"
}

# Show the global variables
task show() {
	echo {{.GLOBAL}}
}

# Generate multiple outputs
task moar_things() -> ("output1.go", "output2.go") {
	do some stuff here
}

task no_comment() {
	some more stuff
}

# Generate output from a variable
task makedocs() -> DOCS {
	echo "making docs"
}

# Generate multiple outputs in variables
task makestuff() -> (DOCS, BUILD) {
	echo "doing things"
}
`

// fullSpokfileStream is the stream of tokens the lexer is confirmed to generate
// given the above fullSpokfile text as input.
// Keep in sync with the expected output in the lexer_text.go integration test.
var fullSpokfileStream = []token.Token{
	tHash,
	newToken(token.COMMENT, " This is a top level comment"),
	tHash,
	newToken(token.COMMENT, " This variable is presumably important later"),
	newToken(token.IDENT, "GLOBAL"),
	tDeclare,
	newToken(token.STRING, `"very important stuff here"`),
	newToken(token.IDENT, "GIT_COMMIT"),
	tDeclare,
	newToken(token.IDENT, "exec"),
	tLParen,
	newToken(token.STRING, `"git rev-parse HEAD"`),
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
	newToken(token.STRING, `"**/*.go"`),
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
	newToken(token.STRING, `"**/*.go"`),
	tRParen,
	tOutput,
	newToken(token.STRING, `"./bin/main"`),
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
	newToken(token.COMMAND, "echo {{.GLOBAL}}"),
	tRBrace,
	tHash,
	newToken(token.COMMENT, " Generate multiple outputs"),
	tTask,
	newToken(token.IDENT, "moar_things"),
	tLParen,
	tRParen,
	tOutput,
	tLParen,
	newToken(token.STRING, `"output1.go"`),
	tComma,
	newToken(token.STRING, `"output2.go"`),
	tRParen,
	tLBrace,
	newToken(token.COMMAND, "do some stuff here"),
	tRBrace,
	tTask,
	newToken(token.IDENT, "no_comment"),
	tLParen,
	tRParen,
	tLBrace,
	newToken(token.COMMAND, "some more stuff"),
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
	tComma,
	newToken(token.IDENT, "BUILD"),
	tRParen,
	tLBrace,
	newToken(token.COMMAND, `echo "doing things"`),
	tRBrace,
	tEOF,
}

// fullSpokfileAST is the expected abstract syntax tree that should be generated when
// parsing the fullSpokfileStream of tokens.
var fullSpokfileAST = ast.Tree{
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
					Command:  "echo {{.GLOBAL}}",
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
					Command:  "some more stuff",
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

// TestParserIntegration uses the real lexer to parse a fully
// populated spokfile as string input and therefore checks the whole
// parsing system end to end.
func TestParserIntegration(t *testing.T) {
	if os.Getenv("SPOK_INTEGRATION_TEST") == "" {
		t.Skip("Set SPOK_INTEGRATION_TEST to run this test.")
	}
	t.Parallel()
	p := New(fullSpokfile)
	tree, err := p.Parse()
	if err != nil {
		t.Fatalf("Parser error: %v", err)
	}

	if diff := cmp.Diff(fullSpokfileAST, tree); diff != "" {
		t.Errorf("AST mismatch (-want +tree):\n%s", diff)
	}
}

// TestParserErrorsIntegration uses the real lexer to test how the parser handles real live
// parse errors and whether we bring back the right context to the user for them to debug easily.
func TestParserErrorsIntegration(t *testing.T) {
	if os.Getenv("SPOK_INTEGRATION_TEST") == "" {
		t.Skip("Set SPOK_INTEGRATION_TEST to run this test.")
	}
	t.Parallel()

	tests := []struct {
		name  string
		input string
		err   string
	}{
		{
			name: "invalid char at end of task body",
			input: `# This is a task
			task test() {
				go test ./...
				go build .
				💥
			}`,
			err: "SyntaxError: Unexpected token 'ð' (Line 5). \n\n5 |\t💥",
		},
		{
			name:  "task no curlies",
			input: `task test("file.go")`,
			err:   "Illegal Token: [EOF]  (Line 1). Expected '{'\n\n1 |\ttask test(\"file.go\")",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(tt.input)
			_, err := p.Parse()
			if err == nil {
				t.Fatal("Expected a parser error but got none")
			}

			if err.Error() != tt.err {
				t.Errorf("Wrong error message.\ngot:\t%#v\nwanted:\t%#v", err.Error(), tt.err)
			}
		})
	}
}

func TestGetContext(t *testing.T) {
	if os.Getenv("SPOK_INTEGRATION_TEST") == "" {
		t.Skip("Set SPOK_INTEGRATION_TEST to run this test.")
	}
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
		token token.Token
	}{
		{
			name: "first line",
			input: `# Top level comment

			GLOBAL := "hello"

			# Run the tests
			task test("**/*.go") {
				go test ./...
			}
			`,
			want: `# Top level comment`,
			token: token.Token{
				Value: " Top level comment",
				Type:  token.COMMENT,
				Line:  1,
			},
		},
		{
			name: "variable declaration IDENT",
			input: `# Top level comment

			GLOBAL := "hello"

			# Run the tests
			task test("**/*.go") {
				go test ./...
			}
			`,
			want: `GLOBAL := "hello"`,
			token: token.Token{
				Value: ":=",
				Type:  token.DECLARE,
				Line:  3,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(tt.input)
			_, err := p.Parse()
			if err != nil {
				t.Fatalf("Parser returned an error: %v", err)
			}

			if got := p.getLine(tt.token); got != tt.want {
				t.Errorf("Wrong context. got %s, wanted %s", got, tt.want)
			}
		})
	}
}

// BenchmarkParserIntegration determines the performance of the parsing
// system as a whole, from raw string to AST.
func BenchmarkParserIntegration(b *testing.B) {
	p := New(fullSpokfile)

	for b.Loop() {
		_, err := p.Parse()
		if err != nil {
			b.Fatalf("Parser produced an error: %v", err)
		}
	}
}
