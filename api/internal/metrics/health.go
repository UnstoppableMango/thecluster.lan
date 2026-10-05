package metrics

import "fmt"

type Health string

const (
	HealthOK  Health = "ok"
	HealthBad Health = "bad"
)

const (
	MemThresholdPct    = 90.0
	RootFSThresholdPct = 90.0
)

// assess marks a node bad when it is NotReady, has no node-exporter data,
// or is over a resource threshold. Cordoning alone does not affect health.
func assess(n *Node) {
	reasons := []string{}

	if !n.Ready {
		reasons = append(reasons, "not ready")
	}
	if !n.Exporter {
		reasons = append(reasons, "no metrics")
	}
	if mem, ok := last(n.Series.Mem); ok && mem > MemThresholdPct {
		reasons = append(reasons, fmt.Sprintf("mem > %.0f%%", MemThresholdPct))
	}
	if n.RootFSPct != nil && *n.RootFSPct > RootFSThresholdPct {
		reasons = append(reasons, fmt.Sprintf("root fs > %.0f%%", RootFSThresholdPct))
	}

	n.Reasons = reasons
	if len(reasons) == 0 {
		n.Health = HealthOK
	} else {
		n.Health = HealthBad
	}
}

func last(points []Point) (float64, bool) {
	if len(points) == 0 {
		return 0, false
	}
	return points[len(points)-1][1], true
}
