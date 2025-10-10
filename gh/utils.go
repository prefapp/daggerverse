package main

import (
	"fmt"
	"gh/internal/dagger"
)

func extractErrorMessage(err error) string {
	switch e := err.(type) {
	case *dagger.ExecError:
		return fmt.Sprintf(`GH execution failed: %s
STDERR: %s
STDOUT: %s`, e.Error(), e.Stderr, e.Stdout)
	default:
		return fmt.Sprintf("Failed: %s", err.Error())
	}
}
