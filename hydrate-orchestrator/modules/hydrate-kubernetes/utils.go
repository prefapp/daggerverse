package main

import (
	"dagger/hydrate-kubernetes/internal/dagger"
)

func containerWithCmds(c *dagger.Container, commands [][]string) *dagger.Container {

	for _, cmd := range commands {

		c = c.WithExec(cmd)

	}

	return c
}
