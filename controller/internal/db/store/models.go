package store

import "time"

// User is a web UI account.
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

// Node is a registered machine running an agent.
type Node struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Hostname     string            `json:"hostname"`
	OS           string            `json:"os"`
	Arch         string            `json:"arch"`
	IP           string            `json:"ip"`
	AgentVersion string            `json:"agent_version"`
	Status       string            `json:"status"`
	LastSeenAt   *time.Time        `json:"last_seen_at,omitempty"`
	Tags         []string          `json:"tags"`
	Capabilities map[string]string `json:"capabilities"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	// Geolocation (from MeshService).
	City    string `json:"city,omitempty"`
	Country string `json:"country,omitempty"`
}

// Metrics is one resource sample for a node.
type Metrics struct {
	NodeID     string    `json:"node_id"`
	TS         time.Time `json:"ts"`
	CPUPercent float64   `json:"cpu_percent"`
	CPUCores   int       `json:"cpu_cores"`
	Load1      float64   `json:"load1"`
	RAMUsed    uint64    `json:"ram_used"`
	RAMTotal   uint64    `json:"ram_total"`
	DiskUsed   uint64    `json:"disk_used"`
	DiskTotal  uint64    `json:"disk_total"`
	UptimeSec  int64     `json:"uptime_sec"`
	GPUs       string    `json:"gpus"`
}

// Job is a unit of remote work.
type Job struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Type       string            `json:"type"`
	NodeID     *string           `json:"node_id,omitempty"`
	Command    string            `json:"command"`
	Args       []string          `json:"args"`
	WorkDir    string            `json:"workdir"`
	Env        map[string]string `json:"env"`
	TimeoutSec int               `json:"timeout_sec"`
	Status     string            `json:"status"`
	ExitCode   *int              `json:"exit_code,omitempty"`
	Error      string            `json:"error,omitempty"`
	Stdout     string            `json:"stdout,omitempty"`
	Stderr     string            `json:"stderr,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
	StartedAt  *time.Time        `json:"started_at,omitempty"`
	FinishedAt *time.Time        `json:"finished_at,omitempty"`
}

// Schedule is a cron job template owned by the controller.
type Schedule struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	CronExpr   string            `json:"cron_expr"`
	Enabled    bool              `json:"enabled"`
	JobName    string            `json:"job_name"`
	JobType    string            `json:"job_type"`
	Command    string            `json:"command"`
	Args       []string          `json:"args"`
	WorkDir    string            `json:"workdir"`
	Env        map[string]string `json:"env"`
	TimeoutSec int               `json:"timeout_sec"`
	NodeID     *string           `json:"node_id,omitempty"`
	Selector   map[string]any    `json:"selector"`
	LastRunAt  *time.Time        `json:"last_run_at,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
}

// AuditEntry records an administrative action.
type AuditEntry struct {
	ID     int64     `json:"id"`
	TS     time.Time `json:"ts"`
	Actor  string    `json:"actor"`
	Action string    `json:"action"`
	Target string    `json:"target"`
	Detail string    `json:"detail"`
}
