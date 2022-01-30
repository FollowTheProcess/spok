package parser

import (
	"testing"

	"github.com/FollowTheProcess/spok/ast"
)

func TestParseComment(t *testing.T) {
	input := `# A comment`

	p := New(input)
	tree, err := p.Parse()
	if err != nil {
		t.Fatalf("Parser returned en error token: %v", err)
	}

	if tree == nil {
		t.Fatalf("Parser returned a nil AST")
	}

	if len(tree.Nodes) != 1 {
		t.Fatalf("Wrong number of ast nodes, got %d, wanted %d", len(tree.Nodes), 1)
	}

	node := tree.Nodes[0]
	comment, ok := node.(ast.CommentNode)
	if !ok {
		t.Fatalf("Node was not a comment node, got %T", node)
	}

	if comment.Text != " A comment" {
		t.Errorf("Wrong comment text: got %s, wanted %s", comment.Text, " A comment")
	}

}
