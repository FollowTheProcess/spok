package graph_test

import (
	"testing"

	"github.com/FollowTheProcess/spok/graph"
	"github.com/FollowTheProcess/spok/task"
)

func TestGraph_AddVertex(t *testing.T) {
	t.Parallel()
	g := graph.New()
	v1 := graph.NewVertex(task.Task{Name: "v1"})

	if g.Size() != 0 {
		t.Errorf("New graph does not have 0 vertices, got %d", g.Size())
	}

	g.AddVertex(v1)

	if g.Size() != 1 {
		t.Error("Vertex was not correctly added to graph")
	}
}

func TestGraph_Size(t *testing.T) {
	t.Parallel()
	g := graph.New()
	v1 := graph.NewVertex(task.Task{Name: "v1"})
	v2 := graph.NewVertex(task.Task{Name: "v2"})

	if g.Size() != 0 {
		t.Errorf("Size() of empty graph was not 0, got %d", g.Size())
	}

	g.AddVertex(v1)

	if g.Size() != 1 {
		t.Errorf("Size() of graph was not 1, got %d", g.Size())
	}

	g.AddVertex(v2)

	if g.Size() != 2 {
		t.Errorf("Size() of graph was not 2, got %d", g.Size())
	}
}

func TestGraph_GetVertex(t *testing.T) {
	t.Parallel()
	g := graph.New()
	v1 := graph.NewVertex(task.Task{Name: "v1"})
	v2 := graph.NewVertex(task.Task{Name: "v2"})

	// Trying to get one now should return nil, false
	v, ok := g.GetVertex("v1")
	if v != nil {
		t.Errorf("GetVertex should return nil, got %v", v)
	}
	if ok {
		t.Errorf("GetVertex should return false, got %v", ok)
	}

	g.AddVertex(v1)
	g.AddVertex(v2)

	v, ok = g.GetVertex("v1")

	if v == nil {
		t.Error("GetVertex should return the vertex, got nil")
	}
	if !ok {
		t.Errorf("GetVertex should return true, got %v", ok)
	}
}

func TestGraph_ContainsVertex(t *testing.T) {
	t.Parallel()
	g := graph.New()
	v1 := graph.NewVertex(task.Task{Name: "v1"})
	v2 := graph.NewVertex(task.Task{Name: "v2"})

	// Shouldn't exist yet
	if g.ContainsVertex("v1") {
		t.Error("v1 is not in the graph but ContainsVertex returned true")
	}
	if g.ContainsVertex("v2") {
		t.Error("v2 is not in the graph but ContainsVertex returned true")
	}

	g.AddVertex(v1)
	g.AddVertex(v2)

	if !g.ContainsVertex("v1") {
		t.Error("v1 is in the graph but ContainsVertex returned false")
	}
	if !g.ContainsVertex("v2") {
		t.Error("v2 is in the graph but ContainsVertex returned false")
	}
}

func TestGraph_AddEdge(t *testing.T) {
	t.Parallel()
	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		g := graph.New()
		v1 := graph.NewVertex(task.Task{Name: "v1"})
		v2 := graph.NewVertex(task.Task{Name: "v2"})

		g.AddVertex(v1)
		g.AddVertex(v2)

		if err := g.AddEdge(v1, v2); err != nil {
			t.Fatalf("AddEdge returned an error: %v", err)
		}

		retrievedV1, ok := g.GetVertex("v1")
		if !ok {
			t.Fatal("v1 not in graph")
		}
		retrievedV2, ok := g.GetVertex("v2")
		if !ok {
			t.Fatal("v2 not in graph")
		}

		// If connection was successful, v1 should have v2 as a child and
		// v2 should have v1 as a parent
		if !retrievedV1.HasChild(v2) {
			t.Error("v1 did not have v2 as a child")
		}

		if !retrievedV2.HasParent(v1) {
			t.Error("v2 did not have v1 as a parent")
		}
	})

	t.Run("parent missing", func(t *testing.T) {
		t.Parallel()
		g := graph.New()
		v2 := graph.NewVertex(task.Task{Name: "v2"})

		g.AddVertex(v2)

		if err := g.AddEdge(graph.NewVertex(task.Task{Name: "v1"}), v2); err == nil {
			t.Error("expected an error, got nil")
		}
	})

	t.Run("child missing", func(t *testing.T) {
		t.Parallel()
		g := graph.New()
		v1 := graph.NewVertex(task.Task{Name: "v1"})

		g.AddVertex(v1)

		if err := g.AddEdge(v1, graph.NewVertex(task.Task{Name: "v2"})); err == nil {
			t.Error("expected an error, got nil")
		}
	})
}

