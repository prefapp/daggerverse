package main

import (
	"context"
	"dagger/hydrate-kubernetes/internal/dagger"
	"encoding/json"
	"strings"
)

type App struct {
	App     string
	Cluster string
	Tenant  string
	Env     string
}

func (m *HydrateKubernetes) RenderApps(

	ctx context.Context,

	// +optional
	// +default="[]"
	affectedPaths string,

	// +optional
	// +default="{\"images\":[]}"
	newImagesMatrix string,

	app string,

) *dagger.Directory {

	affectedPathsList := []string{}

	err := json.Unmarshal([]byte(affectedPaths), &affectedPathsList)

	if err != nil {

		panic(err)

	}

	apps := getAffectedApps(affectedPathsList, app)

	for _, app := range apps {

		stdout, renderErr := m.RenderApp(
			ctx,
			app.Env,
			app.App,
			app.Cluster,
			app.Tenant,
			newImagesMatrix,
		)

		if renderErr != nil {

			panic(renderErr)

		}

		tmpDir := m.SplitRenderInFiles(ctx,
			dag.Directory().
				WithNewFile("rendered.yaml", stdout).
				File("rendered.yaml"),
		)

		m.WetRepoDir = m.WetRepoDir.
			WithoutDirectory("kubernetes/"+app.Cluster+"/"+app.Tenant+"/"+app.Env).
			WithDirectory(
				"kubernetes/"+app.Cluster+"/"+app.Tenant+"/"+app.Env,
				tmpDir,
			)
	}

	return m.WetRepoDir
}

func getAffectedApps(affectedPathsList []string, app string) map[string]App {

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
