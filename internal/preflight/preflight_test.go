package preflight

import (
	"errors"
	"reflect"
	"testing"
)

func names(ns []Node) []string {
	out := []string{}
	for _, n := range ns {
		out = append(out, n.Name)
	}
	return out
}

var campusNodes = []Node{
	{"1", "core1"}, {"2", "core2"}, {"3", "dist1"}, {"4", "dist2"}, {"5", "access1"}, {"6", "access2"},
}

// core1,core2 meshed with dist1,dist2; access1 hangs off dist1 only; access2 off dist2 only.
var campusLinks = []Link{
	{"l1", "1", "2", 1}, {"l2", "1", "3", 10}, {"l3", "2", "3", 10}, {"l4", "1", "4", 10}, {"l5", "2", "4", 10},
	{"l6", "3", "5", 100}, {"l7", "4", "6", 100},
}

func TestRedundantCoreFailureCutsNothingOff(t *testing.T) {
	r, err := Assess(campusNodes, campusLinks, Change{RemoveNode, "1"})
	if err != nil {
		t.Fatal(err)
	}
	if r.Risk != RiskNone || len(r.CutOff) != 0 || r.Target != "core1" {
		t.Fatalf("%+v", r)
	}
}

func TestSingleHomedAccessSwitchIsCutOffWhenDistributionFails(t *testing.T) {
	r, err := Assess(campusNodes, campusLinks, Change{RemoveNode, "3"})
	if err != nil {
		t.Fatal(err)
	}
	if r.Risk != RiskHigh || !reflect.DeepEqual(names(r.CutOff), []string{"access1"}) {
		t.Fatalf("%+v", r)
	}
	if r.ComponentsBefore != 1 || r.ComponentsAfter != 2 {
		t.Fatalf("components %d -> %d", r.ComponentsBefore, r.ComponentsAfter)
	}
}

func TestRemovingABridgeLinkIsFlaggedByName(t *testing.T) {
	r, err := Assess(campusNodes, campusLinks, Change{RemoveLink, "l7"})
	if err != nil {
		t.Fatal(err)
	}
	if r.Target != "dist2 <-> access2" || !reflect.DeepEqual(names(r.CutOff), []string{"access2"}) {
		t.Fatalf("%+v", r)
	}
}

func TestRemovingARedundantLinkIsSafe(t *testing.T) {
	r, _ := Assess(campusNodes, campusLinks, Change{RemoveLink, "l2"})
	if r.Risk != RiskNone || r.ComponentsAfter != 1 {
		t.Fatalf("%+v", r)
	}
}

func TestLeafRemovalDoesNotCountItselfAsCutOff(t *testing.T) {
	r, _ := Assess(campusNodes, campusLinks, Change{RemoveNode, "5"})
	if r.Risk != RiskNone || len(r.CutOff) != 0 {
		t.Fatalf("%+v", r)
	}
}

func TestHubFailureIsCritical(t *testing.T) {
	nodes := []Node{{"h", "hub"}, {"a", "a"}, {"b", "b"}, {"c", "c"}, {"d", "d"}}
	links := []Link{{"1", "h", "a", 0}, {"2", "h", "b", 0}, {"3", "h", "c", 0}, {"4", "h", "d", 0}}
	r, err := Assess(nodes, links, Change{RemoveNode, "h"})
	if err != nil {
		t.Fatal(err)
	}
	if r.Risk != RiskCritical || len(r.CutOff) != 4 {
		t.Fatalf("%+v", r)
	}
}

func TestTiedFragmentsAreAllCutOff(t *testing.T) {
	// a-h-b: removing h leaves two equal halves; neither is "the" main network.
	nodes := []Node{{"h", "hub"}, {"a", "a"}, {"b", "b"}}
	links := []Link{{"1", "h", "a", 0}, {"2", "h", "b", 0}}
	r, _ := Assess(nodes, links, Change{RemoveNode, "h"})
	if !reflect.DeepEqual(names(r.CutOff), []string{"a", "b"}) || r.Risk != RiskCritical {
		t.Fatalf("%+v", r)
	}
}

