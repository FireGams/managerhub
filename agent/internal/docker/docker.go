// Package docker detects and manages Docker containers on the host.
package docker

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// Container describes a Docker container.
type Container struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Image  string `json:"image"`
	Status string `json:"status"`
	State  string `json:"state"`
}

// Manager interacts with the local Docker daemon.
type Manager struct{}

// New creates a Docker manager.
func New() *Manager { return &Manager{} }

// Available reports whether Docker is installed and reachable.
func (m *Manager) Available() bool {
	// Check multiple docker binary locations (systemd PATH is minimal)
	for _, bin := range dockerBinaries() {
		if exec.Command(bin, "info").Run() == nil {
			return true
		}
	}
	// Fallback: check if Docker socket exists
	if _, err := os.Stat("/var/run/docker.sock"); err == nil {
		return true
	}
	return false
}

func dockerBinaries() []string {
	return []string{
		"docker",
		"/usr/bin/docker",
		"/usr/local/bin/docker",
		"/snap/bin/docker",
	}
}

func dockerCmd(args ...string) *exec.Cmd {
	for _, bin := range dockerBinaries() {
		if _, err := exec.LookPath(bin); err == nil {
			return exec.Command(bin, args...)
		}
	}
	return exec.Command("docker", args...)
}

// List returns running and stopped containers.
func (m *Manager) List() ([]Container, error) {
	out, err := dockerCmd("ps", "-a",
		"--format", "{{.ID}}\t{{.Names}}\t{{.Image}}\t{{.Status}}\t{{.State}}").Output()
	if err != nil {
		return nil, err
	}
	var res []Container
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 5)
		if len(parts) < 5 {
			continue
		}
		res = append(res, Container{
			ID: parts[0], Name: parts[1], Image: parts[2],
			Status: parts[3], State: parts[4],
		})
	}
	return res, nil
}

// Action performs start/stop/restart on a container.
func (m *Manager) Action(nameOrID, action string) error {
	switch action {
	case "start", "stop", "restart":
		return dockerCmd(action, nameOrID).Run()
	default:
		return errBadAction(action)
	}
}

// Logs returns the last N lines of container logs.
func (m *Manager) Logs(nameOrID string, tail int) (string, error) {
	if tail <= 0 {
		tail = 100
	}
	out, err := dockerCmd("logs", "--tail", strconv.Itoa(tail), nameOrID).CombinedOutput()
	return string(out), err
}

type badAction string

func errBadAction(a string) error { return badAction(a) }
func (b badAction) Error() string { return "docker: unsupported action " + string(b) }
