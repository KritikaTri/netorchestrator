package topology

import (
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"testing"
)

func build(t testing.TB, links ...Link) *Graph {
	t.Helper()
	g := New()
	for _, l := range links {
		if err := g.AddLink(l.A, l.B, l.Cost); err != nil {
			t.Fatal(err)
		}
	}
	return g
}

func TestAddLinkValidation(t *testing.T) {
	g := New()
	for name, tc := range map[string]struct {
		a, b NodeID
		cost float64
	}{
		"self link":     {"a", "a", 1},
		"empty id":      {"", "b", 1},
		"zero cost":     {"a", "b", 0},
		"negative cost": {"a", "b", -3},
	} {
		if err := g.AddLink(tc.a, tc.b, tc.cost); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
	if len(g.Nodes()) != 0 {
		t.Error("failed AddLink must not leave nodes behind")
	}
}

func TestShortestPathPrefersCheaperLongerRoute(t *testing.T) {
	g := build(t,
		Link{"a", "b", 1}, Link{"b", "d", 1}, Link{"d", "e", 1}, // cost 3, 3 hops
		Link{"a", "e", 10}, // direct but expensive
	)
	p, err := g.ShortestPath("a", "e")
	if err != nil {
		t.Fatal(err)
	}
	if p.Cost != 3 || !reflect.DeepEqual(p.Nodes, []NodeID{"a", "b", "d", "e"}) {
		t.Fatalf("got %+v", p)
	}
}

func TestShortestPathTieBreakIsDeterministic(t *testing.T) {
	// two equal-cost routes a-x-d and a-y-d; must always pick via x.
	g := build(t, Link{"a", "x", 1}, Link{"a", "y", 1}, Link{"x", "d", 1}, Link{"y", "d", 1})
	for i := 0; i < 50; i++ {
		p, _ := g.ShortestPath("a", "d")
		if !reflect.DeepEqual(p.Nodes, []NodeID{"a", "x", "d"}) {
			t.Fatalf("iteration %d: %v", i, p.Nodes)
		}
	}
}

func TestShortestPathErrors(t *testing.T) {
	g := build(t, Link{"a", "b", 1})
	g.AddNode("island")
	if _, err := g.ShortestPath("a", "island"); !errors.Is(err, ErrNoPath) {
		t.Errorf("want ErrNoPath, got %v", err)
	}
	if _, err := g.ShortestPath("a", "nope"); !errors.Is(err, ErrUnknownNode) {
		t.Errorf("want ErrUnknownNode, got %v", err)
	}
	if p, err := g.ShortestPath("a", "a"); err != nil || p.Cost != 0 || len(p.Nodes) != 1 {
		t.Errorf("self path: %+v %v", p, err)
	}
}

func TestComponents(t *testing.T) {
	g := build(t, Link{"a", "b", 1}, Link{"b", "c", 1}, Link{"x", "y", 1})
	g.AddNode("solo")
	got := g.Components()
	want := [][]NodeID{{"a", "b", "c"}, {"x", "y"}, {"solo"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
	if g.Connected() {
		t.Error("graph is not connected")
	}
}

func TestSinglePointsOfFailure(t *testing.T) {
	// triangle (a,b,c) --bridge-- d --bridge-- e ; d is a cut vertex.
	g := build(t,
		Link{"a", "b", 1}, Link{"b", "c", 1}, Link{"a", "c", 1},
		Link{"c", "d", 1}, Link{"d", "e", 1},
	)
	nodes, bridges := g.SinglePointsOfFailure()
	if !reflect.DeepEqual(nodes, []NodeID{"c", "d"}) {
		t.Errorf("articulation points: %v", nodes)
	}
	wantB := []Link{{"c", "d", 1}, {"d", "e", 1}}
	if !reflect.DeepEqual(bridges, wantB) {
		t.Errorf("bridges: %v", bridges)
	}
}

func TestRingHasNoSinglePointOfFailure(t *testing.T) {
	g := build(t, Link{"a", "b", 1}, Link{"b", "c", 1}, Link{"c", "d", 1}, Link{"d", "a", 1})
	n, b := g.SinglePointsOfFailure()
	if len(n) != 0 || len(b) != 0 {
		t.Fatalf("ring should be 2-connected: %v %v", n, b)
	}
}

func TestBlastRadius(t *testing.T) {
	// hub-and-spoke: losing the hub strands all 4 leaves except the largest remainder (1 leaf).
	g := build(t, Link{"hub", "l1", 1}, Link{"hub", "l2", 1}, Link{"hub", "l3", 1}, Link{"hub", "l4", 1})
	n, err := g.BlastRadius("hub")
	if err != nil || n != 3 {
		t.Fatalf("hub: %d %v", n, err)
	}
	if n, _ := g.BlastRadius("l1"); n != 0 {
		t.Errorf("leaf failure should cut nothing, got %d", n)
	}
	if _, err := g.BlastRadius("ghost"); !errors.Is(err, ErrUnknownNode) {
		t.Errorf("want ErrUnknownNode, got %v", err)
	}
}

func TestFromJSON(t *testing.T) {
	g, err := FromJSON(strings.NewReader(`{"nodes":["lonely"],"links":[{"a":"r1","b":"r2","cost":5}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if !g.HasNode("lonely") || len(g.Links()) != 1 {
		t.Fatalf("unexpected graph: %v %v", g.Nodes(), g.Links())
	}
	for _, bad := range []string{
		`{"links":[{"a":"r1","b":"r1","cost":1}]}`,
		`{"links":[{"a":"r1","b":"r2","cost":0}]}`,
		`{"linkz":[]}`,
		`not json`,
	} {
		if _, err := FromJSON(strings.NewReader(bad)); err == nil {
			t.Errorf("expected error for %q", bad)
		}
	}
}

// --- oracle tests -----------------------------------------------------------

func randomGraph(r *rand.Rand, n int, p float64) *Graph {
	g := New()
	for i := 0; i < n; i++ {
		g.AddNode(NodeID(fmt.Sprintf("n%02d", i)))
	}
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			if r.Float64() < p {
				_ = g.AddLink(NodeID(fmt.Sprintf("n%02d", i)), NodeID(fmt.Sprintf("n%02d", j)), float64(1+r.Intn(9)))
			}
		}
	}
	return g
}

// Articulation points and bridges must agree with the definition: removing the
// element increases the number of connected components.
func TestSinglePointsOfFailureMatchesBruteForce(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	for trial := 0; trial < 300; trial++ {
		g := randomGraph(r, 3+r.Intn(10), 0.1+r.Float64()*0.3)
		base := len(g.Components())

		var wantNodes []NodeID
		for _, n := range g.Nodes() {
			c := g.Clone()
			c.RemoveNode(n)
			// removing a node also removes it from the count; compare to base minus 1 if it was isolated
			after := len(c.Components())
			isolated := len(g.Neighbors(n)) == 0
			expected := base
			if isolated {
				expected = base - 1
			}
			if after > expected {
				wantNodes = append(wantNodes, n)
			}
		}
		var wantBridges []Link
		for _, l := range g.Links() {
			c := g.Clone()
			c.RemoveLink(l.A, l.B)
			if len(c.Components()) > base {
				wantBridges = append(wantBridges, l)
			}
		}
		gotNodes, gotBridges := g.SinglePointsOfFailure()
		if !reflect.DeepEqual(nilToEmpty(gotNodes), nilToEmpty(wantNodes)) {
			t.Fatalf("trial %d nodes: got %v want %v\nlinks=%v", trial, gotNodes, wantNodes, g.Links())
		}
		if len(gotBridges) != len(wantBridges) {
			t.Fatalf("trial %d bridges: got %v want %v\nlinks=%v", trial, gotBridges, wantBridges, g.Links())
		}
		for i := range gotBridges {
			if gotBridges[i] != wantBridges[i] {
				t.Fatalf("trial %d bridge %d: got %v want %v", trial, i, gotBridges[i], wantBridges[i])
			}
		}
	}
}

func nilToEmpty(s []NodeID) []NodeID {
	if s == nil {
		return []NodeID{}
	}
	return s
}

// Dijkstra must match Floyd–Warshall on every reachable pair.
func TestShortestPathMatchesFloydWarshall(t *testing.T) {
	r := rand.New(rand.NewSource(7))
	const inf = 1e18
	for trial := 0; trial < 100; trial++ {
		g := randomGraph(r, 2+r.Intn(9), 0.35)
		nodes := g.Nodes()
		idx := map[NodeID]int{}
		for i, n := range nodes {
			idx[n] = i
		}
		d := make([][]float64, len(nodes))
		for i := range d {
			d[i] = make([]float64, len(nodes))
			for j := range d[i] {
				if i != j {
					d[i][j] = inf
				}
			}
		}
		for _, l := range g.Links() {
			d[idx[l.A]][idx[l.B]], d[idx[l.B]][idx[l.A]] = l.Cost, l.Cost
		}
		for k := range nodes {
			for i := range nodes {
				for j := range nodes {
					if d[i][k]+d[k][j] < d[i][j] {
						d[i][j] = d[i][k] + d[k][j]
					}
				}
			}
		}
		for i, a := range nodes {
			for j, b := range nodes {
				p, err := g.ShortestPath(a, b)
				if d[i][j] >= inf {
					if !errors.Is(err, ErrNoPath) {
						t.Fatalf("trial %d %s->%s: want ErrNoPath got %v", trial, a, b, err)
					}
					continue
				}
				if err != nil || p.Cost != d[i][j] {
					t.Fatalf("trial %d %s->%s: got %+v %v want %v", trial, a, b, p, err, d[i][j])
				}
				// the returned path must really be a path with that cost
				sum := 0.0
				for k := 0; k+1 < len(p.Nodes); k++ {
					c, ok := g.adj[p.Nodes[k]][p.Nodes[k+1]]
					if !ok {
						t.Fatalf("path uses missing link %v", p.Nodes[k:k+2])
					}
					sum += c
				}
				if sum != p.Cost {
					t.Fatalf("path cost mismatch %v vs %v", sum, p.Cost)
				}
			}
		}
	}
}

func TestDeepChainDoesNotOverflowStack(t *testing.T) {
	g := New()
	const n = 200000
	for i := 0; i < n-1; i++ {
		_ = g.AddLink(NodeID(fmt.Sprintf("n%07d", i)), NodeID(fmt.Sprintf("n%07d", i+1)), 1)
	}
	nodes, bridges := g.SinglePointsOfFailure()
	if len(bridges) != n-1 || len(nodes) != n-2 {
		t.Fatalf("chain: %d cut nodes, %d bridges", len(nodes), len(bridges))
	}
}

func BenchmarkShortestPathGrid100x100(b *testing.B) {
	g := New()
	id := func(x, y int) NodeID { return NodeID(fmt.Sprintf("%d,%d", x, y)) }
	for x := 0; x < 100; x++ {
		for y := 0; y < 100; y++ {
			if x+1 < 100 {
				_ = g.AddLink(id(x, y), id(x+1, y), 1)
			}
			if y+1 < 100 {
				_ = g.AddLink(id(x, y), id(x, y+1), 1)
			}
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := g.ShortestPath(id(0, 0), id(99, 99)); err != nil {
			b.Fatal(err)
		}
	}
}
