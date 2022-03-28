package graph

import (
	"testing"
)

func TestVertex_InDegree(t *testing.T) {
	tests := []struct {
		name string
		v    Vertex
		want int
	}{
		{
			name: "no inbound edges",
			v: Vertex{
				parents: make(map[string]struct{}),
			},
			want: 0,
		},
		{
			name: "nil map",
			v: Vertex{
				parents: nil,
			},
			want: 0,
		},
		{
			name: "one inbound edge",
			v: Vertex{
				parents: map[string]struct{}{"one": {}},
			},
			want: 1,
		},
		{
			name: "two inbound edges",
			v: Vertex{
				parents: map[string]struct{}{"one": {}, "two": {}},
			},
			want: 2,
		},
		{
			name: "5 inbound edges",
			v: Vertex{
				parents: map[string]struct{}{
					"one":   {},
					"two":   {},
					"three": {},
					"four":  {},
					"five":  {},
				},
			},
			want: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v.InDegree(); got != tt.want {
				t.Errorf("got %d, wanted %d", got, tt.want)
			}
		})
	}
}

func TestVertex_OutDegree(t *testing.T) {
	tests := []struct {
		name string
		v    Vertex
		want int
	}{
		{
			name: "no outbound edges",
			v: Vertex{
				children: make(map[string]struct{}),
			},
			want: 0,
		},
		{
			name: "nil map",
			v: Vertex{
				children: nil,
			},
			want: 0,
		},
		{
			name: "one outbound edge",
			v: Vertex{
				children: map[string]struct{}{"one": {}},
			},
			want: 1,
		},
		{
			name: "two outbound edges",
			v: Vertex{
				children: map[string]struct{}{"one": {}, "two": {}},
			},
			want: 2,
		},
		{
			name: "5 outbound edges",
			v: Vertex{
				children: map[string]struct{}{
					"one":   {},
					"two":   {},
					"three": {},
					"four":  {},
					"five":  {},
				},
			},
			want: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v.OutDegree(); got != tt.want {
				t.Errorf("got %d, wanted %d", got, tt.want)
			}
		})
	}
}

func TestGraph_AddVertex(t *testing.T) {
	graph := New()
	v1 := Vertex{
		parents:  make(map[string]struct{}),
		children: make(map[string]struct{}),
		Name:     "new",
	}

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
		v1 := Vertex{
			parents:  make(map[string]struct{}),
			children: make(map[string]struct{}),
			Name:     "v1",
		}
		v2 := Vertex{
			parents:  make(map[string]struct{}),
			children: make(map[string]struct{}),
			Name:     "v2",
		}

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
		_, ok = retrievedV1.children["v2"]
		if !ok {
			t.Error("v1 did not have v2 as a child")
		}
		_, ok = retrievedV2.parents["v1"]
		if !ok {
			t.Error("v2 did not have v1 as a parent")
		}
	})

	t.Run("parent missing", func(t *testing.T) {
		graph := New()
		v2 := Vertex{
			parents:  make(map[string]struct{}),
			children: make(map[string]struct{}),
			Name:     "v2",
		}

		graph.AddVertex(v2)

		if err := graph.AddEdge("parent", v2.Name); err == nil {
			t.Error("expected an error, got nil")
		}
	})

	t.Run("child missing", func(t *testing.T) {
		graph := New()
		v1 := Vertex{
			parents:  make(map[string]struct{}),
			children: make(map[string]struct{}),
			Name:     "v1",
		}

		graph.AddVertex(v1)

		if err := graph.AddEdge(v1.Name, "child"); err == nil {
			t.Error("expected an error, got nil")
		}
	})
}
