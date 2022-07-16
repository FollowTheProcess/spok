package ast

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestType(t *testing.T) {
	tests := []struct {
		node Node
		name string
		want NodeType
	}{
		{
			name: "comment",
			node: Comment{NodeType: NodeComment},
			want: NodeComment,
		},
		{
			name: "ident",
			node: Comment{NodeType: NodeIdent},
			want: NodeIdent,
		},
		{
			name: "assign",
			node: Assign{NodeType: NodeAssign},
			want: NodeAssign,
		},
		{
			name: "string",
			node: String{NodeType: NodeString},
			want: NodeString,
		},
		{
			name: "function",
			node: Function{NodeType: NodeFunction},
			want: NodeFunction,
		},
		{
			name: "task",
			node: Task{NodeType: NodeTask},
			want: NodeTask,
		},
		{
			name: "command",
			node: Command{NodeType: NodeCommand},
			want: NodeCommand,
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
		node Node
		want string
	}{
		{
			name: "comment",
			node: Comment{Text: "I'm a comment", NodeType: NodeComment},
			want: "# I'm a comment\n",
		},
		{
			name: "ident",
			node: Ident{Name: "ident", NodeType: NodeIdent},
			want: "ident",
		},
		{
			name: "assign string",
			node: Assign{
				Value:    String{Text: "Hello"},
				Name:     Ident{Name: "VALUE"},
				NodeType: NodeAssign,
			},
			want: "VALUE := \"Hello\"\n",
		},
		{
			name: "assign builtin",
			node: Assign{
				Value: Function{
					Name:      Ident{Name: "exec"},
					Arguments: []Node{String{Text: "true"}},
				},
				Name:     Ident{Name: "VALUE"},
				NodeType: NodeAssign,
			},
			want: "VALUE := exec(\"true\")\n\n",
		},
		{
			name: "string",
			node: String{Text: "hello", NodeType: NodeString},
			want: "hello",
		},
		{
			name: "function",
			node: Function{
				Name:      Ident{Name: "join"},
				Arguments: []Node{String{Text: "dir"}, String{Text: "another"}},
				NodeType:  NodeFunction,
			},
			want: "join(\"dir\", \"another\")\n",
		},
		{
			name: "task",
			node: Task{
				Name:         Ident{Name: "test"},
				Docstring:    Comment{Text: "I'm a test task"},
				Dependencies: []Node{String{Text: "**/*.go"}},
				Outputs:      []Node{String{Text: "./bin/main"}},
				Commands: []Command{
					{Command: "go test ./..."},
				},
				NodeType: NodeTask,
			},
			want: `# I'm a test task
task test("**/*.go") -> "./bin/main" {
    go test ./...
}

`,
		},
		{
			name: "command",
			node: Command{Command: "git commit", NodeType: NodeCommand},
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
		node Node
		want string
	}{
		{
			name: "comment",
			node: Comment{Text: "I'm a comment", NodeType: NodeComment},
			want: "# I'm a comment\n",
		},
		{
			name: "ident",
			node: Ident{Name: "ident", NodeType: NodeIdent},
			want: "ident",
		},
		{
			name: "assign string",
			node: Assign{
				Value:    String{Text: "Hello"},
				Name:     Ident{Name: "VALUE"},
				NodeType: NodeAssign,
			},
			want: "VALUE := \"Hello\"\n",
		},
		{
			name: "assign builtin",
			node: Assign{
				Value: Function{
					Name:      Ident{Name: "exec"},
					Arguments: []Node{String{Text: "true"}},
				},
				Name:     Ident{Name: "VALUE"},
				NodeType: NodeAssign,
			},
			want: "VALUE := exec(\"true\")\n\n",
		},
		{
			name: "string",
			node: String{Text: "hello", NodeType: NodeString},
			want: `"hello"`,
		},
		{
			name: "function",
			node: Function{
				Name:      Ident{Name: "join"},
				Arguments: []Node{String{Text: "dir"}, String{Text: "another"}},
				NodeType:  NodeFunction,
			},
			want: "join(\"dir\", \"another\")\n",
		},
		{
			name: "task",
			node: Task{
				Name:         Ident{Name: "test"},
				Docstring:    Comment{Text: "I'm a test task"},
				Dependencies: []Node{String{Text: "**/*.go"}},
				Outputs:      []Node{String{Text: "./bin/main"}},
				Commands: []Command{
					{Command: "go test ./..."},
				},
				NodeType: NodeTask,
			},
			want: `# I'm a test task
task test("**/*.go") -> "./bin/main" {
    go test ./...
}

`,
		},
		{
			name: "command",
			node: Command{Command: "git commit", NodeType: NodeCommand},
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
	tree := Tree{
		Nodes: []Node{
			Comment{
				Text:     " I'm a comment",
				NodeType: NodeComment,
			},
		},
	}

	tree.Append(Ident{Name: "GLOBAL", NodeType: NodeIdent})

	want := Tree{
		Nodes: []Node{
			Comment{
				Text:     " I'm a comment",
				NodeType: NodeComment,
			},
			Ident{
				Name:     "GLOBAL",
				NodeType: NodeIdent,
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
		tree  Tree
		empty bool
	}{
		{
			name:  "empty",
			tree:  Tree{},
			empty: true,
		},
		{
			name: "not empty",
			tree: Tree{
				Nodes: []Node{
					Comment{
						Text:     " Hello",
						NodeType: NodeComment,
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

var fullSpokfileAST = Tree{
	Nodes: []Node{
		Comment{
			Text:     " This is a top level comment",
			NodeType: NodeComment,
		},
		Comment{
			Text:     " This variable is presumably important later",
			NodeType: NodeComment,
		},
		Assign{
			Name: Ident{
				Name:     "GLOBAL",
				NodeType: NodeIdent,
			},
			Value: String{
				Text:     "very important stuff here",
				NodeType: NodeString,
			},
			NodeType: NodeAssign,
		},
		Assign{
			Value: Function{
				Name: Ident{
					Name:     "exec",
					NodeType: NodeIdent,
				},
				Arguments: []Node{
					String{
						Text:     "git rev-parse HEAD",
						NodeType: NodeString,
					},
				}, NodeType: NodeFunction,
			},
			Name: Ident{
				Name:     "GIT_COMMIT",
				NodeType: NodeIdent,
			},
			NodeType: NodeAssign,
		},
		Task{
			Name: Ident{
				Name:     "test",
				NodeType: NodeIdent,
			},
			Docstring: Comment{
				Text:     " Run the project unit tests",
				NodeType: NodeComment,
			},
			Dependencies: []Node{
				Ident{
					Name:     "fmt",
					NodeType: NodeIdent,
				},
			},
			Outputs: []Node{},
			Commands: []Command{
				{
					Command:  "go test -race ./...",
					NodeType: NodeCommand,
				},
			},
			NodeType: NodeTask,
		},
		Task{
			Name: Ident{
				Name:     "fmt",
				NodeType: NodeIdent,
			},
			Docstring: Comment{
				Text:     " Format the project source",
				NodeType: NodeComment,
			},
			Dependencies: []Node{
				String{
					Text:     "**/*.go",
					NodeType: NodeString,
				},
			},
			Outputs: []Node{},
			Commands: []Command{
				{
					Command:  "go fmt ./...",
					NodeType: NodeCommand,
				},
			},
			NodeType: NodeTask,
		},
		Task{
			Name: Ident{
				Name:     "many",
				NodeType: NodeIdent,
			},
			Docstring: Comment{
				Text:     " Do many things",
				NodeType: NodeComment,
			},
			Dependencies: []Node{},
			Outputs:      []Node{},
			Commands: []Command{
				{
					Command:  "line 1",
					NodeType: NodeCommand,
				},
				{
					Command:  "line 2",
					NodeType: NodeCommand,
				},
				{
					Command:  "line 3",
					NodeType: NodeCommand,
				},
				{
					Command:  "line 4",
					NodeType: NodeCommand,
				},
			},
			NodeType: NodeTask,
		},
		Task{
			Name: Ident{
				Name:     "build",
				NodeType: NodeIdent,
			},
			Docstring: Comment{
				Text:     " Compile the project",
				NodeType: NodeComment,
			},
			Dependencies: []Node{
				String{
					Text:     "**/*.go",
					NodeType: NodeString,
				},
			},
			Outputs: []Node{
				String{
					Text:     "./bin/main",
					NodeType: NodeString,
				},
			},
			Commands: []Command{
				{
					Command:  `go build -ldflags="-X main.version=test -X main.commit=7cb0ec5609efb5fe0"`,
					NodeType: NodeCommand,
				},
			},
			NodeType: NodeTask,
		},
		Task{
			Name: Ident{
				Name:     "show",
				NodeType: NodeIdent,
			},
			Docstring: Comment{
				Text:     " Show the global variables",
				NodeType: NodeComment,
			},
			Dependencies: []Node{},
			Outputs:      []Node{},
			Commands: []Command{
				{
					Command:  "echo {{.GLOBAL}}",
					NodeType: NodeCommand,
				},
			},
			NodeType: NodeTask,
		},
		Task{
			Name: Ident{
				Name:     "moar_things",
				NodeType: NodeIdent,
			},
			Docstring: Comment{
				Text:     " Generate multiple outputs",
				NodeType: NodeComment,
			},
			Dependencies: []Node{},
			Outputs: []Node{
				String{
					Text:     "output1.go",
					NodeType: NodeString,
				},
				String{
					Text:     "output2.go",
					NodeType: NodeString,
				},
			},
			Commands: []Command{
				{
					Command:  "do some stuff here",
					NodeType: NodeCommand,
				},
			},
			NodeType: NodeTask,
		},
		Task{
			Name: Ident{
				Name:     "no_comment",
				NodeType: NodeIdent,
			},
			Docstring:    Comment{},
			Dependencies: []Node{},
			Outputs:      []Node{},
			Commands: []Command{
				{
					Command:  `echo "this task has no docstring"`,
					NodeType: NodeCommand,
				},
			},
			NodeType: NodeTask,
		},
		Task{
			Name: Ident{
				Name:     "makedocs",
				NodeType: NodeIdent,
			},
			Docstring: Comment{
				Text:     " Generate output from a variable",
				NodeType: NodeComment,
			},
			Dependencies: []Node{},
			Outputs: []Node{
				Ident{
					Name:     "DOCS",
					NodeType: NodeIdent,
				},
			},
			Commands: []Command{
				{
					Command:  `echo "making docs"`,
					NodeType: NodeCommand,
				},
			},
			NodeType: NodeTask,
		},
		Task{
			Name: Ident{
				Name:     "makestuff",
				NodeType: NodeIdent,
			},
			Docstring: Comment{
				Text:     " Generate multiple outputs in variables",
				NodeType: NodeComment,
			},
			Dependencies: []Node{},
			Outputs: []Node{
				Ident{
					Name:     "DOCS",
					NodeType: NodeIdent,
				},
				Ident{
					Name:     "BUILD",
					NodeType: NodeIdent,
				},
			},
			Commands: []Command{
				{
					Command:  `echo "doing things"`,
					NodeType: NodeCommand,
				},
			},
			NodeType: NodeTask,
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
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = fullSpokfileAST.String()
	}
}
