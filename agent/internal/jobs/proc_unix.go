//go:build !windows

package jobs

import (
	"os/exec"
	"syscall"
)

// setSysProcAttr puts the child in its own process group so cancellation
// can kill the whole tree.
func setSysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// killTree terminates the process group of cmd.
func killTree(cmd *exec.Cmd) {
	if cmd.Process != nil {
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
}

var _ = killTree
