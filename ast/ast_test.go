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
			node: CommentNode{Text: " A comment", NodeType: NodeComment},
			want: "# A comment",
		},
		{
			name: "string",
			node: StringNode{Text: "hello", NodeType: NodeString},
			want: `"hello"`,
		},
		{
			name: "ident",
			node: IdentNode{Name: "GIT_COMMIT", NodeType: NodeIdent},
			want: "GIT_COMMIT",
		},
		{
			name: "assign",
			node: AssignNode{
				Name:     &IdentNode{Name: "GIT_COMMIT", NodeType: NodeIdent},
				Value:    StringNode{Text: "a2736ef997c926", NodeType: NodeString},
				NodeType: NodeAssign,
			},
			want: `GIT_COMMIT := "a2736ef997c926"`,
		},
		{
			name: "command",
			node: CommandNode{Command: "go test ./...", NodeType: NodeCommand},
			want: "go test ./...",
		},
		{
			name: "basic task",
			node: TaskNode{
				Name: &IdentNode{
					Name:     "test",
					NodeType: NodeIdent,
				},
				Dependencies: []Node{
					&StringNode{
						Text:     "file.go",
						NodeType: NodeString,
					},
				},
				Commands: []*CommandNode{
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
			node: TaskNode{
				Name: &IdentNode{
					Name:     "test",
					NodeType: NodeIdent,
				},
				Commands: []*CommandNode{
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
				t.Errorf("got %s, wanted %s", got, tt.want)
			}
		})
	}
}
