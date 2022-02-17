package lexer

import (
	"os"
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
	tComma   = newToken(token.COMMA, ",")
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
		tokens: []token.Token{newToken(token.ERROR, "SyntaxError: Unexpected token '*' (Line 1). \n\t\t\n1 |\t*&^%")},
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
			newToken(token.ERROR, "SyntaxError: String literal missing closing quote: \"hell (Line 1). \n\t\t\n1 |\tTEST := \"hello"),
		},
	},
	{
		name:  "global variable bad ident",
		input: `TEST ^^ := "hello"`,
		tokens: []token.Token{
			newToken(token.IDENT, "TEST"),
			newToken(token.ERROR, "SyntaxError: Unexpected token '^' (Line 1). \n\t\t\n1 |\tTEST ^^ := \"hello\""),
		},
	},
	{
		name:   "global variable whitespace",
		input:  `TEST := "  \t hello\t\t  "`,
		tokens: []token.Token{newToken(token.IDENT, "TEST"), tDeclare, newToken(token.STRING, `"  \t hello\t\t  "`), tEOF},
	},
	{
		name:  "global variable integer",
		input: `TEST := 27`,
		tokens: []token.Token{
			newToken(token.IDENT, "TEST"),
			tDeclare,
			newToken(token.ERROR, "SyntaxError: Unexpected token '2' (Line 1). \n\t\t\n1 |\tTEST := 27"),
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
			newToken(token.ERROR, "SyntaxError: Unexpected token '*' (Line 1). \n\t\t\n1 |\tTEST := *"),
		},
	},
	{
		name:  "global variable function RHS",
		input: `TEST := exec("git rev-parse HEAD")`,
		tokens: []token.Token{
			newToken(token.IDENT, "TEST"),
			tDeclare,
			newToken(token.IDENT, "exec"),
			tLParen,
			newToken(token.STRING, `"git rev-parse HEAD"`),
			tRParen,
			tEOF,
		},
	},
	{
		name:  "global variable join RHS",
		input: `TEST := join(ROOT, "docs", "build")`,
		tokens: []token.Token{
			newToken(token.IDENT, "TEST"),
			tDeclare,
			newToken(token.IDENT, "join"),
			tLParen,
			newToken(token.IDENT, "ROOT"),
			tComma,
			newToken(token.STRING, `"docs"`),
			tComma,
			newToken(token.STRING, `"build"`),
			tRParen,
			tEOF,
		},
	},
	{
		name:  "global variable join RHS trailing comma",
		input: `TEST := join(ROOT, "docs", "build",)`,
		tokens: []token.Token{
			newToken(token.IDENT, "TEST"),
			tDeclare,
			newToken(token.IDENT, "join"),
			tLParen,
			newToken(token.IDENT, "ROOT"),
			tComma,
			newToken(token.STRING, `"docs"`),
			tComma,
			newToken(token.STRING, `"build"`),
			tComma,
			tRParen,
			tEOF,
		},
	},
	{
		name:  "global variable join RHS no commas",
		input: `TEST := join(ROOT "docs" "build")`,
		tokens: []token.Token{
			newToken(token.IDENT, "TEST"),
			tDeclare,
			newToken(token.IDENT, "join"),
			tLParen,
			newToken(token.IDENT, "ROOT"),
			newToken(token.ERROR, "SyntaxError: Unexpected token '\"' (Line 1). \n\t\t\n1 |\tTEST := join(ROOT \"docs\" \"build\")"),
		},
	},
	{
		name:  "global variable join RHS illegal token",
		input: `TEST := join(ROOT, "docs", "build", #)`,
		tokens: []token.Token{
			newToken(token.IDENT, "TEST"),
			tDeclare,
			newToken(token.IDENT, "join"),
			tLParen,
			newToken(token.IDENT, "ROOT"),
			tComma,
			newToken(token.STRING, `"docs"`),
			tComma,
			newToken(token.STRING, `"build"`),
			tComma,
			newToken(token.ERROR, "SyntaxError: Unexpected token '#' (Line 1). \n\t\t\n1 |\tTEST := join(ROOT, \"docs\", \"build\", #)"),
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
		name:  "task bad char before body",
		input: `task test() ^ {}`,
		tokens: []token.Token{
			tTask,
			newToken(token.IDENT, "test"),
			tLParen,
			tRParen,
			newToken(token.ERROR, "SyntaxError: Unexpected token '^' (Line 1). \n\t\t\n1 |\ttask test() ^ {}"),
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
			newToken(token.ERROR, "SyntaxError: Unexpected token '%' (Line 1). \n\t\t\n1 |\ttask test() { ^% }"),
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
			newToken(token.ERROR, "SyntaxError: Unexpected token 'ð' (Line 4). \n\t\t\n4 |\t💥"),
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
			newToken(token.ERROR, "SyntaxError: Unterminated task body (Line 1). \n\t\t\n1 |\ttask test() {"),
		},
	},
	{
		name: "task unterminated body after commands",
		input: `task test() {
			go test ./...
		`,
		tokens: []token.Token{
			tTask,
			newToken(token.IDENT, "test"),
			tLParen,
			tRParen,
			tLBrace,
			newToken(token.COMMAND, "go test ./..."),
			newToken(token.ERROR, "SyntaxError: Unterminated task body (Line 3). \n\t\t\n3 |\t"),
		},
	},
	// TODO: This still fails, need to think about how to address this
	// {
	// 	name: "task unterminated body with another task below",
	// 	input: `task test() {
	// 		go test ./...

	// 	# I'm a comment
	// 	task build() {
	// 		go build ./...
	// 	}
	// 	`,
	// 	tokens: []token.Token{
	// 		tTask,
	// 		newToken(token.IDENT, "test"),
	// 		tLParen,
	// 		tRParen,
	// 		tLBrace,
	// 		newToken(token.COMMAND, "go test ./..."),
	// 		newToken(token.ERROR, "SyntaxError: Unterminated task body (Line 3)"),
	// 	},
	// },
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
			tComma,
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
			tComma,
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
			tComma,
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
			tComma,
			newToken(token.IDENT, "fmt"),
			tComma,
			newToken(token.STRING, `"**/*.go"`),
			tComma,
			newToken(token.IDENT, "build"),
			tComma,
			newToken(token.IDENT, "dave"),
			tComma,
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
			newToken(token.ERROR, "SyntaxError: Invalid character used in task dependency/output (Line 1). \n\t\t\n1 |\ttask test(625) {}"),
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
			tEOF, // Can't do syntax error here because could be a global variable function call, handled in parser
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
	{
		name:  "task with glob output",
		input: `task test("**/*.md") -> "**/*.html" { buildy docs }`,
		tokens: []token.Token{
			tTask,
			newToken(token.IDENT, "test"),
			tLParen,
			newToken(token.STRING, `"**/*.md"`),
			tRParen,
			tOutput,
			newToken(token.STRING, `"**/*.html"`),
			tLBrace,
			newToken(token.COMMAND, "buildy docs"),
			tRBrace,
			tEOF,
		},
	},
	{
		name:  "task with ident output",
		input: `task test("**/*.md") -> BUILD { buildy docs }`,
		tokens: []token.Token{
			tTask,
			newToken(token.IDENT, "test"),
			tLParen,
			newToken(token.STRING, `"**/*.md"`),
			tRParen,
			tOutput,
			newToken(token.IDENT, "BUILD"),
			tLBrace,
			newToken(token.COMMAND, "buildy docs"),
			tRBrace,
			tEOF,
		},
	},
	{
		name:  "task with multi ident output",
		input: `task test("**/*.md") -> (BUILD, SOMETHING) { buildy docs }`,
		tokens: []token.Token{
			tTask,
			newToken(token.IDENT, "test"),
			tLParen,
			newToken(token.STRING, `"**/*.md"`),
			tRParen,
			tOutput,
			tLParen,
			newToken(token.IDENT, "BUILD"),
			tComma,
			newToken(token.IDENT, "SOMETHING"),
			tRParen,
			tLBrace,
			newToken(token.COMMAND, "buildy docs"),
			tRBrace,
			tEOF,
		},
	},
	{
		name:  "task with multi output",
		input: `task test("input.go") -> ("output1.go", "output2.go") { go build input.go }`,
		tokens: []token.Token{
			tTask,
			newToken(token.IDENT, "test"),
			tLParen,
			newToken(token.STRING, `"input.go"`),
			tRParen,
			tOutput,
			tLParen,
			newToken(token.STRING, `"output1.go"`),
			tComma,
			newToken(token.STRING, `"output2.go"`),
			tRParen,
			tLBrace,
			newToken(token.COMMAND, "go build input.go"),
			tRBrace,
			tEOF,
		},
	},
	{
		name: "task multi output multi line",
		input: `task test("input.go") -> ("output1.go", "output2.go") {
			go build input.go
			do something else
			echo "hello"
		}`,
		tokens: []token.Token{
			tTask,
			newToken(token.IDENT, "test"),
			tLParen,
			newToken(token.STRING, `"input.go"`),
			tRParen,
			tOutput,
			tLParen,
			newToken(token.STRING, `"output1.go"`),
			tComma,
			newToken(token.STRING, `"output2.go"`),
			tRParen,
			tLBrace,
			newToken(token.COMMAND, "go build input.go"),
			newToken(token.COMMAND, "do something else"),
			newToken(token.COMMAND, `echo "hello"`),
			tRBrace,
			tEOF,
		},
	},
	{
		name: "task missing output",
		input: `task test("input.go") -> {
			go build input.go
			do something else
			echo "hello"
		}`,
		tokens: []token.Token{
			tTask,
			newToken(token.IDENT, "test"),
			tLParen,
			newToken(token.STRING, `"input.go"`),
			tRParen,
			tOutput,
			newToken(token.ERROR, "SyntaxError: Task declared dependency but none found (Line 1). \n\t\t\n1 |\ttask test(\"input.go\") -> {"),
		},
	},
	{
		name: "task bad token after output",
		input: `task test("input.go") -> ^^{
			go build input.go
			do something else
			echo "hello"
		}`,
		tokens: []token.Token{
			tTask,
			newToken(token.IDENT, "test"),
			tLParen,
			newToken(token.STRING, `"input.go"`),
			tRParen,
			tOutput,
			newToken(token.ERROR, "SyntaxError: Unexpected token '^' (Line 1). \n\t\t\n1 |\ttask test(\"input.go\") -> ^^{"),
		},
	},
	{
		name: "outputs across lines",
		input: `task test("input.go") -> (
			"output1.go",
			"output2.go") {
			go build input.go
			do something else
			echo "hello"
		}`,
		tokens: []token.Token{
			tTask,
			newToken(token.IDENT, "test"),
			tLParen,
			newToken(token.STRING, `"input.go"`),
			tRParen,
			tOutput,
			tLParen,
			newToken(token.STRING, `"output1.go"`),
			tComma,
			newToken(token.STRING, `"output2.go"`),
			tRParen,
			tLBrace,
			newToken(token.COMMAND, "go build input.go"),
			newToken(token.COMMAND, "do something else"),
			newToken(token.COMMAND, `echo "hello"`),
			tRBrace,
			tEOF,
		},
	},
}

