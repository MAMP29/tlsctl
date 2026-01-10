package scan

import (
	"fmt"
	"strings"
)

type ScanEndpoint struct {
	IP          string  `json:"ipAddress"`
	Status      string  `json:"statusMessage"`
	Grade       string  `json:"grade"`
	Progress    int     `json:"progress"`
	HasWarnings bool    `json:"hasWarnings"`
	Duration    int     `json:"duration"`
	Error       *string `json:"statusDetailsMessage,omitempty"`
}

type ScanTask struct {
	Host      string         `json:"host"`
	Status    string         `json:"status"` // "running", "done", "error", "timeout"
	Progress  int            `json:"progress"`
	Error     *string        `json:"statusMessage,omitempty"`
	StartedAt int            `json:"startTime"`
	Endpoints []ScanEndpoint `json:"endpoints"`
}

func (e ScanEndpoint) String() string {
	return fmt.Sprintf(
		"Endpoint{IP=%s, Status=%s, Grade=%s, Progress=%d%%, Warnings=%t, Duration=%d}",
		e.IP,
		e.Status,
		e.Grade,
		e.Progress,
		e.HasWarnings,
		e.Duration,
	)
}

func (t ScanTask) String() string {
	var b strings.Builder

	fmt.Fprintf(&b, "ScanTask{\n")
	fmt.Fprintf(&b, "  Host: %s\n", t.Host)
	fmt.Fprintf(&b, "  Status: %s\n", t.Status)
	fmt.Fprintf(&b, "  Progress: %d%%\n", t.Progress)
	fmt.Fprintf(&b, "  StartedAt: %d\n", t.StartedAt)

	if !t.EndpointsAtZero() {
		fmt.Fprintf(&b, "  Endpoints:\n")
		for _, ep := range t.Endpoints {
			fmt.Fprintf(&b, "    - %s\n", ep.String())
		}
	} else {
		fmt.Fprintf(&b, "  Endpoints: []\n")
	}

	fmt.Fprintf(&b, "}")

	return b.String()
}

func (t ScanTask) EndpointsAtZero() bool {
	return len(t.Endpoints) == 0
}
