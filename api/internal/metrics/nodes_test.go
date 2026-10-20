package metrics

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	"github.com/UnstoppableMango/thecluster.lan/api/internal/model"
)

var testNow = time.Unix(1_800_000_000, 0)

// fakePrometheus serves canned results keyed by exact query string.
// Queries with no entry return an empty result.
func fakePrometheus(t *testing.T, instant map[string][]map[string]any, ranges map[string][]map[string]any) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("query")
		var resultType string
		var result []map[string]any
		switch r.URL.Path {
		case "/api/v1/query":
			resultType, result = "vector", instant[q]
		case "/api/v1/query_range":
			resultType, result = "matrix", ranges[q]
		default:
			http.NotFound(w, r)
			return
		}
		if result == nil {
			result = []map[string]any{}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "success",
			"data":   map[string]any{"resultType": resultType, "result": result},
		})
	}))
	t.Cleanup(srv.Close)
	return srv
}

func sample(labels map[string]string, v string) map[string]any {
	return map[string]any{"metric": labels, "value": []any{float64(testNow.Unix()), v}}
}

func series(labels map[string]string, vs ...string) map[string]any {
	values := make([]any, len(vs))
	for i, v := range vs {
		values[i] = []any{float64(testNow.Unix() - int64(15*(len(vs)-1-i))), v}
	}
	return map[string]any{"metric": labels, "values": values}
}

func node(name string) map[string]string     { return map[string]string{"node": name} }
func nodename(name string) map[string]string { return map[string]string{"nodename": name} }

