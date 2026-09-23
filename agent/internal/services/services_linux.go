//go:build linux

package services

import (
	"os/exec"
	"strings"

	"github.com/managerhub/managerhub/shared/protocol"
)

type systemd struct{}

func newManager() Manager { return systemd{} }

func (systemd) List() ([]protocol.ServiceInfo, error) {
	out, err := exec.Command("systemctl", "list-units", "--type=service", "--all",
		"--no-pager", "--no-legend", "--plain").Output()
	if err != nil {
		return nil, err
	}
	var res []protocol.ServiceInfo
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		name := strings.TrimSuffix(fields[0], ".service")
		res = append(res, protocol.ServiceInfo{
			Name:  name,
			State: mapUnitState(fields[2]),
			Sub:   fields[3],
		})
	}
	for i := range res {
		res[i].Enabled = isEnabled(res[i].Name)
	}
	return res, nil
}

func (systemd) Action(name, action string) error {
	switch action {
	case "start", "stop", "restart":
		return exec.Command("systemctl", action, name+".service").Run()
	case "status":
		return exec.Command("systemctl", "is-active", name+".service").Run()
	default:
		return errBadAction(action)
	}
}

func mapUnitState(s string) string {
	switch s {
	case "active":
		return "active"
	case "inactive", "dead":
		return "inactive"
	case "failed":
		return "failed"
	default:
		return "unknown"
	}
}

func isEnabled(name string) bool {
	err := exec.Command("systemctl", "is-enabled", name+".service").Run()
	return err == nil
}

type badAction string

func errBadAction(a string) error { return badAction(a) }
func (b badAction) Error() string { return "services: unsupported action " + string(b) }
