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
	NodeFunction                 // A spok builtin function e.g. exec
	NodeTask                     // A spok task.
	NodeCommand                  // A spok task command.
)

// Tree represents the entire AST for a spokfile.
type Tree struct {
	Nodes []Node // List of all AST nodes.
}

func (t *Tree) String() string {
	s := &strings.Builder{}
	t.Write(s)
	return s.String()
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

// Comment holds a comment.
type Comment struct {
	Text string // The comment text.
	NodeType
}

func (c Comment) String() string {
	if c.Text != "" {
		return fmt.Sprintf("# %s\n", strings.TrimSpace(c.Text))
	}
	return ""
}

func (c Comment) Write(s *strings.Builder) {
	s.WriteString(c.String())
}

// String holds a string.
type String struct {
	Text string
	NodeType
}

func (s String) String() string {
	return fmt.Sprintf("%q", s.Text)
}

func (s String) Write(sb *strings.Builder) {
	sb.WriteString(s.String())
}

// Ident holds an identifier.
type Ident struct {
	Name string // The name of the identifier.
	NodeType
}

func (i Ident) String() string {
	return i.Name
}

func (i Ident) Write(s *strings.Builder) {
	s.WriteString(i.String())
}

// Assign holds a global variable assignment.
type Assign struct {
	Name  *Ident // The identifier e.g. GIT_COMMIT
	Value Node   // The value it's set to (string, or builtin)
	NodeType
}

func (a Assign) String() string {
	return fmt.Sprintf("%s := %s\n", a.Name.String(), a.Value.String())
}

func (a Assign) Write(s *strings.Builder) {
	s.WriteString(a.String())
}

// Command holds a task command.
type Command struct {
	Command string // The shell command to run
	NodeType
}

func (c Command) String() string {
	return c.Command
}

func (c Command) Write(s *strings.Builder) {
	s.WriteString(c.String())
}

// Task holds a spok task.
type Task struct {
	Name         *Ident
	Docstring    *Comment
	Dependencies []Node
	Outputs      []Node
	Commands     []*Command
	NodeType
}

func (t Task) String() string {
	s := strings.Builder{}
	deps := []string{}
	commands := []string{}

	if len(t.Dependencies) != 0 {
		for _, p := range t.Dependencies {
			deps = append(deps, p.String())
		}
	}

	if len(t.Commands) != 0 {
		for _, c := range t.Commands {
			commands = append(commands, c.String())
		}
	}

	if t.Docstring != nil {
		s.WriteString(t.Docstring.String())
		s.WriteString("\n")
	}
	s.WriteString("task ")
	s.WriteString(t.Name.String())
	s.WriteString("(")
	s.WriteString(strings.Join(deps, ", "))
	s.WriteString(")")
	if len(t.Outputs) != 0 {
		s.WriteString(" -> ")
		switch len(t.Outputs) {
		case 1:
			s.WriteString(t.Outputs[0].String())
		default:
			s.WriteString("(")
			outs := []string{}
			for _, output := range t.Outputs {
				outs = append(outs, output.String())
			}
			s.WriteString(strings.Join(outs, ", "))
			s.WriteString(")")
		}
	}
	s.WriteString(" {\n")
	for _, command := range commands {
		s.WriteString(fmt.Sprintf("    %s\n", command))
	}
	s.WriteString("}\n")

	return s.String()
}

func (t Task) Write(s *strings.Builder) {
	s.WriteString(t.String())
}

type Function struct {
	Name      *Ident
	Arguments []Node
	NodeType
}

func (f Function) String() string {
	args := []string{}

	for _, arg := range f.Arguments {
		args = append(args, arg.String())
	}

	return fmt.Sprintf("%s(%s)\n", f.Name.String(), strings.Join(args, ", "))
}

func (f Function) Write(s *strings.Builder) {
	s.WriteString(f.String())
}
