package cfgdiff

import (
	"reflect"
	"strings"
	"testing"
)

func cmds(t *testing.T, desired, running string) []string {
	t.Helper()
	ch, err := Diff(desired, running)
	if err != nil {
		t.Fatal(err)
	}
	return Commands(ch)
}

func TestIdenticalConfigsProduceNoChanges(t *testing.T) {
	cfg := "hostname r1\n!\ninterface Gi0/1\n description uplink\n ip address 10.0.0.1 255.255.255.0\n"
	if got := cmds(t, cfg, cfg); len(got) != 0 {
		t.Fatalf("expected no commands, got %v", got)
	}
}

func TestWhitespaceAndCommentsAreIgnored(t *testing.T) {
	a := "hostname   r1\n! a comment\n\ninterface Gi0/1\n   description   up\n"
	b := "hostname r1\ninterface Gi0/1\n description up\n"
	if got := cmds(t, a, b); len(got) != 0 {
		t.Fatalf("got %v", got)
	}
}

func TestChangedLineInsideBlock(t *testing.T) {
	desired := "interface Gi0/1\n description NEW\n ip address 10.0.0.1 255.255.255.0\n"
	running := "interface Gi0/1\n description OLD\n ip address 10.0.0.1 255.255.255.0\n"
	want := []string{"interface Gi0/1", " no description OLD", " description NEW"}
	if got := cmds(t, desired, running); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestNewBlockIsAddedWithAllChildren(t *testing.T) {
	desired := "hostname r1\ninterface Gi0/2\n description new\n shutdown\n"
	running := "hostname r1\n"
	want := []string{"interface Gi0/2", " description new", " shutdown"}
	if got := cmds(t, desired, running); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestStaleBlockIsNegatedWithoutChildren(t *testing.T) {
	desired := "hostname r1\n"
	running := "hostname r1\ninterface Loopback9\n ip address 9.9.9.9 255.255.255.255\n"
	want := []string{"no interface Loopback9"}
	if got := cmds(t, desired, running); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestNegatedRunningLineIsReenabledByStrippingNo(t *testing.T) {
	desired := "interface Gi0/1\n"
	running := "interface Gi0/1\n no ip redirects\n"
	want := []string{"interface Gi0/1", " ip redirects"}
	if got := cmds(t, desired, running); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestNestedBlocks(t *testing.T) {
	desired := "router ospf 1\n network 10.0.0.0 0.0.0.255 area 0\n area 1 stub\n"
	running := "router ospf 1\n network 10.0.0.0 0.0.0.255 area 0\n passive-interface default\n"
	want := []string{"router ospf 1", " no passive-interface default", " area 1 stub"}
	if got := cmds(t, desired, running); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestContextHeaderEmittedOncePerBlock(t *testing.T) {
	desired := "interface Gi0/1\n description a\n mtu 9000\n"
	running := "interface Gi0/1\n"
	got := cmds(t, desired, running)
	if n := strings.Count(strings.Join(got, "\n"), "interface Gi0/1"); n != 1 {
		t.Fatalf("header emitted %d times: %q", n, got)
	}
}

// Applying the generated commands to running must yield desired (idempotence).
func TestRoundTrip(t *testing.T) {
	desired := "hostname r1\ninterface Gi0/1\n description uplink\n mtu 9000\nrouter bgp 65000\n neighbor 10.0.0.2 remote-as 65001\n"
	running := "hostname old\ninterface Gi0/1\n description x\ninterface Gi0/9\n shutdown\n"
	ch, err := Diff(desired, running)
	if err != nil {
		t.Fatal(err)
	}
	if len(ch) == 0 {
		t.Fatal("expected changes")
	}
	// naive applier: rebuild running tree from commands
	applied := apply(t, running, Commands(ch))
	again, err := Diff(desired, applied)
	if err != nil {
		t.Fatal(err)
	}
	if len(again) != 0 {
		t.Fatalf("not converged, remaining: %v\napplied:\n%s", Commands(again), applied)
	}
}

func TestParseRejectsTabs(t *testing.T) {
	if _, err := Parse("interface Gi0/1\n\tdescription x\n"); err == nil {
		t.Fatal("expected tab error")
	}
}

// apply is a tiny IOS-like interpreter used only to prove convergence.
func apply(t *testing.T, running string, commands []string) string {
	t.Helper()
	root, err := Parse(running)
	if err != nil {
		t.Fatal(err)
	}
	var ctx []*Node
	cur := func() *Node {
		if len(ctx) == 0 {
			return root
		}
		return ctx[len(ctx)-1]
	}
	find := func(parent *Node, line string) (*Node, int) {
		for i, c := range parent.Children {
			if c.Line == line {
				return c, i
			}
		}
		return nil, -1
	}
	for _, raw := range commands {
		depth := len(raw) - len(strings.TrimLeft(raw, " "))
		line := normalize(raw)
		ctx = ctx[:depth]
		parent := cur()
		if strings.HasPrefix(line, "no ") {
			target := strings.TrimPrefix(line, "no ")
			if _, i := find(parent, target); i >= 0 {
				parent.Children = append(parent.Children[:i], parent.Children[i+1:]...)
			}
			continue
		}
		n, _ := find(parent, line)
		if n == nil {
			n = &Node{Line: line}
			parent.Children = append(parent.Children, n)
		}
		ctx = append(ctx, n)
	}
	var sb strings.Builder
	var dump func(n *Node, d int)
	dump = func(n *Node, d int) {
		for _, c := range n.Children {
			sb.WriteString(strings.Repeat(" ", d) + c.Line + "\n")
			dump(c, d+1)
		}
	}
	dump(root, 0)
	return sb.String()
}
