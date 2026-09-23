//go:build darwin

package services

import (
	"os/exec"
	"strings"

	"github.com/managerhub/managerhub/shared/protocol"
)

type launchd struct{}

func newManager() Manager { return launchd{} }

func (launchd) List() ([]protocol.ServiceInfo, error) {
	out, err := exec.Command("launchctl", "list").Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(out), "\n")
	var res []protocol.ServiceInfo
	for i, line := range lines {
		if i == 0 {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		state := "inactive"
		if fields[0] != "-" {
			state = "active"
		}
		res = append(res, protocol.ServiceInfo{Name: fields[2], State: state, Sub: fields[1]})
	}
	return res, nil
}

func (launchd) Action(name, action string) error {
	switch action {
	case "start":
		return exec.Command("launchctl", "kickstart", "-k", "system/"+name).Run()
	case "stop":
		return exec.Command("launchctl", "kill", "SIGTERM", "system/"+name).Run()
	case "restart":
		_ = exec.Command("launchctl", "kill", "SIGTERM", "system/"+name).Run()
		return exec.Command("launchctl", "kickstart", "-k", "system/"+name).Run()
	case "status":
		return exec.Command("launchctl", "print", "system/"+name).Run()
	default:
		return errBadAction(action)
	}
}

type badAction string

func errBadAction(a string) error { return badAction(a) }
func (b badAction) Error() string { return "services: unsupported action " + string(b) }