func TestSort(t *testing.T) {
	t.Parallel()
	g := graph.New()

	v1 := graph.NewVertex(task.Task{Name: "v1"})
	v2 := graph.NewVertex(task.Task{Name: "v2"})
	v3 := graph.NewVertex(task.Task{Name: "v3"})
	v4 := graph.NewVertex(task.Task{Name: "v4"})
	v5 := graph.NewVertex(task.Task{Name: "v5"})

	g.AddVertex(v1)
	g.AddVertex(v2)
	g.AddVertex(v3)
	g.AddVertex(v4)
	g.AddVertex(v5)

	// v2 depends on v1
	if err := g.AddEdge(v1, v2); err != nil {
		t.Fatalf("graph.AddEdge returned an error: %v", err)
	}

	// v4 depends on v3
	if err := g.AddEdge(v3, v4); err != nil {
		t.Fatalf("graph.AddEdge returned an error: %v", err)
	}

	sorted, err := g.Sort()
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
	t.Parallel()
	g := graph.New()

	v1 := graph.NewVertex(task.Task{Name: "v1"})
	v2 := graph.NewVertex(task.Task{Name: "v2"})
	v3 := graph.NewVertex(task.Task{Name: "v3"})
	v4 := graph.NewVertex(task.Task{Name: "v4"})
	v5 := graph.NewVertex(task.Task{Name: "v5"})

	g.AddVertex(v1)
	g.AddVertex(v2)
	g.AddVertex(v3)
	g.AddVertex(v4)
	g.AddVertex(v5)

	// Purposely make it not a DAG (no vertices with in-degree of 0)
	// easiest way is just connect everything to everything else so there's a cycle

	// v2 depends on v1
	if err := g.AddEdge(v1, v2); err != nil {
		t.Fatalf("graph.AddEdge returned an error: %v", err)
	}

	// v3 depends on v2
	if err := g.AddEdge(v2, v3); err != nil {
		t.Fatalf("graph.AddEdge returned an error: %v", err)
	}

	// v4 depends on v3
	if err := g.AddEdge(v3, v4); err != nil {
		t.Fatalf("graph.AddEdge returned an error: %v", err)
	}

	// v4 depends on v1
	if err := g.AddEdge(v1, v4); err != nil {
		t.Fatalf("graph.AddEdge returned an error: %v", err)
	}

	// v5 depends on v4
	if err := g.AddEdge(v4, v5); err != nil {
		t.Fatalf("graph.AddEdge returned an error: %v", err)
	}

	// Complete the cycle: v1 also depends on v5
	if err := g.AddEdge(v4, v1); err != nil {
		t.Fatalf("graph.AddEdge returned an error: %v", err)
	}

	_, err := g.Sort()
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
func makeGraph() *graph.Graph {
	g := graph.New()

	v1 := graph.NewVertex(task.Task{Name: "v1"})
	v2 := graph.NewVertex(task.Task{Name: "v2"})
	v3 := graph.NewVertex(task.Task{Name: "v3"})
	v4 := graph.NewVertex(task.Task{Name: "v4"})
	v5 := graph.NewVertex(task.Task{Name: "v5"})

	g.AddVertex(v1)
	g.AddVertex(v2)
	g.AddVertex(v3)
	g.AddVertex(v4)
	g.AddVertex(v5)

	// v2 depends on v1
	_ = g.AddEdge(v1, v2) //nolint: errcheck

	// v4 depends on v3
	_ = g.AddEdge(v3, v4) //nolint: errcheck

	return g
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
