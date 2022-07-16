// Package graph implements a specialised directed acyclic graph (DAG) and the required
// topological sorting needed for spok's task dependency system.
package graph

import (
	"errors"
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
	inDegree int               // Number of incoming edges
}

// NewVertex creates and returns a new Vertex.
func NewVertex(task task.Task) *Vertex {
	return &Vertex{
		parents:  set.New[*Vertex](),
		children: set.New[*Vertex](),
		Task:     task,
		Name:     task.Name,
		inDegree: 0,
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
		vertex.inDegree = vertex.parents.Length() // Compute in degree for each vertex

		// Put all vertices with 0 in-degree into the queue
		if vertex.inDegree == 0 {
			zeroInDegreeQueue.Push(vertex)
		}
	}

	// Bailout point: if there is not at least 1 vertex with 0 in-degree
	// it's not a DAG and cannot be sorted
	if zeroInDegreeQueue.IsEmpty() {
		return nil, errors.New("Not a DAG")
	}

	// While queue is not empty
	for !zeroInDegreeQueue.IsEmpty() {
		// Only error here is pop from empty queue, but we know
		// the queue is not empty in this loop so no point checking
		vertex, _ := zeroInDegreeQueue.Pop() // Pop a vertex off the queue

		// Add it to the result slice
		result = append(result, vertex)

		// For each child, reduce in-degree by 1
		for _, child := range vertex.children.Items() {
			child.inDegree--

			// If any are now 0, add to the queue
			if child.inDegree == 0 {
				zeroInDegreeQueue.Push(child)
			}
		}
	}

	return result, nil
}