func TestSnapshot(t *testing.T) {
	instant := map[string][]map[string]any{
		queryNodes: {
			sample(node("healthy"), "1"),
			sample(node("notready"), "1"),
			sample(node("hotmem"), "1"),
			sample(node("noexporter"), "1"),
			sample(node("cordoned"), "1"),
			sample(node("fullroot"), "1"),
			sample(node("cp"), "1"),
		},
		queryReady: {
			sample(node("healthy"), "1"),
			sample(node("notready"), "0"),
			sample(node("hotmem"), "1"),
			sample(node("noexporter"), "1"),
			sample(node("cordoned"), "1"),
			sample(node("fullroot"), "1"),
			sample(node("cp"), "1"),
		},
		queryCordoned: {
			sample(node("healthy"), "0"),
			sample(node("cordoned"), "1"),
		},
		queryRole: {
			sample(map[string]string{"node": "cp", "role": "control-plane"}, "1"),
			sample(map[string]string{"node": "healthy", "role": "worker"}, "1"),
		},
		queryUname: {
			sample(map[string]string{"nodename": "healthy", "machine": "x86_64"}, "1"),
			sample(map[string]string{"nodename": "notready", "machine": "aarch64"}, "1"),
			sample(map[string]string{"nodename": "hotmem", "machine": "x86_64"}, "1"),
			sample(map[string]string{"nodename": "cordoned", "machine": "aarch64"}, "1"),
			sample(map[string]string{"nodename": "fullroot", "machine": "x86_64"}, "1"),
			sample(map[string]string{"nodename": "cp", "machine": "aarch64"}, "1"),
		},
		queryRootFS: {
			sample(nodename("healthy"), "41.5"),
			sample(nodename("fullroot"), "95"),
		},
		queryClusterCPU: {sample(map[string]string{}, "14.2")},
		queryClusterMem: {sample(map[string]string{}, "NaN")},
	}
	ranges := map[string][]map[string]any{
		queryCPU: {series(nodename("healthy"), "10", "20", "NaN", "30")},
		queryMem: {
			series(nodename("healthy"), "50", "51"),
			series(nodename("hotmem"), "85", "95"),
		},
		queryDisk: {series(nodename("healthy"), "1024")},
	}

	p := NewPrometheus(fakePrometheus(t, instant, ranges).URL)
	p.Now = func() time.Time { return testNow }

	snap, err := p.Snapshot(context.Background())
	if err != nil {
		t.Fatalf("Snapshot: %v", err)
	}

	names := make([]string, len(snap.Nodes))
	byName := map[string]model.Node{}
	for i, n := range snap.Nodes {
		names[i] = n.Name
		byName[n.Name] = n
	}

	wantOrder := []string{"cp", "cordoned", "fullroot", "healthy", "hotmem", "noexporter", "notready"}
	if !slices.Equal(names, wantOrder) {
		t.Errorf("order = %v, want %v", names, wantOrder)
	}

	cases := []struct {
		name    string
		health  model.Health
		reasons []string
	}{
		{"healthy", model.HealthOk, []string{}},
		{"cp", model.HealthOk, []string{}},
		{"cordoned", model.HealthOk, []string{}},
		{"notready", model.HealthBad, []string{"not ready"}},
		{"hotmem", model.HealthBad, []string{"mem > 90%"}},
		{"noexporter", model.HealthBad, []string{"no metrics"}},
		{"fullroot", model.HealthBad, []string{"root fs > 90%"}},
	}
	for _, c := range cases {
		n := byName[c.name]
		if n.Health != c.health || !slices.Equal(n.Reasons, c.reasons) {
			t.Errorf("%s: health=%s reasons=%v, want %s %v", c.name, n.Health, n.Reasons, c.health, c.reasons)
		}
	}

	if !byName["cordoned"].Cordoned || byName["healthy"].Cordoned {
		t.Errorf("cordoned flags wrong: %+v", byName)
	}
	if r := byName["cp"].Role; r != "control-plane" {
		t.Errorf("cp role = %q", r)
	}
	if r := byName["notready"].Role; r != "worker" {
		t.Errorf("default role = %q, want worker", r)
	}
	if a := byName["healthy"].Arch; a != "amd64" {
		t.Errorf("healthy arch = %q", a)
	}
	if a := byName["cp"].Arch; a != "arm64" {
		t.Errorf("cp arch = %q", a)
	}

	h := byName["healthy"]
	if h.RootFSPct == nil || *h.RootFSPct != 41.5 {
		t.Errorf("healthy rootfs = %v", h.RootFSPct)
	}
	if len(h.Series.CPU) != 3 {
		t.Errorf("expected NaN dropped from cpu series, got %v", h.Series.CPU)
	}
	if len(h.Series.Disk) != 1 || h.Series.Disk[0][1] != 1024 {
		t.Errorf("disk series = %v", h.Series.Disk)
	}
	if byName["noexporter"].Series.Net == nil {
		t.Error("series slices must be non-nil so they encode as []")
	}

	if snap.Cluster.CPUPct == nil || *snap.Cluster.CPUPct != 14.2 {
		t.Errorf("cluster cpu = %v", snap.Cluster.CPUPct)
	}
	if snap.Cluster.MemPct != nil {
		t.Errorf("cluster mem should be nil for NaN, got %v", *snap.Cluster.MemPct)
	}
}

func TestSnapshotPrometheusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"status":"error","errorType":"bad_data","error":"parse error"}`))
	}))
	t.Cleanup(srv.Close)

	if _, err := NewPrometheus(srv.URL).Snapshot(context.Background()); err == nil {
		t.Fatal("expected error")
	}
}

func TestSnapshotJSONShape(t *testing.T) {
	instant := map[string][]map[string]any{
		queryNodes: {sample(node("solo"), "1")},
	}
	p := NewPrometheus(fakePrometheus(t, instant, nil).URL)
	p.Now = func() time.Time { return testNow }

	snap, err := p.Snapshot(context.Background())
	if err != nil {
		t.Fatalf("Snapshot: %v", err)
	}
	b, err := json.Marshal(snap.Nodes[0])
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got map[string]any
	_ = json.Unmarshal(b, &got)
	s := got["series"].(map[string]any)
	for _, k := range []string{"cpu", "mem", "disk", "net"} {
		if _, ok := s[k].([]any); !ok {
			t.Errorf("series.%s = %v, want []", k, s[k])
		}
	}
	if _, ok := got["reasons"].([]any); !ok {
		t.Errorf("reasons = %v, want array", got["reasons"])
	}
}
