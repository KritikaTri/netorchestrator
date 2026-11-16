package preflight

import (
	"fmt"
	"math/rand"
	"testing"
)

// ring + n/2 random chords, seeded; same shape the load generator uses.
func benchTopology(n int) ([]Node, []Link) {
	r := rand.New(rand.NewSource(1))
	nodes := make([]Node, n)
	for i := range nodes {
		nodes[i] = Node{ID: fmt.Sprintf("n%d", i)}
	}
	var links []Link
	add := func(a, b int) {
		links = append(links, Link{ID: fmt.Sprintf("l%d", len(links)), Source: nodes[a].ID, Target: nodes[b].ID})
	}
	for i := 0; i < n; i++ {
		add(i, (i+1)%n)
	}
	seen := map[[2]int]bool{}
	for len(links) < n+n/2 {
		a, b := r.Intn(n), r.Intn(n)
		if a == b || seen[[2]int{a, b}] || seen[[2]int{b, a}] {
			continue
		}
		seen[[2]int{a, b}] = true
		add(a, b)
	}
	return nodes, links
}

func BenchmarkAssess(b *testing.B) {
	for _, n := range []int{100, 1000, 10000} {
		nodes, links := benchTopology(n)
		b.Run(fmt.Sprintf("nodes=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if _, err := Assess(nodes, links, Change{RemoveNode, "n0"}); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkAnalyze(b *testing.B) {
	for _, n := range []int{100, 1000, 10000} {
		nodes, links := benchTopology(n)
		b.Run(fmt.Sprintf("nodes=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if _, err := Analyze(nodes, links); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
