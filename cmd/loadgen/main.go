// Command loadgen drives a NetOrchestrator endpoint with synthetic topologies
// and reports throughput and latency percentiles. It exists so every
// performance claim in the docs can be reproduced with one command.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"netorchestrator/internal/platform"
)

type node struct {
	ID string `json:"id"`
}
type link struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
}

// topology builds a connected sparse graph: a ring (guarantees connectivity)
// plus ~n/2 random chords, resembling a campus/DC fabric. Seeded => reproducible.
func topology(n int, seed int64) ([]node, []link) {
	r := rand.New(rand.NewSource(seed))
	nodes := make([]node, n)
	for i := range nodes {
		nodes[i] = node{fmt.Sprintf("n%d", i)}
	}
	var links []link
	add := func(a, b int) {
		links = append(links, link{fmt.Sprintf("l%d", len(links)), nodes[a].ID, nodes[b].ID})
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

func main() {
	target := flag.String("url", "http://localhost:8080", "base URL (gateway or service)")
	endpoint := flag.String("endpoint", "/v1/analyze/resilience", "endpoint path")
	secret := flag.String("jwt-secret", os.Getenv("GATEWAY_JWT_SECRET"), "mint a bearer token with this HS256 secret (optional)")
	nodes := flag.Int("nodes", 50, "topology size")
	variants := flag.Int("variants", 1, "distinct topologies to rotate through (1 = all requests identical => cache hits)")
	conc := flag.Int("c", 32, "concurrent clients")
	dur := flag.Duration("d", 10*time.Second, "test duration")
	flag.Parse()

	bodies := make([][]byte, *variants)
	for v := range bodies {
		ns, ls := topology(*nodes, int64(v+1))
		req := map[string]any{"nodes": ns, "links": ls}
		if *endpoint == "/v1/analyze/preflight" {
			req["change"] = map[string]string{"kind": "remove_node", "id": "n0"}
		}
		bodies[v], _ = json.Marshal(req)
	}
	auth := ""
	if *secret != "" {
		auth = "Bearer " + platform.SignHS256([]byte(*secret), platform.Claims{Subject: "loadgen", Expires: time.Now().Add(time.Hour).Unix()})
	}
	client := &http.Client{Transport: &http.Transport{MaxIdleConnsPerHost: *conc * 2}, Timeout: 30 * time.Second}

	var (
		mu       sync.Mutex
		lats     []float64
		codes    = map[int]int{}
		hits     atomic.Int64
		errs     atomic.Int64
		counter  atomic.Int64
		deadline = time.Now().Add(*dur)
		wg       sync.WaitGroup
	)
	start := time.Now()
	for i := 0; i < *conc; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			local := make([]float64, 0, 1024)
			lc := map[int]int{}
			for time.Now().Before(deadline) {
				body := bodies[int(counter.Add(1))%len(bodies)]
				req, _ := http.NewRequest("POST", *target+*endpoint, bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				if auth != "" {
					req.Header.Set("Authorization", auth)
				}
				t0 := time.Now()
				resp, err := client.Do(req)
				if err != nil {
					errs.Add(1)
					continue
				}
				_, _ = io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
				local = append(local, time.Since(t0).Seconds()*1000)
				lc[resp.StatusCode]++
				if resp.Header.Get("X-Cache") == "HIT" {
					hits.Add(1)
				}
			}
			mu.Lock()
			lats = append(lats, local...)
			for k, v := range lc {
				codes[k] += v
			}
			mu.Unlock()
		}()
	}
	wg.Wait()
	elapsed := time.Since(start).Seconds()
	sort.Float64s(lats)
	q := func(p float64) float64 {
		if len(lats) == 0 {
			return 0
		}
		return lats[int(float64(len(lats)-1)*p)]
	}
	fmt.Printf("endpoint=%s nodes=%d variants=%d clients=%d duration=%.1fs\n", *endpoint, *nodes, *variants, *conc, elapsed)
	fmt.Printf("requests=%d  rps=%.0f  transport_errors=%d  cache_hits=%d\n", len(lats), float64(len(lats))/elapsed, errs.Load(), hits.Load())
	fmt.Printf("latency_ms  p50=%.2f  p95=%.2f  p99=%.2f  max=%.2f\n", q(.5), q(.95), q(.99), q(1))
	fmt.Printf("status=%v body_bytes=%d\n", codes, len(bodies[0]))
}
