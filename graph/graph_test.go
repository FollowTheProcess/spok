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

		if err := graph.AddEdge(v1, v2); err != nil {
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
		if !retrievedV1.children.Contains(v2) {
			t.Error("v1 did not have v2 as a child")
		}

		if !retrievedV2.parents.Contains(v1) {
			t.Error("v2 did not have v1 as a parent")
		}
	})

	t.Run("parent missing", func(t *testing.T) {
		graph := New()
		v2 := NewVertex("v2")

		graph.AddVertex(v2)

		if err := graph.AddEdge(NewVertex("v1"), v2); err == nil {
			t.Error("expected an error, got nil")
		}
	})

	t.Run("child missing", func(t *testing.T) {
		graph := New()
		v1 := NewVertex("v1")

		graph.AddVertex(v1)

		if err := graph.AddEdge(v1, NewVertex("v2")); err == nil {
			t.Error("expected an error, got nil")
		}
	})
}

func TestSort(t *testing.T) {
	graph := New()

	v1 := NewVertex("v1")
	v2 := NewVertex("v2")
	v3 := NewVertex("v3")
	v4 := NewVertex("v4")
	v5 := NewVertex("v5")

	graph.AddVertex(v1)
	graph.AddVertex(v2)
	graph.AddVertex(v3)
	graph.AddVertex(v4)
	graph.AddVertex(v5)

	// v2 depends on v1
	if err := graph.AddEdge(v1, v2); err != nil {
		t.Fatalf("graph.AddEdge returned an error: %v", err)
	}

	// v4 depends on v3
	if err := graph.AddEdge(v3, v4); err != nil {
		t.Fatalf("graph.AddEdge returned an error: %v", err)
	}

	sorted, err := graph.Sort()
	if err != nil {
		t.Fatalf("graph.Sort returned an error: %v", err)
	}

	// A DAG may have more than one possible topological sort
	possibles := [][]string{
		{"v5", "v1", "v3", "v2", "v4"},
		{"v1", "v3", "v5", "v2", "v4"},
		{"v3", "v5", "v1", "v4", "v2"},
	}

	var got []string
	for _, vertex := range sorted {
		got = append(got, vertex.Name)
	}

	if !isInPossibleSolutions(got, possibles) {
		t.Errorf("DAG not sorted correctly: got:\n%#v\nwanted one of:\n%#v", got, possibles)
	}
}

func TestSortNotADAG(t *testing.T) {
	graph := New()

	v1 := NewVertex("v1")
	v2 := NewVertex("v2")
	v3 := NewVertex("v3")
	v4 := NewVertex("v4")
	v5 := NewVertex("v5")

	graph.AddVertex(v1)
	graph.AddVertex(v2)
	graph.AddVertex(v3)
	graph.AddVertex(v4)
	graph.AddVertex(v5)

	// Purposely make it not a DAG (no vertices with in-degree of 0)
	// easiest way is just connect everything to everything else so there's a cycle

	// v2 depends on v1
	if err := graph.AddEdge(v1, v2); err != nil {
		t.Fatalf("graph.AddEdge returned an error: %v", err)
	}

	// v3 depends on v2
	if err := graph.AddEdge(v2, v3); err != nil {
		t.Fatalf("graph.AddEdge returned an error: %v", err)
	}

	// v4 depends on v3
	if err := graph.AddEdge(v3, v4); err != nil {
		t.Fatalf("graph.AddEdge returned an error: %v", err)
	}

	// v4 depends on v1
	if err := graph.AddEdge(v1, v4); err != nil {
		t.Fatalf("graph.AddEdge returned an error: %v", err)
	}

	// v5 depends on v4
	if err := graph.AddEdge(v4, v5); err != nil {
		t.Fatalf("graph.AddEdge returned an error: %v", err)
	}

	// Complete the cycle: v1 also depends on v5
	if err := graph.AddEdge(v4, v1); err != nil {
		t.Fatalf("graph.AddEdge returned an error: %v", err)
	}

	_, err := graph.Sort()
	if err == nil {
		t.Fatal("expected not a dag error, got nil")
	}
}

func isInPossibleSolutions(result []string, possibles [][]string) bool {
	for _, possible := range possibles {
		if equal(result, possible) {
			return true
		}
	}

	return false
}

func equal[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}

	for i, item := range a {
		if b[i] != item {
			return false
		}
	}

	return true
}

// makeGraph makes a simple DAG with a few connections for things like benchmarks.
func makeGraph() *Graph {
	graph := New()

	v1 := NewVertex("v1")
	v2 := NewVertex("v2")
	v3 := NewVertex("v3")
	v4 := NewVertex("v4")
	v5 := NewVertex("v5")

	graph.AddVertex(v1)
	graph.AddVertex(v2)
	graph.AddVertex(v3)
	graph.AddVertex(v4)
	graph.AddVertex(v5)

	// v2 depends on v1
	_ = graph.AddEdge(v1, v2)

	// v4 depends on v3
	_ = graph.AddEdge(v3, v4)

	return graph
}

func BenchmarkGraphSort(b *testing.B) {
	// Because the graph.Sort method alters the state of the graph (reducing vertex.InDegree)
	// a new graph must be constructed for each run meaning this is actually quite slow to run (~1 minute)
	// but we stop and start the timer at the right places to ensure just the sorting code's performance is measured
	for n := 0; n < b.N; n++ {
		b.StopTimer()
		graph := makeGraph()
		b.StartTimer()
		_, err := graph.Sort()
		if err != nil {
			b.Fatalf("graph.Sort returned an error: %v", err)
		}
	}
}
