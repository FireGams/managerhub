// Package runner detects and controls GitHub Actions self-hosted runners.
package runner

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/managerhub/managerhub/shared/protocol"
)

// Detector finds local GitHub Actions runners.
type Detector struct {
	svc ServiceController
}

// ServiceController abstracts service control for runner services.
type ServiceController interface {
	List() ([]protocol.ServiceInfo, error)
	Action(name, action string) error
}

// NewDetector builds a detector using a service controller.
func NewDetector(svc ServiceController) *Detector { return &Detector{svc: svc} }

// List returns every detected self-hosted runner.
func (d *Detector) List() ([]protocol.RunnerInfo, error) {
	services, _ := d.svc.List()
	var out []protocol.RunnerInfo
	seen := map[string]bool{}

	for _, s := range services {
		lower := strings.ToLower(s.Name)
		if !strings.Contains(lower, "actions.runner") && !strings.Contains(lower, "github") {
			continue
		}
		if !strings.Contains(lower, "runner") {
			continue
		}
		info := protocol.RunnerInfo{
			Name:      s.Name,
			Service:   s.Name,
			ServiceOn: s.State == "active",
			Status:    "offline",
		}
		if s.State == "active" {
			info.Status = "idle"
		}
		if wd := findRunnerDir(); wd != "" {
			info.WorkDir = wd
			info.Name = readRunnerName(wd, s.Name)
			info.Repo = readRunnerRepo(wd)
		}
		seen[s.Name] = true
		out = append(out, info)
	}

	// Also pick up runners started in console mode (no service).
	if wd := findRunnerDir(); wd != "" && len(out) == 0 {
		out = append(out, protocol.RunnerInfo{
			Name: readRunnerName(wd, "runner"), WorkDir: wd,
			Status: "unknown", Service: "", ServiceOn: false,
		})
	}
	return out, nil
}

// Action performs start/stop/restart on the runner service; logs reads the log file.
func (d *Detector) Action(name, action string) (string, error) {
	if action == "logs" {
		return readRunnerLogs(findRunnerDir()), nil
	}
	return "", d.svc.Action(name, action)
}

func findRunnerDir() string {
	candidates := []string{
		"/opt/actions-runner",
		"/home/runner/actions-runner",
		"/opt/github-runner",
	}
	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates,
			filepath.Join(home, "actions-runner"),
			filepath.Join(home, "runner"))
	}
	for _, c := range candidates {
		if st, err := os.Stat(filepath.Join(c, ".runner")); err == nil && !st.IsDir() {
			return c
		}
	}
	return ""
}

func readRunnerName(dir, fallback string) string {
	b, err := os.ReadFile(filepath.Join(dir, ".runner"))
	if err != nil {
		return fallback
	}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "agentName") || strings.Contains(line, "\"agentName\"") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.Trim(strings.TrimSpace(parts[1]), `",`)
			}
		}
	}
	return fallback
}

func readRunnerRepo(dir string) string {
	b, err := os.ReadFile(filepath.Join(dir, ".credentials"))
	if err != nil {
		b, err = os.ReadFile(filepath.Join(dir, ".runner"))
		if err != nil {
			return ""
		}
	}
	s := string(b)
	for _, key := range []string{"githubUrl", "serverUrl", "url"} {
		if i := strings.Index(s, key); i >= 0 {
			rest := s[i:]
			if j := strings.Index(rest, "http"); j >= 0 {
				rest = rest[j:]
				if k := strings.IndexAny(rest, "\"'\n "); k > 0 {
					return rest[:k]
				}
			}
		}
	}
	return ""
}

func readRunnerLogs(dir string) string {
	if dir == "" {
		return "runner directory not found"
	}
	logDir := filepath.Join(dir, "_diag")
	entries, err := os.ReadDir(logDir)
	if err != nil {
		return "no logs: " + err.Error()
	}
	var latest string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".log") {
			if latest == "" || e.Name() > latest {
				latest = e.Name()
			}
		}
	}
	if latest == "" {
		return "no log files"
	}
	b, err := os.ReadFile(filepath.Join(logDir, latest))
	if err != nil {
		return err.Error()
	}
	s := string(b)
	if len(s) > 8000 {
		s = s[len(s)-8000:]
	}
	return s
}
