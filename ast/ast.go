// Package ast defines spok's abstract syntax tree.
package ast

import (
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

const (
	// Average spokfile has about 2000 characters.
	bufferGrowSize = 2000
)

// Tree represents the entire AST for a spokfile.
type Tree struct {
	Nodes []Node // List of all AST nodes.
}

// String allows us to pretty print an entire file for e.g. automatic formatting.
func (t Tree) String() string {
	s := &strings.Builder{}
	s.Grow(bufferGrowSize)
	t.Write(s)
	return s.String()
}

// Write out the entire AST to a strings.Builder.
func (t Tree) Write(s *strings.Builder) {
	for _, n := range t.Nodes {
		n.Write(s)
	}
}

// Append adds an AST node to the Tree.
func (t *Tree) Append(node Node) {
	t.Nodes = append(t.Nodes, node)
}

// IsEmpty returns whether or not the AST contains no nodes i.e. empty file.
func (t *Tree) IsEmpty() bool {
	return len(t.Nodes) == 0
}

// Node is an element in the AST.
type Node interface {
	Type() NodeType           // Return the type of the current node.
	String() string           // Pretty print the node.
	Literal() string          // The go literal representation of the node, saves us from using type conversion.
	Write(s *strings.Builder) // Write the formatted syntax back out to a builder.
}

// Comment holds a comment.
type Comment struct {
	Text string // The comment text.
	NodeType
}

func (c Comment) String() string {
	if c.Text != "" {
		return "# " + strings.TrimSpace(c.Text) + "\n"
	}
	return ""
}

// Literal returns the go literal version of the comment e.g. "# This is a comment".
func (c Comment) Literal() string {
	return c.String()
}

func (c Comment) Write(s *strings.Builder) {
	s.WriteString(c.String())
}

// String holds a string.
type String struct {
	Text string
	NodeType
}

// String returns the pretty representation of a string, used for printing out a spokfile during e.g. --fmt.
func (s String) String() string {
	return `"` + s.Text + `"`
}

// Literal returns the go literal value of the string, usable in source code e.g. "hello".
func (s String) Literal() string {
	return s.Text
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

// Literal returns the go representation of an Ident e.g. "GLOBAL".
func (i Ident) Literal() string {
	return i.Name
}

func (i Ident) Write(s *strings.Builder) {
	s.WriteString(i.String())
}

// Assign holds a global variable assignment.
type Assign struct {
	Value Node  // The value it's set to
	Name  Ident // The name of the identifier
	NodeType
}

func (a Assign) String() string {
	return a.Name.String() + " := " + a.Value.String() + "\n"
}

func (a Assign) Literal() string {
	return a.String()
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

func (c Command) Literal() string {
	return c.Command
}

func (c Command) Write(s *strings.Builder) {
	s.WriteString(c.String())
}

// Task holds a spok task.
type Task struct {
	Name         Ident     // The name of the task
	Docstring    Comment   // Task docstring comment
	Dependencies []Node    // Task dependencies
	Outputs      []Node    // Task outputs
	Commands     []Command // Shell commands to run
	NodeType
}

func (t Task) String() string {
	s := strings.Builder{}

	deps := make([]string, 0, len(t.Dependencies))
	commands := make([]string, 0, len(t.Commands))

	if len(t.Dependencies) != 0 {
		for _, dep := range t.Dependencies {
			deps = append(deps, dep.String())
		}
	}

	if len(t.Commands) != 0 {
		for _, c := range t.Commands {
			commands = append(commands, c.String())
		}
	}

	s.WriteString(t.Docstring.String())

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
			outs := make([]string, 0, len(t.Outputs))
			for _, output := range t.Outputs {
				outs = append(outs, output.String())
			}
			s.WriteString(strings.Join(outs, ", "))
			s.WriteString(")")
		}
	}
	s.WriteString(" {\n")
	for _, command := range commands {
		s.WriteString("    " + command + "\n")
	}
	s.WriteString("}\n\n")

	return s.String()
}

func (t Task) Literal() string {
	return t.String()
}

func (t Task) Write(s *strings.Builder) {
	s.WriteString(t.String())
}

// Function holds a spok builtin function e.g. 'exec' or 'join'.
type Function struct {
	Name      Ident  // Function name
	Arguments []Node // Functions arguments
	NodeType
}

func (f Function) String() string {
	args := make([]string, 0, len(f.Arguments))

	for _, arg := range f.Arguments {
		args = append(args, arg.String())
	}
	return f.Name.String() + "(" + strings.Join(args, ", ") + ")"
}

func (f Function) Literal() string {
	return f.String()
}

func (f Function) Write(s *strings.Builder) {
	s.WriteString(f.String())
}
