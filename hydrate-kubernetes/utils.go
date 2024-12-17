package main

import (
	"dagger/hydrate-kubernetes/internal/dagger"

	"gopkg.in/yaml.v3"
)

func installDeps(depsFileContent string, c *dagger.Container) *dagger.Container {

	deps := DepsFile{}

	err := yaml.Unmarshal([]byte(depsFileContent), &deps)

	if err != nil {

		panic(err)

	}

	for _, pkg := range deps.Dependencies {

		c = c.WithExec([]string{"apk", "add", pkg})

	}

	return c
}
