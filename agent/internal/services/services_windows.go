//go:build windows

package services

import (
	"os/exec"
	"strings"

	"github.com/managerhub/managerhub/shared/protocol"
)

type wsc struct{}

func newManager() Manager { return wsc{} }

func (wsc) List() ([]protocol.ServiceInfo, error) {
	out, err := exec.Command("powershell", "-NoProfile", "-Command",
		"Get-Service | Select-Object -Property Name,Status,StartType | ConvertTo-Json").Output()
	if err != nil {
		return nil, err
	}
	return parsePSJSON(string(out)), nil
}

func (wsc) Action(name, action string) error {
	switch action {
	case "start", "stop", "restart":
		return exec.Command("powershell", "-NoProfile", "-Command",
			action+"-Service -Name '"+name+"' -Force").Run()
	case "status":
		return exec.Command("powershell", "-NoProfile", "-Command",
			"(Get-Service -Name '"+name+"').Status").Run()
	default:
		return errBadAction(action)
	}
}

func parsePSJSON(s string) []protocol.ServiceInfo {
	// minimal parse: array of {Name,Status,StartType}
	var res []protocol.ServiceInfo
	s = strings.TrimSpace(s)
	if s == "" || s == "null" {
		return res
	}
	// fallback line parser for simple arrays
	for _, part := range strings.Split(s, "},{") {
		name := between(part, "\"Name\":\"", "\"")
		status := between(part, "\"Status\":\"", "\"")
		if name == "" {
			continue
		}
		state := "unknown"
		if strings.EqualFold(status, "Running") {
			state = "active"
		} else if strings.EqualFold(status, "Stopped") {
			state = "inactive"
		}
		res = append(res, protocol.ServiceInfo{Name: name, State: state})
	}
	return res
}

func between(s, a, b string) string {
	i := strings.Index(s, a)
	if i < 0 {
		return ""
	}
	s = s[i+len(a):]
	j := strings.Index(s, b)
	if j < 0 {
		return s
	}
	return s[:j]
}

type badAction string

func errBadAction(a string) error { return badAction(a) }
func (b badAction) Error() string { return "services: unsupported action " + string(b) }
