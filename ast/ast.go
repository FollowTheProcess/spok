// Package ast defines spok's abstract syntax tree.
package ast

import "strings"

// NodeType identifies the type of an AST node.
type NodeType int

// Type returns itself and allows easy embedding into AST nodes
// to enable e.g. NodeTask.Type().
func (t NodeType) Type() NodeType {
	return t
}

//go:generate stringer -type=NodeType -output=nodetype_string.go
const (
	NodeComment  NodeType = iota // A spok comment, preceded by a '#'.
	NodeIdent                    // An identifier e.g. global variable or name of a task.
	NodeString                   // A quoted string literal e.g "hello".
	NodeInteger                  // An integer literal e.g. 27.
	NodeFunction                 // A spok builtin function e.g. exec
	NodeTask                     // A spok task.
)

// Tree represents the entire AST for a spokfile.
type Tree struct {
	Nodes []Node // List of all AST nodes.
}

// Node is an element in the AST.
type Node interface {
	Type() NodeType
	String() string
}

// CommentNode holds a comment.
type CommentNode struct {
	Text string // The comment text.
	NodeType
}

func (c CommentNode) String() string {
	return c.Text
}

func (c CommentNode) write(s *strings.Builder) { // nolint: unused
	// TODO: Unused for now but we know we'll need it
	s.WriteString(c.String())
}
