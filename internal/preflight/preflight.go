// Package preflight answers "what happens if I take this device or link out of
// service?" before an operator or workflow does it. It is a thin, pure layer
// over internal/topology: callers map their own models into the small Node and
// Link types here, so the logic is testable without a database.
package preflight

import (
	"errors"
	"fmt"
	"sort"

	"netorchestrator/internal/topology"
)

// Node is a device in the network.
type Node struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Link is an undirected connection between two node IDs. A zero Cost counts as 1.
type Link struct {
	ID     string  `json:"id"`
	Source string  `json:"source"`
	Target string  `json:"target"`
	Cost   float64 `json:"cost,omitempty"`
}

// ChangeKind is the type of planned change.
type ChangeKind string

const (
	// RemoveNode simulates a device going down (maintenance, failure, decommission).
	RemoveNode ChangeKind = "remove_node"
	// RemoveLink simulates a link going down.
	RemoveLink ChangeKind = "remove_link"
)

// Change is the planned modification to evaluate.
type Change struct {
	Kind ChangeKind `json:"kind"`
	ID   string     `json:"id"`
}

// Risk grades the impact of a change.
type Risk string

const (
	// RiskNone: every device that was reachable stays reachable.
	RiskNone Risk = "none"
	// RiskHigh: some devices lose connectivity to the main network.
	RiskHigh Risk = "high"
	// RiskCritical: half or more of the devices lose connectivity.
	RiskCritical Risk = "critical"
)

var (
	// ErrUnknownTarget means the node or link in the change is not in the topology.
	ErrUnknownTarget = errors.New("preflight: unknown change target")
	// ErrInvalidChange means the change kind is not supported.
	ErrInvalidChange = errors.New("preflight: unsupported change kind")
	// ErrInvalidTopology means the supplied nodes/links are inconsistent.
	ErrInvalidTopology = errors.New("preflight: invalid topology")
)

// Report is the result of assessing one change.
type Report struct {
	Change             Change   `json:"change"`
	Target             string   `json:"target"`
	Risk               Risk     `json:"risk"`
	CutOff             []Node   `json:"cut_off"`
	ComponentsBefore   int      `json:"components_before"`
	ComponentsAfter    int      `json:"components_after"`
	AlreadyPartitioned bool     `json:"already_partitioned"`
	Warnings           []string `json:"warnings"`
}

// Resilience summarises the weak points of the current topology.
type Resilience struct {
	Nodes              int      `json:"nodes"`
	Links              int      `json:"links"`
	Components         int      `json:"components"`
	ArticulationPoints []Node   `json:"articulation_points"`
	Bridges            []string `json:"bridges"`
	Redundant          bool     `json:"redundant"`
}

type model struct {
	g       *topology.Graph
	names   map[string]string // node id -> name
	linkEnd map[string][2]string
}

func build(nodes []Node, links []Link) (*model, error) {
	m := &model{g: topology.New(), names: map[string]string{}, linkEnd: map[string][2]string{}}
	for _, n := range nodes {
		if n.ID == "" {
			return nil, fmt.Errorf("%w: node with empty id", ErrInvalidTopology)
		}
		if _, dup := m.names[n.ID]; dup {
			return nil, fmt.Errorf("%w: duplicate node id %q", ErrInvalidTopology, n.ID)
		}
		name := n.Name
		if name == "" {
			name = n.ID
		}
		m.names[n.ID] = name
		m.g.AddNode(topology.NodeID(n.ID))
	}
	for _, l := range links {
		for _, end := range []string{l.Source, l.Target} {
			if _, ok := m.names[end]; !ok {
				return nil, fmt.Errorf("%w: link %q references unknown node %q", ErrInvalidTopology, l.ID, end)
			}
		}
		cost := l.Cost
		if cost == 0 {
			cost = 1
		}
		if err := m.g.AddLink(topology.NodeID(l.Source), topology.NodeID(l.Target), cost); err != nil {
			return nil, fmt.Errorf("%w: link %q: %v", ErrInvalidTopology, l.ID, err)
		}
		if l.ID != "" {
			m.linkEnd[l.ID] = [2]string{l.Source, l.Target}
		}
	}
	return m, nil
}

func (m *model) label(a, b string) string { return m.names[a] + " <-> " + m.names[b] }

func (m *model) nodes(ids []topology.NodeID) []Node {
	out := make([]Node, 0, len(ids))
	for _, id := range ids {
		out = append(out, Node{ID: string(id), Name: m.names[string(id)]})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Name != out[j].Name {
			return out[i].Name < out[j].Name
		}
		return out[i].ID < out[j].ID
	})
	return out
}

