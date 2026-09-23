//go:build windows

package jobs

import "os/exec"

func setSysProcAttr(cmd *exec.Cmd) {}

func killTree(cmd *exec.Cmd) {
	if cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
}
