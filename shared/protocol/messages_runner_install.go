package protocol

// RunnerInstall requests installation of a GitHub Actions runner.
// Scope can be "repo" (URL = github.com/owner/repo) or "org" (URL = github.com/org).
type RunnerInstall struct {
	ReqID   string `json:"req_id"`
	RepoURL string `json:"repo_url"` // https://github.com/owner/repo or https://github.com/org
	Token   string `json:"token"`    // registration token from GitHub
	Name    string `json:"name"`     // runner name
	Labels  string `json:"labels"`   // custom labels, comma-separated
	WorkDir string `json:"workdir"`  // install directory (default: /opt/actions-runner)
}

// RunnerInstallResult reports the outcome of a runner installation.
type RunnerInstallResult struct {
	ReqID  string `json:"req_id"`
	OK     bool   `json:"ok"`
	Output string `json:"output"`
	Error  string `json:"error,omitempty"`
}
