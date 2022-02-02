// Package ast defines spok's abstract syntax tree.
package ast

import (
	"fmt"
	"strings"
)

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
	NodeAssign                   // A global variable assignment.
	NodeString                   // A quoted string literal e.g "hello".
	NodeInteger                  // An integer literal e.g. 27.
	NodeFunction                 // A spok builtin function e.g. exec
	NodeTask                     // A spok task.
	NodeCommand                  // A spok task command.
)

// Tree represents the entire AST for a spokfile.
type Tree struct {
	Nodes []Node // List of all AST nodes.
}

// Write out the entire AST to a strings.Builder.
func (t *Tree) Write(s *strings.Builder) {
	for _, n := range t.Nodes {
		n.Write(s)
	}
}

// Append adds an AST node to the Tree.
func (t *Tree) Append(node Node) {
	t.Nodes = append(t.Nodes, node)
}

// isEmpty returns whether or not the AST contains no nodes i.e. empty file.
func (t *Tree) IsEmpty() bool {
	return len(t.Nodes) == 0
}

// Node is an element in the AST.
type Node interface {
	Type() NodeType           // Return the type of the current node.
	String() string           // Pretty print the node.
	Write(s *strings.Builder) // Write the formatted syntax back out to a builder.
}

// CommentNode holds a comment.
type CommentNode struct {
	Text string // The comment text.
	NodeType
}

func (c CommentNode) String() string {
	return fmt.Sprintf("# %s", strings.TrimSpace(c.Text))
}

func (c CommentNode) Write(s *strings.Builder) {
	s.WriteString(c.String())
}

// StringNode holds a string.
type StringNode struct {
	Text string
	NodeType
}

func (s StringNode) String() string {
	return fmt.Sprintf("%q", s.Text)
}

func (s StringNode) Write(sb *strings.Builder) {
	sb.WriteString(s.String())
}

// IntegerNode holds an integer.
type IntegerNode struct {
	Value int
	NodeType
}

func (i IntegerNode) String() string {
	return fmt.Sprintf("%d", i.Value)
}

func (i IntegerNode) Write(s *strings.Builder) {
	s.WriteString(i.String())
}

// IdentNode holds an identifier.
type IdentNode struct {
	Name string // The name of the identifier.
	NodeType
}

func (i IdentNode) String() string {
	return i.Name
}

func (i IdentNode) Write(s *strings.Builder) {
	s.WriteString(i.String())
}

// AssignNode holds a global variable assignment.
type AssignNode struct {
	Name  *IdentNode // The identifier e.g. GIT_COMMIT
	Value Node       // The value it's set to (string, integer, or builtin)
	NodeType
}

func (a AssignNode) String() string {
	return fmt.Sprintf("%s := %s", a.Name.String(), a.Value.String())
}

func (a AssignNode) Write(s *strings.Builder) {
	s.WriteString(a.String())
}

// CommandNode holds a task command.
type CommandNode struct {
	Command string // The shell command to run
	NodeType
}

func (c CommandNode) String() string {
	return c.Command
}

func (c CommandNode) Write(s *strings.Builder) {
	s.WriteString(c.String())
}

// TaskNode holds a spok task.
type TaskNode struct {
	Name         *IdentNode
	Dependencies []Node
	Commands     []*CommandNode
	NodeType
}

func (t TaskNode) String() string {
	s := strings.Builder{}
	deps := []string{}
	commands := []string{}

	for _, p := range t.Dependencies {
		deps = append(deps, p.String())
	}

	for _, c := range t.Commands {
		commands = append(commands, c.String())
	}
	s.WriteString("task ")
	s.WriteString(t.Name.String())
	s.WriteString("(")
	s.WriteString(strings.Join(deps, ", "))
	s.WriteString(")")
	s.WriteString(" {\n\t")
	s.WriteString(strings.Join(commands, "\n"))
	s.WriteString("\n")
	s.WriteString("}")

	return s.String()
}

func (t TaskNode) Write(s *strings.Builder) {
	s.WriteString(t.String())
}
