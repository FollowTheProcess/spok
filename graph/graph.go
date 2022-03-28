// Package graph implements a specialised directed acyclic graph (DAG) and the required
// topological sorting needed for spok's task dependency system.
package graph

import "fmt"

// Vertex represents a single node in the graph.
type Vertex struct {
	parents  map[string]struct{} // The direct parents of this vertex
	children map[string]struct{} // The direct children of this vertex
	Name     string              // Uniquely identifiable name
}

// InDegree returns the number of incoming edges to this vertex.
func (v Vertex) InDegree() int {
	return len(v.parents)
}

// OutDegree returns the number of outgoing edges to this vertex.
func (v Vertex) OutDegree() int {
	return len(v.children)
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
func (g *Graph) AddEdge(parent, child string) error {
	parentVertex, ok := g.vertices[parent]
	if !ok {
		return fmt.Errorf("parent vertex %q not in graph", parent)
	}
	childVertex, ok := g.vertices[child]
	if !ok {
		return fmt.Errorf("child vertex %q not in graph", child)
	}

	// Create the connection
	parentVertex.children[child] = struct{}{}
	childVertex.parents[parent] = struct{}{}

	return nil
}
