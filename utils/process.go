//go:build !windows

package utils

import (
	"os/exec"
	"syscall"
)

func EnablePgid(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func KillPgid(cmd *exec.Cmd) error {
	if pgid, err := syscall.Getpgid(cmd.Process.Pid); err == nil {
		return syscall.Kill(-pgid, syscall.SIGKILL)
	}
	return nil
}
