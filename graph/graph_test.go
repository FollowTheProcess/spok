package graph

import (
	"testing"
)

func TestGraph_AddVertex(t *testing.T) {
	graph := New()
	v1 := NewVertex("new")

	if len(graph.vertices) != 0 {
		t.Errorf("New graph does not have 0 vertices, got %d", len(graph.vertices))
	}

	graph.AddVertex(v1)

	if len(graph.vertices) != 1 {
		t.Error("Vertex was not correctly added to graph")
	}
}

func TestGraph_AddEdge(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		graph := New()
		v1 := NewVertex("v1")
		v2 := NewVertex("v2")

		graph.AddVertex(v1)
		graph.AddVertex(v2)

		if err := graph.AddEdge(v1.Name, v2.Name); err != nil {
			t.Fatalf("AddEdge returned an error: %v", err)
		}

		retrievedV1, ok := graph.vertices["v1"]
		if !ok {
			t.Fatal("v1 not in graph")
		}
		retrievedV2, ok := graph.vertices["v2"]
		if !ok {
			t.Fatal("v2 not in graph")
		}

		// If connection was successful, v1 should have v2 as a child and
		// v2 should have v1 as a parent
		if !retrievedV1.children.Contains("v2") {
			t.Error("v1 did not have v2 as a child")
		}

		if !retrievedV2.parents.Contains("v1") {
			t.Error("v2 did not have v1 as a parent")
		}
	})

	t.Run("parent missing", func(t *testing.T) {
		graph := New()
		v2 := NewVertex("v2")

		graph.AddVertex(v2)

		if err := graph.AddEdge("parent", v2.Name); err == nil {
			t.Error("expected an error, got nil")
		}
	})

	t.Run("child missing", func(t *testing.T) {
		graph := New()
		v1 := NewVertex("v1")

		graph.AddVertex(v1)

		if err := graph.AddEdge(v1.Name, "child"); err == nil {
			t.Error("expected an error, got nil")
		}
	})
}
