package main

import (
	"context"
	"dagger/hydrate-kubernetes/internal/dagger"
	"encoding/json"
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

	appsFromAffectedPaths := getAffectedAppsFromAffectedPaths(affectedPathsList, app)

	appsFromInputMatrix := getAffectedAppFromInputMatrix(newImagesMatrix)

	apps := combineMaps(appsFromAffectedPaths, appsFromInputMatrix)

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
