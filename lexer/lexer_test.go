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
	tEOF     = newToken(token.EOF, "")
	tHash    = newToken(token.HASH, "#")
	tDeclare = newToken(token.DECLARE, ":=")
	tTask    = newToken(token.TASK, "task")
	tLParen  = newToken(token.LPAREN, "(")
	tRParen  = newToken(token.RPAREN, ")")
	tLBrace  = newToken(token.LBRACE, "{")
	tRBrace  = newToken(token.RBRACE, "}")
	tOutput  = newToken(token.OUTPUT, "->")
)

var lexTests = []lexTest{
	{
		name:   "empty",
		input:  "",
		tokens: []token.Token{tEOF},
	},
	{
		name:   "bad input",
		input:  "*&^%",
		tokens: []token.Token{newToken(token.ERROR, "SyntaxError: Unexpected token '*' (Line 1, Position 0)")},
	},
	{
		name:   "hash",
		input:  "#",
		tokens: []token.Token{tHash, newToken(token.COMMENT, ""), tEOF},
	},
	{
		name:   "hash newline",
		input:  "#\n",
		tokens: []token.Token{tHash, newToken(token.COMMENT, ""), tEOF},
	},
	{
		name:   "comment",
		input:  "# A comment",
		tokens: []token.Token{tHash, newToken(token.COMMENT, " A comment"), tEOF},
	},
	{
		name:   "comment no space",
		input:  "#A comment",
		tokens: []token.Token{tHash, newToken(token.COMMENT, "A comment"), tEOF},
	},
	{
		name:   "comment newline",
		input:  "# A comment\n",
		tokens: []token.Token{tHash, newToken(token.COMMENT, " A comment"), tEOF},
	},
	{
		name:   "comment no space newline",
		input:  "#A comment\n",
		tokens: []token.Token{tHash, newToken(token.COMMENT, "A comment"), tEOF},
	},
	{
		name:   "whitespace",
		input:  "      \t\n\t\t\n\n\n   ",
		tokens: []token.Token{tEOF},
	},
	{
		name:   "global variable string",
		input:  `TEST := "hello"`,
		tokens: []token.Token{newToken(token.IDENT, "TEST"), tDeclare, newToken(token.STRING, `"hello"`), tEOF},
	},
	{
		name:  "global variable unterminated string",
		input: `TEST := "hello`,
		tokens: []token.Token{
			newToken(token.IDENT, "TEST"),
			tDeclare,
			newToken(token.ERROR, "SyntaxError: Unterminated string literal (Line 1, Position 14)"),
		},
	},
	{
		name:  "global variable bad ident",
		input: `TEST ^^ := "hello"`,
		tokens: []token.Token{
			newToken(token.IDENT, "TEST"),
			newToken(token.ERROR, "SyntaxError: Unexpected token '^' (Line 1, Position 5)"),
		},
	},
	{
		name:   "global variable whitespace",
		input:  `TEST := "  \t hello\t\t  "`,
		tokens: []token.Token{newToken(token.IDENT, "TEST"), tDeclare, newToken(token.STRING, `"  \t hello\t\t  "`), tEOF},
	},
	{
		name:   "global variable integer",
		input:  `TEST := 27`,
		tokens: []token.Token{newToken(token.IDENT, "TEST"), tDeclare, newToken(token.INTEGER, "27"), tEOF},
	},
	{
		name:  "global variable bad integer",
		input: `TEST := 272h`,
		tokens: []token.Token{
			newToken(token.IDENT, "TEST"),
			tDeclare,
			newToken(token.ERROR, "SyntaxError: Invalid integer literal (Line 1, Position 11)"),
		},
	},
	{
		name:  "global variable float",
		input: `TEST := 27.6`,
		tokens: []token.Token{
			newToken(token.IDENT, "TEST"),
			tDeclare,
			newToken(token.ERROR, "SyntaxError: Invalid integer literal (Line 1, Position 10)"),
		},
	},
	{
		name:   "global variable ident",
		input:  `TEST := VAR`,
		tokens: []token.Token{newToken(token.IDENT, "TEST"), tDeclare, newToken(token.IDENT, "VAR"), tEOF},
	},
	{
		name:  "global variable bad RHS",
		input: `TEST := *`,
		tokens: []token.Token{
			newToken(token.IDENT, "TEST"),
			tDeclare,
			newToken(token.ERROR, "SyntaxError: Unexpected token 'U+FFFFFFFFFFFFFFFF' (Line 1, Position 9)"),
		},
	},
	{
		name:  "basic task",
		input: `task test("file.go") { go test ./... }`,
		tokens: []token.Token{
			tTask,
			newToken(token.IDENT, "test"),
			tLParen,
			newToken(token.STRING, `"file.go"`),
			tRParen,
			tLBrace,
			newToken(token.COMMAND, "go test ./..."),
			tRBrace,
			tEOF,
		},
	},
	{
		name:  "task no args",
		input: `task test() { go test ./... }`,
		tokens: []token.Token{
			tTask,
			newToken(token.IDENT, "test"),
			tLParen,
			tRParen,
			tLBrace,
			newToken(token.COMMAND, "go test ./..."),
			tRBrace,
			tEOF,
		},
	},
	{
		name:  "task quotes in body",
		input: `task test() { echo "hello" }`,
		tokens: []token.Token{
			tTask,
			newToken(token.IDENT, "test"),
			tLParen,
			tRParen,
			tLBrace,
			newToken(token.COMMAND, `echo "hello"`),
			tRBrace,
			tEOF,
		},
	},
	{
		name: "multi line task",
		input: `task test("file.go") {
			go test ./...
			go build .
			some command go
		}`,
		tokens: []token.Token{
			tTask,
			newToken(token.IDENT, "test"),
			tLParen,
			newToken(token.STRING, `"file.go"`),
			tRParen,
			tLBrace,
			newToken(token.COMMAND, "go test ./..."),
			newToken(token.COMMAND, "go build ."),
			newToken(token.COMMAND, "some command go"),
			tRBrace,
			tEOF,
		},
	},
	{
		name: "multi line task with quotes",
		input: `task test("file.go") {
			go test ./...
			go build .
			echo "hello"
		}`,
		tokens: []token.Token{
			tTask,
			newToken(token.IDENT, "test"),
			tLParen,
			newToken(token.STRING, `"file.go"`),
			tRParen,
			tLBrace,
			newToken(token.COMMAND, "go test ./..."),
			newToken(token.COMMAND, "go build ."),
			newToken(token.COMMAND, `echo "hello"`),
			tRBrace,
			tEOF,
		},
	},
	{
		name:  "task empty body",
		input: `task test() {}`,
		tokens: []token.Token{
			tTask,
			newToken(token.IDENT, "test"),
			tLParen,
			tRParen,
			tLBrace,
			tRBrace,
			tEOF,
		},
	},
	{
		name:  "task invalid chars in body",
		input: `task test() { ^% }`,
		tokens: []token.Token{
			tTask,
			newToken(token.IDENT, "test"),
			tLParen,
			tRParen,
			tLBrace,
			newToken(token.ERROR, "SyntaxError: Unexpected token '%!'(MISSING) (Line 1, Position 15)"),
		},
	},
	{
		name: "task invalid chars end of body",
		input: `task test() {
			go test ./...
			go build .
			💥
		}`,
		tokens: []token.Token{
			tTask,
			newToken(token.IDENT, "test"),
			tLParen,
			tRParen,
			tLBrace,
			newToken(token.COMMAND, "go test ./..."),
			newToken(token.COMMAND, "go build ."),
			newToken(token.ERROR, `SyntaxError: Unexpected token 'U+000A' (Line 7, Position 52)`),
		},
	},
	{
		name:  "task unterminated body",
		input: `task test() {`,
		tokens: []token.Token{
			tTask,
			newToken(token.IDENT, "test"),
			tLParen,
			tRParen,
			tLBrace,
			newToken(token.ERROR, "SyntaxError: Unterminated task body (Line 1, Position 13)"),
		},
	},
	{
		name:  "task whitespace body",
		input: "task test() {  \t\t \n\n  \t  }",
		tokens: []token.Token{
			tTask,
			newToken(token.IDENT, "test"),
			tLParen,
			tRParen,
			tLBrace,
			tRBrace,
			tEOF,
		},
	},
	{
		name:  "task whitespace args",
		input: "task test(  \t\t \n\n \t  ) {}",
		tokens: []token.Token{
			tTask,
			newToken(token.IDENT, "test"),
			tLParen,
			tRParen,
			tLBrace,
			tRBrace,
			tEOF,
		},
	},
	{
		name:  "task multiple string args",
		input: `task test("file1.go", "file2.go") {}`,
		tokens: []token.Token{
			tTask,
			newToken(token.IDENT, "test"),
			tLParen,
			newToken(token.STRING, `"file1.go"`),
			newToken(token.STRING, `"file2.go"`),
			tRParen,
			tLBrace,
			tRBrace,
			tEOF,
		},
	},
	{
		name:  "task glob pattern arg",
		input: `task test("**/*.go") {}`,
		tokens: []token.Token{
			tTask,
			newToken(token.IDENT, "test"),
			tLParen,
			newToken(token.STRING, `"**/*.go"`),
			tRParen,
			tLBrace,
			tRBrace,
			tEOF,
		},
	},
	{
		name:  "task ident arg",
		input: `task test(some_variable) {}`,
		tokens: []token.Token{
			tTask,
			newToken(token.IDENT, "test"),
			tLParen,
			newToken(token.IDENT, "some_variable"),
			tRParen,
			tLBrace,
			tRBrace,
			tEOF,
		},
	},
	{
		name:  "task string and ident arg",
		input: `task test("file1.go", fmt) {}`,
		tokens: []token.Token{
			tTask,
			newToken(token.IDENT, "test"),
			tLParen,
			newToken(token.STRING, `"file1.go"`),
			newToken(token.IDENT, "fmt"),
			tRParen,
			tLBrace,
			tRBrace,
			tEOF,
		},
	},
	{
		name:  "task ident and string arg",
		input: `task test(fmt, "file1.go") {}`,
		tokens: []token.Token{
			tTask,
			newToken(token.IDENT, "test"),
			tLParen,
			newToken(token.IDENT, "fmt"),
			newToken(token.STRING, `"file1.go"`),
			tRParen,
			tLBrace,
			tRBrace,
			tEOF,
		},
	},
	{
		name:  "task many args",
		input: `task test(lint, fmt, "**/*.go", build, dave, "hello") {}`,
		tokens: []token.Token{
			tTask,
			newToken(token.IDENT, "test"),
			tLParen,
			newToken(token.IDENT, "lint"),
			newToken(token.IDENT, "fmt"),
			newToken(token.STRING, `"**/*.go"`),
			newToken(token.IDENT, "build"),
			newToken(token.IDENT, "dave"),
			newToken(token.STRING, `"hello"`),
			tRParen,
			tLBrace,
			tRBrace,
			tEOF,
		},
	},
	{
		name:  "task invalid char args",
		input: `task test(625) {}`,
		tokens: []token.Token{
			tTask,
			newToken(token.IDENT, "test"),
			tLParen,
			newToken(token.ERROR, "SyntaxError: Invalid character used in task dependency [2] (Line 1, Position 11). Only strings and declared variables may be used."),
		},
	},
	{
		name:  "task no curlies",
		input: `task test("file.go")`,
		tokens: []token.Token{
			tTask,
			newToken(token.IDENT, "test"),
			tLParen,
			newToken(token.STRING, `"file.go"`),
			tRParen,
			newToken(token.ERROR, "SyntaxError: Task has no body (Line 1, Position 20)"),
		},
	},
	{
		name:  "task with interpolation",
		input: `task test("file.go") { echo GLOBAL_VARIABLE }`,
		tokens: []token.Token{
			tTask,
			newToken(token.IDENT, "test"),
			tLParen,
			newToken(token.STRING, `"file.go"`),
			tRParen,
			tLBrace,
			newToken(token.COMMAND, "echo GLOBAL_VARIABLE"),
			tRBrace,
			tEOF,
		},
	},
	{
		name:  "task with single output",
		input: `task test("input.go") -> "output.go" { go build input.go }`,
		tokens: []token.Token{
			tTask,
			newToken(token.IDENT, "test"),
			tLParen,
			newToken(token.STRING, `"input.go"`),
			tRParen,
			tOutput,
			newToken(token.STRING, `"output.go"`),
			tLBrace,
			newToken(token.COMMAND, "go build input.go"),
			tRBrace,
			tEOF,
		},
	},
}

// collect gathers the emitted tokens into a slice for comparison.
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

// equal compares to slices of tokens for equality.
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
				t.Errorf("%s: got\n\t%+v\nexpected\n\t%+v", test.name, tokens, test.tokens)
			}
		})
	}
}