func TestAlreadyPartitionedNetworkIsMeasuredAgainstLargestPart(t *testing.T) {
	nodes := []Node{{"a", "a"}, {"b", "b"}, {"c", "c"}, {"x", "x"}, {"y", "y"}}
	links := []Link{{"1", "a", "b", 0}, {"2", "b", "c", 0}, {"3", "x", "y", 0}}
	r, _ := Assess(nodes, links, Change{RemoveNode, "b"})
	if !r.AlreadyPartitioned || !reflect.DeepEqual(names(r.CutOff), []string{"a", "c"}) {
		t.Fatalf("%+v", r)
	}
	// the already-isolated x,y must not be blamed on this change
	r2, _ := Assess(nodes, links, Change{RemoveLink, "3"})
	if len(r2.CutOff) != 0 || r2.Risk != RiskNone {
		t.Fatalf("%+v", r2)
	}
}

func TestErrors(t *testing.T) {
	if _, err := Assess(campusNodes, campusLinks, Change{RemoveNode, "nope"}); !errors.Is(err, ErrUnknownTarget) {
		t.Errorf("unknown node: %v", err)
	}
	if _, err := Assess(campusNodes, campusLinks, Change{RemoveLink, "nope"}); !errors.Is(err, ErrUnknownTarget) {
		t.Errorf("unknown link: %v", err)
	}
	if _, err := Assess(campusNodes, campusLinks, Change{"explode", "1"}); !errors.Is(err, ErrInvalidChange) {
		t.Errorf("bad kind: %v", err)
	}
	if _, err := Assess(campusNodes, []Link{{"x", "1", "ghost", 0}}, Change{RemoveNode, "1"}); !errors.Is(err, ErrInvalidTopology) {
		t.Errorf("dangling link: %v", err)
	}
	if _, err := Assess([]Node{{"1", "a"}, {"1", "dup"}}, nil, Change{RemoveNode, "1"}); !errors.Is(err, ErrInvalidTopology) {
		t.Errorf("duplicate id: %v", err)
	}
	if _, err := Assess([]Node{{"1", "a"}}, []Link{{"l", "1", "1", 0}}, Change{RemoveNode, "1"}); !errors.Is(err, ErrInvalidTopology) {
		t.Errorf("self link: %v", err)
	}
}

func TestNameFallsBackToID(t *testing.T) {
	r, _ := Assess([]Node{{"only", ""}}, nil, Change{RemoveNode, "only"})
	if r.Target != "only" {
		t.Fatalf("%+v", r)
	}
}

func TestAnalyze(t *testing.T) {
	res, err := Analyze(campusNodes, campusLinks)
	if err != nil {
		t.Fatal(err)
	}
	if res.Redundant || res.Components != 1 || res.Nodes != 6 || res.Links != 7 {
		t.Fatalf("%+v", res)
	}
	if !reflect.DeepEqual(names(res.ArticulationPoints), []string{"dist1", "dist2"}) {
		t.Errorf("aps: %v", names(res.ArticulationPoints))
	}
	if !reflect.DeepEqual(res.Bridges, []string{"dist1 <-> access1", "dist2 <-> access2"}) {
		t.Errorf("bridges: %v", res.Bridges)
	}
}

func TestAnalyzeRing(t *testing.T) {
	nodes := []Node{{"a", "a"}, {"b", "b"}, {"c", "c"}}
	links := []Link{{"1", "a", "b", 0}, {"2", "b", "c", 0}, {"3", "c", "a", 0}}
	res, _ := Analyze(nodes, links)
	if !res.Redundant || len(res.ArticulationPoints) != 0 || len(res.Bridges) != 0 {
		t.Fatalf("%+v", res)
	}
}
