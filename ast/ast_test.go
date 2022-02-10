package ast

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestAppend(t *testing.T) {
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

func TestNodeString(t *testing.T) {
	tests := []struct {
		node Node
		name string
		want string
	}{
		{
			name: "comment",
			node: Comment{Text: " A comment", NodeType: NodeComment},
			want: "# A comment\n",
		},
		{
			name: "string",
			node: String{Text: "hello", NodeType: NodeString},
			want: `"hello"`,
		},
		{
			name: "ident",
			node: Ident{Name: "GIT_COMMIT", NodeType: NodeIdent},
			want: "GIT_COMMIT",
		},
		{
			name: "assign",
			node: Assign{
				Name:     Ident{Name: "GIT_COMMIT", NodeType: NodeIdent},
				Value:    String{Text: "a2736ef997c926", NodeType: NodeString},
				NodeType: NodeAssign,
			},
			want: `GIT_COMMIT := "a2736ef997c926"` + "\n",
		},
		{
			name: "command",
			node: Command{Command: "go test ./...", NodeType: NodeCommand},
			want: "go test ./...",
		},
		{
			name: "function",
			node: Function{
				Name: Ident{
					Name:     "exec",
					NodeType: NodeIdent,
				},
				Arguments: []Node{
					String{Text: "git rev-parse HEAD", NodeType: NodeString},
				}},
			want: `exec("git rev-parse HEAD")` + "\n",
		},
		{
			name: "basic task",
			node: Task{
				Name: Ident{
					Name:     "test",
					NodeType: NodeIdent,
				},
				Dependencies: []Node{
					&String{
						Text:     "file.go",
						NodeType: NodeString,
					},
				},
				Commands: []Command{
					{
						Command:  "go test ./...",
						NodeType: NodeCommand,
					},
				},
				NodeType: NodeTask,
			},
			want: `task test("file.go") {
    go test ./...
}

`,
		},
		{
			name: "task no args",
			node: Task{
				Name: Ident{
					Name:     "test",
					NodeType: NodeIdent,
				},
				Commands: []Command{
					{
						Command:  "go test ./...",
						NodeType: NodeCommand,
					},
				},
				NodeType: NodeTask,
			},
			want: `task test() {
    go test ./...
}

`,
		},
		{
			name: "task with single output",
			node: Task{
				Name: Ident{
					Name:     "build",
					NodeType: NodeIdent,
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
						Command:  "go build .",
						NodeType: NodeCommand,
					},
				},
				NodeType: NodeTask,
			},
			want: `task build("**/*.go") -> "./bin/main" {
    go build .
}

`,
		},
		{
			name: "task with multiple outputs",
			node: Task{
				Name: Ident{
					Name:     "build",
					NodeType: NodeIdent,
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
					Ident{
						Name:     "SOMETHINGELSE",
						NodeType: NodeIdent,
					},
				},
				Commands: []Command{
					{
						Command:  "go build .",
						NodeType: NodeCommand,
					},
				},
				NodeType: NodeTask,
			},
			want: `task build("**/*.go") -> ("./bin/main", SOMETHINGELSE) {
    go build .
}

`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.node.String()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("String mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

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
					Command:  "echo GLOBAL",
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
