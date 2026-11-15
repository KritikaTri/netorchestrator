// Package cfgdiff computes the minimal change set between two IOS-style
// hierarchical device configurations and renders it as the CLI commands that
// would turn the running config into the desired one.
//
// Semantics (documented so they can be reviewed, not hidden):
//   - Indentation defines hierarchy; comment lines ("!") and blanks are ignored.
//   - Sibling order is NOT significant (fine for interface/router blocks, wrong
//     for order-sensitive constructs such as ACL sequences or route-maps).
//   - A line present only in "running" is removed with a "no" prefix; if it
//     already starts with "no " the prefix is stripped instead.
//   - A whole block present only in "running" is removed by negating its header.
package cfgdiff

import (
	"fmt"
	"strings"
)

// Node is one config line plus the lines indented beneath it.
type Node struct {
	Line     string
	Children []*Node
}

// Op is the kind of change.
type Op int

const (
	// Add means the line must be configured.
	Add Op = iota + 1
	// Remove means the line must be negated.
	Remove
)

func (o Op) String() string {
	if o == Add {
		return "add"
	}
	return "remove"
}

// Change is a single line-level difference. Path holds the enclosing lines
// followed by the line itself.
type Change struct {
	Op   Op
	Path []string
}

// Line returns the changed line.
func (c Change) Line() string { return c.Path[len(c.Path)-1] }

// Parse builds the hierarchy from config text.
func Parse(text string) (*Node, error) {
	root := &Node{}
	type level struct {
		indent int
		node   *Node
	}
	stack := []level{{indent: -1, node: root}}
	for i, raw := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" || strings.HasPrefix(trimmed, "!") {
			continue
		}
		if strings.ContainsRune(raw[:len(raw)-len(strings.TrimLeft(raw, " \t"))], '\t') {
			return nil, fmt.Errorf("cfgdiff: line %d: tab indentation is not supported", i+1)
		}
		indent := len(raw) - len(strings.TrimLeft(raw, " "))
		for len(stack) > 1 && indent <= stack[len(stack)-1].indent {
			stack = stack[:len(stack)-1]
		}
		parent := stack[len(stack)-1]
		n := &Node{Line: normalize(trimmed)}
		parent.node.Children = append(parent.node.Children, n)
		stack = append(stack, level{indent: indent, node: n})
	}
	return root, nil
}

func normalize(s string) string { return strings.Join(strings.Fields(s), " ") }

// Diff returns the changes needed to go from running to desired.
func Diff(desired, running string) ([]Change, error) {
	d, err := Parse(desired)
	if err != nil {
		return nil, fmt.Errorf("desired: %w", err)
	}
	r, err := Parse(running)
	if err != nil {
		return nil, fmt.Errorf("running: %w", err)
	}
	return diffNodes(d, r, nil), nil
}

func index(nodes []*Node) map[string]*Node {
	m := make(map[string]*Node, len(nodes))
	for _, n := range nodes {
		if _, dup := m[n.Line]; !dup {
			m[n.Line] = n
		}
	}
	return m
}

func diffNodes(want, have *Node, parents []string) []Change {
	var out []Change
	haveIdx, wantIdx := index(have.Children), index(want.Children)

	// Removals first (negate stale lines), then additions, each in file order.
	for _, h := range have.Children {
		if _, ok := wantIdx[h.Line]; !ok {
			out = append(out, Change{Op: Remove, Path: join(parents, h.Line)})
		}
	}
	for _, w := range want.Children {
		h, ok := haveIdx[w.Line]
		if !ok {
			out = append(out, addSubtree(w, parents)...)
			continue
		}
		if len(w.Children) > 0 || len(h.Children) > 0 {
			out = append(out, diffNodes(w, h, join(parents, w.Line))...)
		}
	}
	return out
}

func addSubtree(n *Node, parents []string) []Change {
	p := join(parents, n.Line)
	out := []Change{{Op: Add, Path: p}}
	for _, c := range n.Children {
		out = append(out, addSubtree(c, p)...)
	}
	return out
}

func join(parents []string, line string) []string {
	p := make([]string, len(parents)+1)
	copy(p, parents)
	p[len(parents)] = line
	return p
}

// Commands renders changes as an indented command list, emitting the parent
// context lines needed to reach each change exactly once.
func Commands(changes []Change) []string {
	var out []string
	var open []string // context currently entered
	emit := func(depth int, s string) { out = append(out, strings.Repeat(" ", depth)+s) }

	for _, c := range changes {
		parents := c.Path[:len(c.Path)-1]
		// keep the longest shared prefix of the open context
		keep := 0
		for keep < len(open) && keep < len(parents) && open[keep] == parents[keep] {
			keep++
		}
		open = open[:keep]
		for i := keep; i < len(parents); i++ {
			emit(i, parents[i])
			open = append(open, parents[i])
		}
		line := c.Line()
		if c.Op == Remove {
			if strings.HasPrefix(line, "no ") {
				line = strings.TrimPrefix(line, "no ")
			} else {
				line = "no " + line
			}
			emit(len(parents), line)
			continue
		}
		emit(len(parents), line)
		open = append(open, line)
	}
	return out
}
