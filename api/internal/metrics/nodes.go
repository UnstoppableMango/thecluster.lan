package metrics

import (
	"context"
	"sort"
	"sync"
	"time"
)

const (
	Window = 15 * time.Minute
	Step   = 15 * time.Second
)

// Excludes virtual block devices so disk throughput is not double counted.
const diskDevices = `device!~"loop.*|ram.*|zram.*|dm-.*|md.*|sr.*"`

// Excludes loopback and per-pod virtual interfaces so network throughput reflects the physical NICs.
const netDevices = `device!~"lo|veth.*|cali.*|cilium.*|lxc.*|flannel.*|cni.*|vxlan.*|tunl.*|kube-ipvs.*"`

// byNode relabels a per-instance node-exporter expression by node name.
func byNode(expr string) string {
	return `max by (nodename) ((` + expr + `) * on(instance) group_left(nodename) node_uname_info)`
}

var (
	queryNodes      = `kube_node_info`
	queryReady      = `kube_node_status_condition{condition="Ready",status="true"}`
	queryCordoned   = `kube_node_spec_unschedulable`
	queryRole       = `kube_node_role`
	queryUname      = `node_uname_info`
	queryRootFS     = byNode(`100 * (1 - node_filesystem_avail_bytes{mountpoint="/",fstype!="rootfs"} / node_filesystem_size_bytes{mountpoint="/",fstype!="rootfs"})`)
	queryCPU        = byNode(`100 * (1 - avg by (instance) (rate(node_cpu_seconds_total{mode="idle"}[1m])))`)
	queryMem        = byNode(`100 * (1 - node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)`)
	queryDisk       = byNode(`sum by (instance) (rate(node_disk_read_bytes_total{` + diskDevices + `}[1m]) + rate(node_disk_written_bytes_total{` + diskDevices + `}[1m]))`)
	queryNet        = byNode(`sum by (instance) (rate(node_network_receive_bytes_total{` + netDevices + `}[1m]) + rate(node_network_transmit_bytes_total{` + netDevices + `}[1m]))`)
	queryClusterCPU = `100 * (1 - sum(rate(node_cpu_seconds_total{mode="idle"}[1m])) / count(node_cpu_seconds_total{mode="idle"}))`
	queryClusterMem = `100 * (1 - sum(node_memory_MemAvailable_bytes) / sum(node_memory_MemTotal_bytes))`
)

type Snapshot struct {
	Updated time.Time `json:"updated"`
	Cluster Cluster   `json:"cluster"`
	Nodes   []Node    `json:"nodes"`
}

type Cluster struct {
	CPUPct *float64 `json:"cpuPct"`
	MemPct *float64 `json:"memPct"`
}

type Node struct {
	Name      string     `json:"name"`
	Role      string     `json:"role"`
	Arch      string     `json:"arch,omitempty"`
	Health    Health     `json:"health"`
	Reasons   []string   `json:"reasons"`
	Ready     bool       `json:"ready"`
	Cordoned  bool       `json:"cordoned"`
	Exporter  bool       `json:"exporter"`
	RootFSPct *float64   `json:"rootFsPct"`
	Series    NodeSeries `json:"series"`
}

// NodeSeries holds the last Window of samples: CPU and Mem in percent, Disk and Net in bytes/sec.
type NodeSeries struct {
	CPU  []Point `json:"cpu"`
	Mem  []Point `json:"mem"`
	Disk []Point `json:"disk"`
	Net  []Point `json:"net"`
}

type Source interface {
	Snapshot(ctx context.Context) (*Snapshot, error)
}

// Prometheus builds snapshots from kube-state-metrics and node-exporter data.
type Prometheus struct {
	Client *Client
	Now    func() time.Time
}

func NewPrometheus(baseURL string) *Prometheus {
	return &Prometheus{Client: NewClient(baseURL), Now: time.Now}
}

