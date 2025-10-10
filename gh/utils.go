package main

import (
	"fmt"
	"gh/internal/dagger"
)

func extractErrorMessage(err error) string {
	switch e := err.(type) {
	case *dagger.ExecError:
		return fmt.Sprintf("Command failed:\nSTDERR: %s\nSTDOUT: %s", e.Stderr, e.Stdout)
	default:
		return fmt.Sprintf("Dagger failed: %s", err.Error())
	}
}
