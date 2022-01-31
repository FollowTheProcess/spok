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
