// Package topology models a network as an undirected weighted graph and
// answers the questions an orchestrator needs before it touches a device:
// "what is the best path?", "what breaks if this node dies?", "is the network
// still one piece?".
//
// It depends only on the standard library so it can be reused by every
// service in this repository and tested in isolation.
package topology

import (
	"container/heap"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"sort"
)

// NodeID identifies a device (router, switch, host...).
type NodeID string

// Link is an undirected, weighted connection between two nodes.
type Link struct {
	A    NodeID  `json:"a"`
	B    NodeID  `json:"b"`
	Cost float64 `json:"cost"`
}

// Path is an ordered list of nodes from source to destination and its total cost.
type Path struct {
	Nodes []NodeID `json:"nodes"`
	Cost  float64  `json:"cost"`
}

var (
	// ErrUnknownNode is returned when a node is not part of the graph.
	ErrUnknownNode = errors.New("topology: unknown node")
	// ErrNoPath is returned when the destination cannot be reached.
	ErrNoPath = errors.New("topology: no path")
)

// Graph is an undirected weighted graph. It is NOT safe for concurrent
// mutation; wrap it with your own lock if shared between goroutines.
type Graph struct {
	adj map[NodeID]map[NodeID]float64
}

// New returns an empty graph.
func New() *Graph { return &Graph{adj: make(map[NodeID]map[NodeID]float64)} }

// AddNode registers a node. Adding an existing node is a no-op.
func (g *Graph) AddNode(id NodeID) {
	if _, ok := g.adj[id]; !ok {
		g.adj[id] = make(map[NodeID]float64)
	}
}

// AddLink connects a and b (creating them if needed). Re-adding an existing
// link updates its cost. Cost must be positive and finite; self-links are rejected.
func (g *Graph) AddLink(a, b NodeID, cost float64) error {
	switch {
	case a == b:
		return fmt.Errorf("topology: self-link on %q", a)
	case a == "" || b == "":
		return errors.New("topology: empty node id")
	case !(cost > 0) || math.IsInf(cost, 0):
		return fmt.Errorf("topology: link %s-%s needs a positive finite cost, got %v", a, b, cost)
	}
	g.AddNode(a)
	g.AddNode(b)
	g.adj[a][b] = cost
	g.adj[b][a] = cost
	return nil
}

// RemoveLink deletes the link between a and b if present.
func (g *Graph) RemoveLink(a, b NodeID) {
	delete(g.adj[a], b)
	delete(g.adj[b], a)
}

// RemoveNode deletes a node and every link attached to it.
func (g *Graph) RemoveNode(id NodeID) {
	for peer := range g.adj[id] {
		delete(g.adj[peer], id)
	}
	delete(g.adj, id)
}

// HasNode reports whether id is in the graph.
func (g *Graph) HasNode(id NodeID) bool { _, ok := g.adj[id]; return ok }

// Nodes returns all node IDs sorted for deterministic output.
func (g *Graph) Nodes() []NodeID {
	out := make([]NodeID, 0, len(g.adj))
	for id := range g.adj {
		out = append(out, id)
	}
	sortIDs(out)
	return out
}

// Neighbors returns the sorted peers of id.
func (g *Graph) Neighbors(id NodeID) []NodeID {
	out := make([]NodeID, 0, len(g.adj[id]))
	for p := range g.adj[id] {
		out = append(out, p)
	}
	sortIDs(out)
	return out
}

// Links returns every link once (A < B), sorted.
func (g *Graph) Links() []Link {
	var out []Link
	for a, peers := range g.adj {
		for b, c := range peers {
			if a < b {
				out = append(out, Link{A: a, B: b, Cost: c})
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].A != out[j].A {
			return out[i].A < out[j].A
		}
		return out[i].B < out[j].B
	})
	return out
}

// Clone returns a deep copy.
func (g *Graph) Clone() *Graph {
	c := New()
	for a, peers := range g.adj {
		c.AddNode(a)
		for b, cost := range peers {
			c.adj[a][b] = cost
		}
	}
	return c
}

// FromJSON builds a graph from {"nodes": ["r1"], "links": [{"a":"r1","b":"r2","cost":10}]}.
// "nodes" is optional and only needed for isolated devices.
func FromJSON(r io.Reader) (*Graph, error) {
	var doc struct {
		Nodes []NodeID `json:"nodes"`
		Links []Link   `json:"links"`
	}
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&doc); err != nil {
		return nil, fmt.Errorf("topology: decode: %w", err)
	}
	g := New()
	for _, n := range doc.Nodes {
		if n == "" {
			return nil, errors.New("topology: empty node id")
		}
		g.AddNode(n)
	}
	for _, l := range doc.Links {
		if err := g.AddLink(l.A, l.B, l.Cost); err != nil {
			return nil, err
		}
	}
	return g, nil
}

// ---------------------------------------------------------------------------
// Shortest path (Dijkstra)
// ---------------------------------------------------------------------------

type item struct {
	id   NodeID
	dist float64
}
type minHeap []item

func (h minHeap) Len() int { return len(h) }
func (h minHeap) Less(i, j int) bool {
	if h[i].dist != h[j].dist {
		return h[i].dist < h[j].dist
	}
	return h[i].id < h[j].id // deterministic tie-break
}
func (h minHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h *minHeap) Push(x any)   { *h = append(*h, x.(item)) }
func (h *minHeap) Pop() any {
	old := *h
	n := len(old)
	it := old[n-1]
	*h = old[:n-1]
	return it
}

