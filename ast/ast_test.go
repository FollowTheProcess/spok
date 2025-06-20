package ast_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.followtheprocess.codes/spok/ast"
)

func TestType(t *testing.T) {
	t.Parallel()
	tests := []struct {
		node ast.Node
		name string
		want ast.NodeType
	}{
		{
			name: "comment",
			node: ast.Comment{NodeType: ast.NodeComment},
			want: ast.NodeComment,
		},
		{
			name: "ident",
			node: ast.Comment{NodeType: ast.NodeIdent},
			want: ast.NodeIdent,
		},
		{
			name: "assign",
			node: ast.Assign{NodeType: ast.NodeAssign},
			want: ast.NodeAssign,
		},
		{
			name: "string",
			node: ast.String{NodeType: ast.NodeString},
			want: ast.NodeString,
		},
		{
			name: "function",
			node: ast.Function{NodeType: ast.NodeFunction},
			want: ast.NodeFunction,
		},
		{
			name: "task",
			node: ast.Task{NodeType: ast.NodeTask},
			want: ast.NodeTask,
		},
		{
			name: "command",
			node: ast.Command{NodeType: ast.NodeCommand},
			want: ast.NodeCommand,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.node.Type(); got != tt.want {
				t.Errorf("Wrong node type.\nGot %s\nWant %s", got, tt.want)
			}
		})
	}
}

func TestLiteral(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		node ast.Node
		want string
	}{
		{
			name: "comment",
			node: ast.Comment{Text: "I'm a comment", NodeType: ast.NodeComment},
			want: "# I'm a comment\n",
		},
		{
			name: "ident",
			node: ast.Ident{Name: "ident", NodeType: ast.NodeIdent},
			want: "ident",
		},
		{
			name: "assign string",
			node: ast.Assign{
				Value:    ast.String{Text: "Hello"},
				Name:     ast.Ident{Name: "VALUE"},
				NodeType: ast.NodeAssign,
			},
			want: "VALUE := \"Hello\"\n",
		},
		{
			name: "assign builtin",
			node: ast.Assign{
				Value: ast.Function{
					Name:      ast.Ident{Name: "exec"},
					Arguments: []ast.Node{ast.String{Text: "true"}},
				},
				Name:     ast.Ident{Name: "VALUE"},
				NodeType: ast.NodeAssign,
			},
			want: "VALUE := exec(\"true\")\n",
		},
		{
			name: "string",
			node: ast.String{Text: "hello", NodeType: ast.NodeString},
			want: "hello",
		},
		{
			name: "function",
			node: ast.Function{
				Name:      ast.Ident{Name: "join"},
				Arguments: []ast.Node{ast.String{Text: "dir"}, ast.String{Text: "another"}},
				NodeType:  ast.NodeFunction,
			},
			want: "join(\"dir\", \"another\")",
		},
		{
			name: "task",
			node: ast.Task{
				Name:         ast.Ident{Name: "test"},
				Docstring:    ast.Comment{Text: "I'm a test task"},
				Dependencies: []ast.Node{ast.String{Text: "**/*.go"}},
				Outputs:      []ast.Node{ast.String{Text: "./bin/main"}},
				Commands: []ast.Command{
					{Command: "go test ./..."},
				},
				NodeType: ast.NodeTask,
			},
			want: `# I'm a test task
task test("**/*.go") -> "./bin/main" {
    go test ./...
}

`,
		},
		{
			name: "command",
			node: ast.Command{Command: "git commit", NodeType: ast.NodeCommand},
			want: "git commit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.node.Literal(); got != tt.want {
				t.Errorf("%s wrong Literal()\nGot %q\nWanted %q", tt.node.Type(), got, tt.want)
			}
		})
	}
}

