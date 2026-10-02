package caddyproc

import (
	"errors"
	"os/exec"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

var ErrNotFound = errors.New("caddy process not found")

func FindPID() (int, error) {
	out, err := exec.Command("pgrep", "-x", "caddy").Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 1 {
			return 0, ErrNotFound
		}
		return 0, err
	}
	for _, line := range strings.Fields(string(out)) {
		if pid, err := strconv.Atoi(line); err == nil {
			return pid, nil
		}
	}
	return 0, ErrNotFound
}

func SignalReload(pid int) error {
	// see https://caddyserver.com/docs/command-line#signals
	return unix.Kill(pid, unix.SIGUSR1)
}