func (p *Prometheus) Snapshot(ctx context.Context) (*Snapshot, error) {
	now := p.Now()
	start := now.Add(-Window)

	instant := map[string][]Sample{}
	ranges := map[string][]Series{}
	var mu sync.Mutex

	var tasks []func() error
	for _, q := range []string{queryNodes, queryReady, queryCordoned, queryRole, queryUname, queryRootFS, queryClusterCPU, queryClusterMem} {
		tasks = append(tasks, func() error {
			res, err := p.Client.Query(ctx, q, now)
			mu.Lock()
			instant[q] = res
			mu.Unlock()
			return err
		})
	}
	for _, q := range []string{queryCPU, queryMem, queryDisk, queryNet} {
		tasks = append(tasks, func() error {
			res, err := p.Client.QueryRange(ctx, q, start, now, Step)
			mu.Lock()
			ranges[q] = res
			mu.Unlock()
			return err
		})
	}
	if err := runAll(tasks); err != nil {
		return nil, err
	}

	nodes := map[string]*Node{}
	for _, s := range instant[queryNodes] {
		name := s.Labels["node"]
		if name == "" {
			continue
		}
		nodes[name] = &Node{
			Name:    name,
			Role:    "worker",
			Reasons: []string{},
			Series:  NodeSeries{CPU: []Point{}, Mem: []Point{}, Disk: []Point{}, Net: []Point{}},
		}
	}

	for _, s := range instant[queryReady] {
		if n := nodes[s.Labels["node"]]; n != nil {
			n.Ready = s.Value == 1
		}
	}
	for _, s := range instant[queryCordoned] {
		if n := nodes[s.Labels["node"]]; n != nil {
			n.Cordoned = s.Value == 1
		}
	}
	for _, s := range instant[queryRole] {
		if n := nodes[s.Labels["node"]]; n != nil && n.Role != "control-plane" {
			n.Role = s.Labels["role"]
		}
	}
	for _, s := range instant[queryUname] {
		if n := nodes[s.Labels["nodename"]]; n != nil {
			n.Exporter = true
			n.Arch = normalizeArch(s.Labels["machine"])
		}
	}
	for _, s := range instant[queryRootFS] {
		if n := nodes[s.Labels["nodename"]]; n != nil {
			n.RootFSPct = ptr(s.Value)
		}
	}

	assign := func(q string, field func(*NodeSeries) *[]Point) {
		for _, s := range ranges[q] {
			if n := nodes[s.Labels["nodename"]]; n != nil {
				*field(&n.Series) = s.Points
			}
		}
	}
	assign(queryCPU, func(s *NodeSeries) *[]Point { return &s.CPU })
	assign(queryMem, func(s *NodeSeries) *[]Point { return &s.Mem })
	assign(queryDisk, func(s *NodeSeries) *[]Point { return &s.Disk })
	assign(queryNet, func(s *NodeSeries) *[]Point { return &s.Net })

	snap := &Snapshot{
		Updated: now.UTC(),
		Cluster: Cluster{
			CPUPct: first(instant[queryClusterCPU]),
			MemPct: first(instant[queryClusterMem]),
		},
		Nodes: make([]Node, 0, len(nodes)),
	}
	for _, n := range nodes {
		assess(n)
		snap.Nodes = append(snap.Nodes, *n)
	}
	sort.Slice(snap.Nodes, func(i, j int) bool {
		a, b := snap.Nodes[i], snap.Nodes[j]
		if (a.Role == "control-plane") != (b.Role == "control-plane") {
			return a.Role == "control-plane"
		}
		return a.Name < b.Name
	})

	return snap, nil
}

func runAll(tasks []func() error) error {
	errs := make(chan error, len(tasks))
	var wg sync.WaitGroup
	for _, task := range tasks {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- task()
		}()
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

func normalizeArch(machine string) string {
	switch machine {
	case "x86_64":
		return "amd64"
	case "aarch64":
		return "arm64"
	default:
		return machine
	}
}

func first(samples []Sample) *float64 {
	if len(samples) == 0 {
		return nil
	}
	return ptr(samples[0].Value)
}

func ptr(v float64) *float64 {
	return &v
}
