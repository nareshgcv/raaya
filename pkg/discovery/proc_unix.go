//go:build unix

package discovery

import (
	"os/exec"
	"syscall"
)

// setProcessGroup starts the server in its own process group so that
// killProcess also stops children (npx and uvx launch the real server as one).
func setProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func killProcess(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}
	_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	_ = cmd.Process.Kill()
}
