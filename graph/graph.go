// Package graph implements a specialised directed acyclic graph (DAG) and the required
// topological sorting needed for spok's task dependency system.
package graph

import (
	"fmt"

	"github.com/FollowTheProcess/collections/queue"
	"github.com/FollowTheProcess/collections/set"
	"github.com/FollowTheProcess/spok/task"
)

// Vertex represents a single node in the graph.
type Vertex struct {
	parents  *set.Set[*Vertex] // The direct parents of this vertex
	children *set.Set[*Vertex] // The direct children of this vertex
	Name     string            // Uniquely identifiable name
	Task     task.Task         // The actual underlying task represented by this vertex
	InDegree int               // Number of incoming edges
}

// NewVertex creates and returns a new Vertex.
func NewVertex(task task.Task) *Vertex {
	return &Vertex{
		parents:  set.New[*Vertex](),
		children: set.New[*Vertex](),
		Task:     task,
		Name:     task.Name,
		InDegree: 0,
	}
}

// Graph is a DAG designed to hold spok tasks.
type Graph struct {
	vertices map[string]*Vertex // Map of vertex name to vertex
}

// New constructs and returns a new Graph.
func New() *Graph {
	return &Graph{vertices: make(map[string]*Vertex)}
}

// AddVertex adds the passed vertex to the graph, if a vertex
// with that name already exists it will be overwritten.
func (g *Graph) AddVertex(v *Vertex) {
	g.vertices[v.Name] = v
}

// AddEdge creates an edge connection from parent to child vertices.
func (g *Graph) AddEdge(parent, child *Vertex) error {
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

// Sort topologically sorts the graph and returns a vertex slice in the correct order.
func (g *Graph) Sort() ([]*Vertex, error) {
	zeroInDegreeQueue := queue.New[*Vertex]()
	result := make([]*Vertex, 0, len(g.vertices))

	for _, vertex := range g.vertices {
		vertex.InDegree = vertex.parents.Length() // Compute in degree for each vertex

		// Put all vertices with 0 in-degree into the queue
		if vertex.InDegree == 0 {
			zeroInDegreeQueue.Push(vertex)
		}
	}

	// Bailout point: if there is not at least 1 vertex with 0 in-degree
	// it's not a DAG and cannot be sorted
	if zeroInDegreeQueue.IsEmpty() {
		return nil, fmt.Errorf("not a DAG")
	}

	// While queue is not empty
	for !zeroInDegreeQueue.IsEmpty() {
		vertex, err := zeroInDegreeQueue.Pop() // Pop a vertex off the queue
		if err != nil {
			return nil, err
		}

		// Add it to the result slice
		result = append(result, vertex)

		// For each child, reduce in-degree by 1
		for _, child := range vertex.children.Items() {
			child.InDegree--

			// If any are now 0, add to the queue
			if child.InDegree == 0 {
				zeroInDegreeQueue.Push(child)
			}
		}
	}

	return result, nil
}