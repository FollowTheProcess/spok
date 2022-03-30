// Package graph implements a specialised directed acyclic graph (DAG) and the required
// topological sorting needed for spok's task dependency system.
package graph

import (
	"fmt"

	"github.com/FollowTheProcess/collections/set"
)

// Vertex represents a single node in the graph.
type Vertex struct {
	parents  *set.Set[Vertex] // The direct parents of this vertex
	children *set.Set[Vertex] // The direct children of this vertex
	Name     string           // Uniquely identifiable name
}

// NewVertex creates and returns a new Vertex.
func NewVertex(name string) Vertex {
	return Vertex{
		parents:  set.New[Vertex](),
		children: set.New[Vertex](),
		Name:     name,
	}
}

// InDegree returns the number of incoming edges to this vertex.
func (v Vertex) InDegree() int {
	return v.parents.Length()
}

// OutDegree returns the number of outgoing edges to this vertex.
func (v Vertex) OutDegree() int {
	return v.children.Length()
}

// Graph is a DAG designed to hold spok tasks.
type Graph struct {
	vertices map[string]Vertex // Map of vertex name to vertex
}

// New constructs and returns a new Graph.
func New() *Graph {
	return &Graph{vertices: make(map[string]Vertex)}
}

// AddVertex adds the passed vertex to the graph, if a vertex
// with that name already exists it will be overwritten.
func (g *Graph) AddVertex(v Vertex) {
	g.vertices[v.Name] = v
}

// AddEdge creates an edge connection from parent to child vertices.
func (g *Graph) AddEdge(parent, child Vertex) error {
	parentVertex, ok := g.vertices[parent.Name]
	if !ok {
		return fmt.Errorf("parent vertex %q not in graph", parent.Name)
	}
	childVertex, ok := g.vertices[child.Name]
	if !ok {
		return fmt.Errorf("child vertex %q not in graph", child.Name)
	}

	// Create the connection
	parentVertex.children.Add(child)
	childVertex.parents.Add(parent)

	return nil
}
