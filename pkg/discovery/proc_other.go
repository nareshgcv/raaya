//go:build !unix

package discovery

import "os/exec"

func setProcessGroup(*exec.Cmd) {}

func killProcess(cmd *exec.Cmd) {
	if cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
}