// collect gathers the emitted tokens into a slice for comparison.
func collect(t *lexTest) (tokens []token.Token) {
	l := New(t.input)
	for {
		tok := l.NextToken()
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
			test := test
			tokens := collect(&test)
			if !equal(tokens, test.tokens) {
				t.Errorf("%s: got\n\t%+v\nexpected\n\t%+v", test.name, tokens, test.tokens)
			}
		})
	}
}

//
// INTEGRATION TESTS START HERE
//
// Thar be larger, integration tests below.

// The tests below will only run if SPOK_INTEGRATION_TEST is set, making it easy to run only isolated
// tests while developing to limit potentially distracting failing integration test output until ready.
//
// The benchmarks below will run when invoking go test with the -bench flag, there is no concept of unit
// or integration benchmarks here.
//

// A more or less complete spokfile with all the allowed constructs to act as
// an integration test and benchmark.
var fullSpokfile = `# This is a top level comment

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
	echo GLOBAL
}

# Generate multiple outputs
task moar_things() -> ("output1.go", "output2.go") {
	do some stuff here
}

task no_comment() {
	echo "this task has no docstring"
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
	tComma,
	newToken(token.IDENT, "BUILD"),
	tRParen,
	tLBrace,
	newToken(token.COMMAND, `echo "doing things"`),
	tRBrace,
	tEOF,
}

// TestLexerIntegration tests the lexer against a fully populated, syntactically valid spokfile.
func TestLexerIntegration(t *testing.T) {
	if os.Getenv("SPOK_INTEGRATION_TEST") == "" {
		t.Skip("Set SPOK_INTEGRATION_TEST to run this test.")
	}

	l := New(fullSpokfile)
	var tokens []token.Token
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == token.EOF || tok.Type == token.ERROR {
			break
		}
	}

	if !equal(tokens, fullSpokfileStream) {
		t.Errorf("got\n\t%+v\nexpected\n\t%+v", tokens, fullSpokfile)
	}
}

// BenchmarkLexFullSpokfile determines the performance of lexing the integration spokfile above.
func BenchmarkLexFullSpokfile(b *testing.B) {
	l := New(fullSpokfile)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for {
			tok := l.NextToken()
			if tok.Type == token.EOF || tok.Type == token.ERROR {
				break
			}
		}
	}
}
