// Command netorchctl is a small, dependency-free CLI over the orchestrator's
// analysis libraries. It works offline on files, so it is safe to run in CI
// or a change-review pipeline before anything touches a device.
//
//	netorchctl path   -topology net.json -from r1 -to r9
//	netorchctl spof   -topology net.json
//	netorchctl blast  -topology net.json -node core1
//	netorchctl cfgdiff -desired desired.cfg -running running.cfg
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"netorchestrator/internal/cfgdiff"
	"netorchestrator/internal/topology"
)

const usage = `usage: netorchctl <command> [flags]

commands:
  path     shortest path between two nodes
  spof     articulation points and bridges (single points of failure)
  blast    nodes cut off from the main network if one node fails
  cfgdiff  commands needed to turn a running config into the desired one
`

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return 2
	}
	var err error
	switch args[0] {
	case "path":
		err = cmdPath(args[1:], stdout)
	case "spof":
		err = cmdSPOF(args[1:], stdout)
	case "blast":
		err = cmdBlast(args[1:], stdout)
	case "cfgdiff":
		err = cmdCfgDiff(args[1:], stdout)
	case "-h", "--help", "help":
		fmt.Fprint(stdout, usage)
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command %q\n\n%s", args[0], usage)
		return 2
	}
	if err != nil {
		var ue usageError
		if errors.As(err, &ue) {
			fmt.Fprintln(stderr, err)
			return 2
		}
		fmt.Fprintln(stderr, "error:", err)
		return 1
	}
	return 0
}

type usageError string

func (e usageError) Error() string { return string(e) }

func newFlags(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	return fs
}

func loadGraph(path string) (*topology.Graph, error) {
	if path == "" {
		return nil, usageError("-topology is required")
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return topology.FromJSON(f)
}

func printJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func cmdPath(args []string, out io.Writer) error {
	fs := newFlags("path")
	file := fs.String("topology", "", "topology JSON file")
	from := fs.String("from", "", "source node")
	to := fs.String("to", "", "destination node")
	if err := fs.Parse(args); err != nil {
		return usageError(err.Error())
	}
	if *from == "" || *to == "" {
		return usageError("-from and -to are required")
	}
	g, err := loadGraph(*file)
	if err != nil {
		return err
	}
	p, err := g.ShortestPath(topology.NodeID(*from), topology.NodeID(*to))
	if err != nil {
		return err
	}
	return printJSON(out, p)
}

func cmdSPOF(args []string, out io.Writer) error {
	fs := newFlags("spof")
	file := fs.String("topology", "", "topology JSON file")
	if err := fs.Parse(args); err != nil {
		return usageError(err.Error())
	}
	g, err := loadGraph(*file)
	if err != nil {
		return err
	}
	nodes, bridges := g.SinglePointsOfFailure()
	if nodes == nil {
		nodes = []topology.NodeID{}
	}
	if bridges == nil {
		bridges = []topology.Link{}
	}
	return printJSON(out, map[string]any{"articulation_points": nodes, "bridges": bridges})
}

func cmdBlast(args []string, out io.Writer) error {
	fs := newFlags("blast")
	file := fs.String("topology", "", "topology JSON file")
	node := fs.String("node", "", "node that fails")
	if err := fs.Parse(args); err != nil {
		return usageError(err.Error())
	}
	if *node == "" {
		return usageError("-node is required")
	}
	g, err := loadGraph(*file)
	if err != nil {
		return err
	}
	n, err := g.BlastRadius(topology.NodeID(*node))
	if err != nil {
		return err
	}
	return printJSON(out, map[string]any{"failed": *node, "nodes_cut_off": n})
}

func cmdCfgDiff(args []string, out io.Writer) error {
	fs := newFlags("cfgdiff")
	desired := fs.String("desired", "", "desired config file")
	running := fs.String("running", "", "running config file")
	if err := fs.Parse(args); err != nil {
		return usageError(err.Error())
	}
	if *desired == "" || *running == "" {
		return usageError("-desired and -running are required")
	}
	d, err := os.ReadFile(*desired)
	if err != nil {
		return err
	}
	r, err := os.ReadFile(*running)
	if err != nil {
		return err
	}
	changes, err := cfgdiff.Diff(string(d), string(r))
	if err != nil {
		return err
	}
	if len(changes) == 0 {
		fmt.Fprintln(out, "! no changes: running config already matches desired")
		return nil
	}
	fmt.Fprintln(out, strings.Join(cfgdiff.Commands(changes), "\n"))
	return nil
}
