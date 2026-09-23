package protocol

// Hello is sent by the agent right after the WS handshake.
type Hello struct {
	NodeID       string            `json:"node_id,omitempty"`
	Name         string            `json:"name"`
	Hostname     string            `json:"hostname"`
	OS           string            `json:"os"`
	Arch         string            `json:"arch"`
	AgentVer     string            `json:"agent_version"`
	IP           string            `json:"ip"`
	Tags         []string          `json:"tags,omitempty"`
	Capabilities map[string]string `json:"capabilities,omitempty"`
	CPULimitPct  float64           `json:"cpu_limit_pct,omitempty"`
	RAMLimitMB   uint64            `json:"ram_limit_mb,omitempty"`
}

// HelloAck confirms registration and returns the authoritative node ID.
type HelloAck struct {
	NodeID            string `json:"node_id"`
	HeartbeatInterval int    `json:"heartbeat_interval_sec"`
	MetricsInterval   int    `json:"metrics_interval_sec"`
	ServerTime        int64  `json:"server_time"`
}

// Heartbeat keeps the session alive and reports liveness.
type Heartbeat struct {
	UptimeSec int64 `json:"uptime_sec"`
}

// Metrics is a snapshot of host resource usage.
type Metrics struct {
	CPUPercent  float64   `json:"cpu_percent"`
	CPUCores    int       `json:"cpu_cores"`
	Load1       float64   `json:"load1"`
	RAMUsed     uint64    `json:"ram_used"`
	RAMTotal    uint64    `json:"ram_total"`
	DiskUsed    uint64    `json:"disk_used"`
	DiskTotal   uint64    `json:"disk_total"`
	UptimeSec   int64     `json:"uptime_sec"`
	GPUs        []GPUInfo `json:"gpus,omitempty"`
	Processes   int       `json:"processes,omitempty"`
	CPULimitPct float64   `json:"cpu_limit_pct,omitempty"`
	RAMLimitMB  uint64    `json:"ram_limit_mb,omitempty"`
	BatteryPct  *int      `json:"battery_pct,omitempty"`
	Charging    *bool     `json:"charging,omitempty"`
	NetworkType string    `json:"network_type,omitempty"`
}

// GPUInfo describes a detected GPU device.
type GPUInfo struct {
	Vendor string `json:"vendor"`
	Model  string `json:"model"`
	VRAM   uint64 `json:"vram,omitempty"`
}

// JobAssign asks the agent to execute a job.
type JobAssign struct {
	JobID      string            `json:"job_id"`
	Name       string            `json:"name"`
	Type       string            `json:"type"`
	Command    string            `json:"command"`
	Args       []string          `json:"args,omitempty"`
	WorkDir    string            `json:"workdir,omitempty"`
	Env        map[string]string `json:"env,omitempty"`
	TimeoutSec int               `json:"timeout_sec"`
}

// JobCancel requests termination of a running job.
type JobCancel struct {
	JobID string `json:"job_id"`
}

// JobOutput streams stdout/stderr of a job.
type JobOutput struct {
	JobID  string `json:"job_id"`
	Stream string `json:"stream"` // stdout | stderr
	Data   string `json:"data"`
	Seq    int64  `json:"seq"`
}

// JobResult reports final state of a job.
type JobResult struct {
	JobID    string `json:"job_id"`
	Status   string `json:"status"`
	ExitCode int    `json:"exit_code"`
	Error    string `json:"error,omitempty"`
}
