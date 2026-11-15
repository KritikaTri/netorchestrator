package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, name, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func do(args ...string) (code int, out, errOut string) {
	var o, e bytes.Buffer
	code = run(args, &o, &e)
	return code, o.String(), e.String()
}

const topo = `{"links":[
 {"a":"r1","b":"r2","cost":1},{"a":"r2","b":"r3","cost":1},{"a":"r1","b":"r3","cost":5},
 {"a":"r3","b":"edge","cost":1}]}`

func TestPath(t *testing.T) {
	f := write(t, "t.json", topo)
	code, out, _ := do("path", "-topology", f, "-from", "r1", "-to", "edge")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	var p struct {
		Nodes []string
		Cost  float64
	}
	if err := json.Unmarshal([]byte(out), &p); err != nil {
		t.Fatal(err)
	}
	if p.Cost != 3 || strings.Join(p.Nodes, ">") != "r1>r2>r3>edge" {
		t.Fatalf("got %+v", p)
	}
}

func TestSPOFAndBlast(t *testing.T) {
	f := write(t, "t.json", topo)
	code, out, _ := do("spof", "-topology", f)
	if code != 0 || !strings.Contains(out, `"r3"`) {
		t.Fatalf("spof: %d %s", code, out)
	}
	code, out, _ = do("blast", "-topology", f, "-node", "r3")
	if code != 0 || !strings.Contains(out, `"nodes_cut_off": 1`) {
		t.Fatalf("blast: %d %s", code, out)
	}
}

func TestCfgDiff(t *testing.T) {
	d := write(t, "d.cfg", "interface Gi0/1\n description new\n")
	r := write(t, "r.cfg", "interface Gi0/1\n description old\n")
	code, out, _ := do("cfgdiff", "-desired", d, "-running", r)
	if code != 0 || !strings.Contains(out, " no description old") || !strings.Contains(out, " description new") {
		t.Fatalf("%d %q", code, out)
	}
	code, out, _ = do("cfgdiff", "-desired", d, "-running", d)
	if code != 0 || !strings.Contains(out, "no changes") {
		t.Fatalf("%d %q", code, out)
	}
}

func TestExitCodes(t *testing.T) {
	f := write(t, "t.json", topo)
	cases := []struct {
		name string
		args []string
		want int
	}{
		{"no args", nil, 2},
		{"unknown command", []string{"frobnicate"}, 2},
		{"missing flags", []string{"path", "-topology", f}, 2},
		{"bad flag", []string{"path", "-nope"}, 2},
		{"missing file", []string{"spof", "-topology", "/does/not/exist.json"}, 1},
		{"unreachable", []string{"path", "-topology", write(t, "i.json", `{"nodes":["x"],"links":[{"a":"a","b":"b","cost":1}]}`), "-from", "a", "-to", "x"}, 1},
		{"unknown node", []string{"blast", "-topology", f, "-node", "ghost"}, 1},
		{"help", []string{"help"}, 0},
	}
	for _, c := range cases {
		if got, _, _ := do(c.args...); got != c.want {
			t.Errorf("%s: exit %d, want %d", c.name, got, c.want)
		}
	}
}
