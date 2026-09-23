package protocol

// DockerListResult returns detected containers.
type DockerListResult struct {
	ReqID      string              `json:"req_id"`
	Available  bool                `json:"available"`
	Containers []map[string]string `json:"containers"`
	Error      string              `json:"error,omitempty"`
}

// DockerAction requests start/stop/restart/logs on a container.
type DockerAction struct {
	ReqID  string `json:"req_id"`
	Name   string `json:"name"`
	Action string `json:"action"`
}

// DockerActionResult reports the outcome (logs included for action=logs).
type DockerActionResult struct {
	ReqID  string `json:"req_id"`
	Name   string `json:"name"`
	Action string `json:"action"`
	OK     bool   `json:"ok"`
	Logs   string `json:"logs,omitempty"`
	Error  string `json:"error,omitempty"`
}
