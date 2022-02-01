package ast

import "testing"

func TestCommentNode(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{
			name: "normal",
			text: "A comment",
			want: "# A comment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := CommentNode{Text: tt.text, NodeType: NodeComment}
			if got := node.String(); got != tt.want {
				t.Errorf("got %s, wanted %s", got, tt.want)
			}
		})
	}
}

func TestIdentNode(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{
			name: "test",
			text: "x",
			want: "x",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := IdentNode{Name: tt.text, NodeType: NodeIdent}
			if got := node.String(); got != tt.want {
				t.Errorf("got %s, wanted %s", got, tt.want)
			}
		})
	}
}

func TestAssignNode(t *testing.T) {
	tests := []struct {
		name  string
		left  *IdentNode
		right Node
		want  string
	}{
		{
			name:  "string",
			left:  &IdentNode{Name: "GIT_COMMIT", NodeType: NodeIdent},
			right: StringNode{Text: "abd825efd017df", NodeType: NodeString},
			want:  `GIT_COMMIT := "abd825efd017df"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assign := AssignNode{Name: tt.left, Value: tt.right, NodeType: NodeAssign}
			if got := assign.String(); got != tt.want {
				t.Errorf("got %s, wanted %s", got, tt.want)
			}
		})
	}
}
