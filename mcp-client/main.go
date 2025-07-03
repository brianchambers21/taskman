package main

import (
	"os"
	"os/exec"
)

func main() {
	// Delegate to the actual client binary
	cmd := exec.Command("go", append([]string{"run", "./cmd/client"}, os.Args[1:]...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	
	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		}
		os.Exit(1)
	}
}