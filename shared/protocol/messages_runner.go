package protocol

// RunnerInfo describes a GitHub Actions self-hosted runner.
type RunnerInfo struct {
	Name       string `json:"name"`
	Repo       string `json:"repo,omitempty"`
	Org        string `json:"org,omitempty"`
	Status     string `json:"status"` // idle|busy|offline|unknown
	Service    string `json:"service,omitempty"`
	ServiceOn  bool   `json:"service_on"`
	WorkDir    string `json:"workdir,omitempty"`
	Version    string `json:"version,omitempty"`
}

// RunnerListResult returns detected runners.
type RunnerListResult struct {
	ReqID   string       `json:"req_id"`
	Runners []RunnerInfo `json:"runners"`
	Error   string       `json:"error,omitempty"`
}

// RunnerAction requests start/stop/restart/logs on a runner service.
type RunnerAction struct {
	ReqID  string `json:"req_id"`
	Name   string `json:"name"`
	Action string `json:"action"` // start|stop|restart|logs
}

// RunnerActionResult reports the outcome (logs truncated to Lines).
type RunnerActionResult struct {
	ReqID  string `json:"req_id"`
	Name   string `json:"name"`
	Action string `json:"action"`
	OK     bool   `json:"ok"`
	Logs   string `json:"logs,omitempty"`
	Error  string `json:"error,omitempty"`
}

// ErrorPayload is a generic error frame.
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	RefID   string `json:"ref_id,omitempty"`
}

// PingPong payload for keepalive probes.
type PingPong struct {
	Nonce string `json:"nonce"`
}