// ShortestPath returns the lowest-cost path from src to dst. Ties are broken
// deterministically (lexicographically smallest predecessor).
func (g *Graph) ShortestPath(src, dst NodeID) (Path, error) {
	if !g.HasNode(src) {
		return Path{}, fmt.Errorf("%w: %q", ErrUnknownNode, src)
	}
	if !g.HasNode(dst) {
		return Path{}, fmt.Errorf("%w: %q", ErrUnknownNode, dst)
	}
	dist := map[NodeID]float64{src: 0}
	prev := map[NodeID]NodeID{}
	done := map[NodeID]bool{}
	h := &minHeap{{src, 0}}
	for h.Len() > 0 {
		cur := heap.Pop(h).(item)
		if done[cur.id] {
			continue
		}
		done[cur.id] = true
		if cur.id == dst {
			break
		}
		for _, p := range g.Neighbors(cur.id) {
			if done[p] {
				continue
			}
			nd := cur.dist + g.adj[cur.id][p]
			old, seen := dist[p]
			if !seen || nd < old || (nd == old && cur.id < prev[p]) {
				dist[p] = nd
				prev[p] = cur.id
				heap.Push(h, item{p, nd})
			}
		}
	}
	if !done[dst] {
		return Path{}, fmt.Errorf("%w: %s -> %s", ErrNoPath, src, dst)
	}
	var rev []NodeID
	for n := dst; ; n = prev[n] {
		rev = append(rev, n)
		if n == src {
			break
		}
	}
	nodes := make([]NodeID, len(rev))
	for i, n := range rev {
		nodes[len(rev)-1-i] = n
	}
	return Path{Nodes: nodes, Cost: dist[dst]}, nil
}

// ---------------------------------------------------------------------------
// Connectivity and failure analysis
// ---------------------------------------------------------------------------

// Components returns the connected components, each sorted, ordered by
// descending size then by first node.
func (g *Graph) Components() [][]NodeID {
	seen := make(map[NodeID]bool, len(g.adj))
	var comps [][]NodeID
	for _, start := range g.Nodes() {
		if seen[start] {
			continue
		}
		var comp []NodeID
		stack := []NodeID{start}
		seen[start] = true
		for len(stack) > 0 {
			n := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			comp = append(comp, n)
			for p := range g.adj[n] {
				if !seen[p] {
					seen[p] = true
					stack = append(stack, p)
				}
			}
		}
		sortIDs(comp)
		comps = append(comps, comp)
	}
	sort.Slice(comps, func(i, j int) bool {
		if len(comps[i]) != len(comps[j]) {
			return len(comps[i]) > len(comps[j])
		}
		return comps[i][0] < comps[j][0]
	})
	return comps
}

// Connected reports whether the whole graph is a single component.
func (g *Graph) Connected() bool { return len(g.adj) <= 1 || len(g.Components()) == 1 }

// SinglePointsOfFailure returns the articulation points (nodes whose loss
// splits the network) and bridges (links whose loss splits the network),
// computed with Tarjan's low-link algorithm in O(V+E). The traversal is
// iterative so very deep topologies cannot overflow the goroutine stack.
func (g *Graph) SinglePointsOfFailure() (nodes []NodeID, bridges []Link) {
	type frame struct {
		node   NodeID
		parent NodeID
		nbrs   []NodeID
		next   int
	}
	disc := map[NodeID]int{}
	low := map[NodeID]int{}
	isCut := map[NodeID]bool{}
	timer := 0

	for _, root := range g.Nodes() {
		if _, ok := disc[root]; ok {
			continue
		}
		rootChildren := 0
		timer++
		disc[root], low[root] = timer, timer
		stack := []*frame{{node: root, nbrs: g.Neighbors(root)}}
		for len(stack) > 0 {
			f := stack[len(stack)-1]
			if f.next < len(f.nbrs) {
				v := f.nbrs[f.next]
				f.next++
				if v == f.parent {
					continue // undirected simple graph: skip the tree edge back
				}
				if _, seen := disc[v]; seen {
					if disc[v] < low[f.node] {
						low[f.node] = disc[v]
					}
					continue
				}
				timer++
				disc[v], low[v] = timer, timer
				if f.node == root {
					rootChildren++
				}
				stack = append(stack, &frame{node: v, parent: f.node, nbrs: g.Neighbors(v)})
				continue
			}
			// finished f: propagate to parent
			stack = stack[:len(stack)-1]
			if len(stack) == 0 {
				break
			}
			p := stack[len(stack)-1]
			if low[f.node] < low[p.node] {
				low[p.node] = low[f.node]
			}
			if low[f.node] > disc[p.node] {
				a, b := p.node, f.node
				if b < a {
					a, b = b, a
				}
				bridges = append(bridges, Link{A: a, B: b, Cost: g.adj[a][b]})
			}
			if p.node != root && low[f.node] >= disc[p.node] {
				isCut[p.node] = true
			}
		}
		if rootChildren > 1 {
			isCut[root] = true
		}
	}
	for n := range isCut {
		nodes = append(nodes, n)
	}
	sortIDs(nodes)
	sort.Slice(bridges, func(i, j int) bool {
		if bridges[i].A != bridges[j].A {
			return bridges[i].A < bridges[j].A
		}
		return bridges[i].B < bridges[j].B
	})
	return nodes, bridges
}

// BlastRadius reports how many *other* nodes lose connectivity to the main
// network if `failed` goes down: the nodes outside the largest surviving
// component. It returns ErrUnknownNode for a node not in the graph.
func (g *Graph) BlastRadius(failed NodeID) (int, error) {
	if !g.HasNode(failed) {
		return 0, fmt.Errorf("%w: %q", ErrUnknownNode, failed)
	}
	c := g.Clone()
	c.RemoveNode(failed)
	comps := c.Components()
	if len(comps) <= 1 {
		return 0, nil
	}
	cut := 0
	for _, comp := range comps[1:] {
		cut += len(comp)
	}
	return cut, nil
}

func sortIDs(ids []NodeID) { sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] }) }