// Assess evaluates a single planned change against the topology.
func Assess(nodes []Node, links []Link, change Change) (*Report, error) {
	m, err := build(nodes, links)
	if err != nil {
		return nil, err
	}
	before := m.g.Components()
	rep := &Report{
		Change:             change,
		ComponentsBefore:   len(before),
		AlreadyPartitioned: len(before) > 1,
		CutOff:             []Node{},
		Warnings:           []string{},
	}
	if rep.AlreadyPartitioned {
		rep.Warnings = append(rep.Warnings,
			fmt.Sprintf("network is already split into %d disconnected parts; impact is measured against the largest part", len(before)))
	}

	after := m.g.Clone()
	removed := ""
	switch change.Kind {
	case RemoveNode:
		if _, ok := m.names[change.ID]; !ok {
			return nil, fmt.Errorf("%w: node %q", ErrUnknownTarget, change.ID)
		}
		rep.Target = m.names[change.ID]
		removed = change.ID
		after.RemoveNode(topology.NodeID(change.ID))
	case RemoveLink:
		ends, ok := m.linkEnd[change.ID]
		if !ok {
			return nil, fmt.Errorf("%w: link %q", ErrUnknownTarget, change.ID)
		}
		rep.Target = m.label(ends[0], ends[1])
		after.RemoveLink(topology.NodeID(ends[0]), topology.NodeID(ends[1]))
	default:
		return nil, fmt.Errorf("%w: %q", ErrInvalidChange, change.Kind)
	}

	afterComps := after.Components()
	rep.ComponentsAfter = len(afterComps)

	// "Main network" = the largest component before the change. After the
	// change, its surviving devices may split into fragments; the largest
	// fragment keeps the main network and every other fragment is cut off.
	// If the largest fragment is not unique, no fragment can claim to be the
	// main network, so every device in the tied fragments counts as cut off.
	mainBefore := map[topology.NodeID]bool{}
	if len(before) > 0 {
		for _, id := range before[0] {
			if string(id) != removed {
				mainBefore[id] = true
			}
		}
	}
	var frags [][]topology.NodeID
	for _, comp := range afterComps {
		var f []topology.NodeID
		for _, id := range comp {
			if mainBefore[id] {
				f = append(f, id)
			}
		}
		if len(f) > 0 {
			frags = append(frags, f)
		}
	}
	biggest, ties := 0, 0
	for _, f := range frags {
		switch {
		case len(f) > biggest:
			biggest, ties = len(f), 1
		case len(f) == biggest:
			ties++
		}
	}
	var cut []topology.NodeID
	for _, f := range frags {
		if len(f) == biggest && ties == 1 {
			continue // the unique largest fragment stays the main network
		}
		cut = append(cut, f...)
	}
	rep.CutOff = m.nodes(cut)

	switch remaining := len(nodes); {
	case len(cut) == 0:
		rep.Risk = RiskNone
	case remaining > 0 && len(cut)*2 >= remaining:
		rep.Risk = RiskCritical
	default:
		rep.Risk = RiskHigh
	}
	if rep.Risk != RiskNone {
		rep.Warnings = append(rep.Warnings,
			fmt.Sprintf("%d device(s) would lose connectivity to the main network", len(cut)))
	}
	return rep, nil
}

// Analyze lists the single points of failure in the current topology.
func Analyze(nodes []Node, links []Link) (*Resilience, error) {
	m, err := build(nodes, links)
	if err != nil {
		return nil, err
	}
	aps, bridges := m.g.SinglePointsOfFailure()
	res := &Resilience{
		Nodes:              len(m.names),
		Links:              len(m.g.Links()),
		Components:         len(m.g.Components()),
		ArticulationPoints: m.nodes(aps),
		Bridges:            []string{},
	}
	for _, b := range bridges {
		res.Bridges = append(res.Bridges, m.label(string(b.A), string(b.B)))
	}
	sort.Strings(res.Bridges)
	res.Redundant = len(aps) == 0 && len(bridges) == 0 && res.Components <= 1
	return res, nil
}

// BuildGraph validates nodes/links and returns the topology graph, for callers
// that need other analyses (e.g. shortest path) over the same input model.
func BuildGraph(nodes []Node, links []Link) (*topology.Graph, error) {
	m, err := build(nodes, links)
	if err != nil {
		return nil, err
	}
	return m.g, nil
}