func TestWrite(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		node ast.Node
		want string
	}{
		{
			name: "comment",
			node: ast.Comment{Text: "I'm a comment", NodeType: ast.NodeComment},
			want: "# I'm a comment\n",
		},
		{
			name: "ident",
			node: ast.Ident{Name: "ident", NodeType: ast.NodeIdent},
			want: "ident",
		},
		{
			name: "assign string",
			node: ast.Assign{
				Value:    ast.String{Text: "Hello"},
				Name:     ast.Ident{Name: "VALUE"},
				NodeType: ast.NodeAssign,
			},
			want: "VALUE := \"Hello\"\n",
		},
		{
			name: "assign builtin",
			node: ast.Assign{
				Value: ast.Function{
					Name:      ast.Ident{Name: "exec"},
					Arguments: []ast.Node{ast.String{Text: "true"}},
				},
				Name:     ast.Ident{Name: "VALUE"},
				NodeType: ast.NodeAssign,
			},
			want: "VALUE := exec(\"true\")\n",
		},
		{
			name: "string",
			node: ast.String{Text: "hello", NodeType: ast.NodeString},
			want: `"hello"`,
		},
		{
			name: "function",
			node: ast.Function{
				Name:      ast.Ident{Name: "join"},
				Arguments: []ast.Node{ast.String{Text: "dir"}, ast.String{Text: "another"}},
				NodeType:  ast.NodeFunction,
			},
			want: "join(\"dir\", \"another\")",
		},
		{
			name: "task",
			node: ast.Task{
				Name:         ast.Ident{Name: "test"},
				Docstring:    ast.Comment{Text: "I'm a test task"},
				Dependencies: []ast.Node{ast.String{Text: "**/*.go"}},
				Outputs:      []ast.Node{ast.String{Text: "./bin/main"}},
				Commands: []ast.Command{
					{Command: "go test ./..."},
				},
				NodeType: ast.NodeTask,
			},
			want: `# I'm a test task
task test("**/*.go") -> "./bin/main" {
    go test ./...
}

`,
		},
		{
			name: "command",
			node: ast.Command{Command: "git commit", NodeType: ast.NodeCommand},
			want: "git commit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := &strings.Builder{}
			tt.node.Write(builder)
			if got := builder.String(); got != tt.want {
				t.Errorf("%s\nGot %s\nWanted %s", tt.node.Type(), got, tt.want)
			}
		})
	}
}

func TestAppend(t *testing.T) {
	t.Parallel()
	tree := ast.Tree{
		Nodes: []ast.Node{
			ast.Comment{
				Text:     " I'm a comment",
				NodeType: ast.NodeComment,
			},
		},
	}

	tree.Append(ast.Ident{Name: "GLOBAL", NodeType: ast.NodeIdent})

	want := ast.Tree{
		Nodes: []ast.Node{
			ast.Comment{
				Text:     " I'm a comment",
				NodeType: ast.NodeComment,
			},
			ast.Ident{
				Name:     "GLOBAL",
				NodeType: ast.NodeIdent,
			},
		},
	}

	if diff := cmp.Diff(want, tree); diff != "" {
		t.Errorf("AST mismatch (-want +tree):\n%s", diff)
	}
}

func TestIsEmpty(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		tree  ast.Tree
		empty bool
	}{
		{
			name:  "empty",
			tree:  ast.Tree{},
			empty: true,
		},
		{
			name: "not empty",
			tree: ast.Tree{
				Nodes: []ast.Node{
					ast.Comment{
						Text:     " Hello",
						NodeType: ast.NodeComment,
					},
				},
			},
			empty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.tree.IsEmpty() != tt.empty {
				t.Errorf("got %v, wanted %v", tt.tree.IsEmpty(), tt.empty)
			}
		})
	}
}

const fullSpokfile = `# This is a top level comment
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

func TestWriteWholeTree(t *testing.T) {
	t.Parallel()
	if diff := cmp.Diff(fullSpokfile, fullSpokfileAST.String()); diff != "" {
		t.Errorf("AST mismatch (-want +got):\n%s", diff)
	}
}

func BenchmarkWriteWholeTree(b *testing.B) {
	for b.Loop() {
		_ = fullSpokfileAST.String()
	}
}
