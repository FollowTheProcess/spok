package ast

import "testing"

func TestNodeString(t *testing.T) {
	tests := []struct {
		node Node
		name string
		want string
	}{
		{
			name: "comment",
			node: Comment{Text: " A comment", NodeType: NodeComment},
			want: "# A comment",
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
				Name:     &Ident{Name: "GIT_COMMIT", NodeType: NodeIdent},
				Value:    String{Text: "a2736ef997c926", NodeType: NodeString},
				NodeType: NodeAssign,
			},
			want: `GIT_COMMIT := "a2736ef997c926"`,
		},
		{
			name: "command",
			node: Command{Command: "go test ./...", NodeType: NodeCommand},
			want: "go test ./...",
		},
		{
			name: "function",
			node: Function{
				Name: &Ident{
					Name:     "exec",
					NodeType: NodeIdent,
				},
				Arguments: []Node{
					String{Text: "git rev-parse HEAD", NodeType: NodeString},
				}},
			want: `exec("git rev-parse HEAD")`,
		},
		{
			name: "basic task",
			node: Task{
				Name: &Ident{
					Name:     "test",
					NodeType: NodeIdent,
				},
				Dependencies: []Node{
					&String{
						Text:     "file.go",
						NodeType: NodeString,
					},
				},
				Commands: []*Command{
					{
						Command:  "go test ./...",
						NodeType: NodeCommand,
					},
				},
				NodeType: NodeTask,
			},
			want: `task test("file.go") {
    go test ./...
}`,
		},
		{
			name: "task no args",
			node: Task{
				Name: &Ident{
					Name:     "test",
					NodeType: NodeIdent,
				},
				Commands: []*Command{
					{
						Command:  "go test ./...",
						NodeType: NodeCommand,
					},
				},
				NodeType: NodeTask,
			},
			want: `task test() {
    go test ./...
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.node.String(); got != tt.want {
				t.Errorf("got %#v, wanted %#v", got, tt.want)
			}
		})
	}
}
