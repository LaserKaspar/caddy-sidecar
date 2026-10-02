package main

import (
	"errors"
	"fmt"
	"os"

	"main/internal/caddyproc"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] != "--reload" {
		fmt.Fprintln(os.Stderr, "Usage: caddy --reload")
		os.Exit(2)
	}

	// find caddy process by name
	pid, err := caddyproc.FindPID()
	if err != nil {
		if errors.Is(err, caddyproc.ErrNotFound) {
			fmt.Fprintln(os.Stderr, "caddy process not found")
		} else {
			fmt.Fprintln(os.Stderr, "process lookup failed:", err)
		}
		os.Exit(1)
	}

	// send SIGUSR1 to reload
	if err := caddyproc.SignalReload(pid); err != nil {
		fmt.Fprintf(os.Stderr, "failed to send SIGUSR1: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("reload signal sent to caddy (PID %d)\n", pid)
}
