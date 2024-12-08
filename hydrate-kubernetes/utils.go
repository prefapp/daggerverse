package main

import (
	"dagger/hydrate-kubernetes/internal/dagger"
	"encoding/json"
	"strings"

	"gopkg.in/yaml.v3"
)

func combineMaps(appsFromAffectedPaths map[string]App, appsFromInputMatrix map[string]App) map[string]App {

	apps := map[string]App{}

	for key, app := range appsFromAffectedPaths {

		apps[key] = app

	}

	for key, app := range appsFromInputMatrix {

		apps[key] = app

	}

	return apps
}

func getAffectedAppsFromAffectedPaths(affectedPathsList []string, app string) map[string]App {

	apps := map[string]App{}

	for _, affectedPath := range affectedPathsList {

		parts := splitAffectedPath(affectedPath)

		if parts == nil {

			continue

		}

		app := App{
			App:     app,
			Cluster: parts[1],
			Tenant:  parts[2],
			Env:     parts[3],
		}

		apps[app.App+app.Cluster+app.Tenant+app.Env] = app

	}

	return apps
}

func getAffectedAppFromInputMatrix(newImagesMatrix string) map[string]App {

	inputMatrix := ImageMatrix{}

	err := json.Unmarshal([]byte(newImagesMatrix), &inputMatrix)

	if err != nil {

		panic(err)

	}

	apps := map[string]App{}

	for _, image := range inputMatrix.Images {

		cluster := strings.Split(image.BaseFolder, "/")[1]

		apps[image.App+cluster+image.Tenant+image.Env] = App{
			App:     image.App,
			Cluster: cluster,
			Tenant:  image.Tenant,
			Env:     image.Env,
		}

	}

	return apps

}

func splitAffectedPath(affectedPath string) []string {

	parts := strings.Split(affectedPath, "/")

	if parts[0] != "kubernetes" {

		return nil

	}

	if len(parts) == 3 && (strings.HasSuffix(parts[2], ".yaml") || strings.HasSuffix(parts[2], ".yml")) {

		envPart := strings.Split(parts[2], ".")

		return []string{"kubernetes", parts[0], parts[1], envPart[0]}

	}

	return parts

}

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
