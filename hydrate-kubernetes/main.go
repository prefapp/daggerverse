package main

import (
	"context"
	"dagger/hydrate-kubernetes/internal/dagger"
	"encoding/json"
)

type HydrateKubernetes struct {
	Container               *dagger.Container
	ValuesDir               *dagger.Directory
	WetRepoDir              *dagger.Directory
	Helmfile                *dagger.File
	ValuesGoTmpl            *dagger.File
	HelmRegistryLoginNeeded bool
	HelmRegistry            string
	HelmRegistryUser        string
	HelmRegistryPassword    *dagger.Secret
}

func New(

	ctx context.Context,

	// The Helmfile image tag to use https://github.com/helmfile/helmfile/pkgs/container/helmfile
	// +optional
	// +default="latest"
	helmfileImageTag string,

	// The Helmfile image to use
	// +optional
	// +default="ghcr.io/helmfile/helmfile"
	helmfileImage string,

	// The path to the values directory, where the helm values are stored
	valuesDir *dagger.Directory,

	// The path to the wet repo directory, where the wet manifests are stored
	// +optional
	wetRepoDir *dagger.Directory,

	// extra packages to install
	// +optional
	depsFile *dagger.File,

	// The path to the helmfile.yaml file
	// +optional
	helmfile *dagger.File,

	// The path to the values.go.tmpl file
	// +optional
	valuesGoTmpl *dagger.File,

	// +optional
	// +default=false
	helmRegistryLoginNeeded bool,

	// +optional
	helmRegistry string,

	// +optional
	helmRegistryUser string,

	// +optional
	helmRegistryPassword *dagger.Secret,

) *HydrateKubernetes {

	c := dag.
		Container().
		From(helmfileImage + ":" + helmfileImageTag)

	depsFileContent, err := depsFile.Contents(ctx)

	if err != nil {

		panic(err)

	}

	c = installDeps(depsFileContent, c)

	if helmfile == nil {

		helmfile = dag.CurrentModule().Source().File("./helm/helmfile.yaml")

	}

	if valuesGoTmpl == nil {

		valuesGoTmpl = dag.CurrentModule().Source().File("./helm/values.yaml.gotmpl")

	}

	return &HydrateKubernetes{

		Container: c,

		ValuesDir: valuesDir,

		WetRepoDir: wetRepoDir,

		Helmfile: helmfile,

		ValuesGoTmpl: valuesGoTmpl,

		HelmRegistryLoginNeeded: helmRegistryLoginNeeded,

		HelmRegistry: helmRegistry,

		HelmRegistryUser: helmRegistryUser,

		HelmRegistryPassword: helmRegistryPassword,
	}
}

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

		renderedChartFile, renderErr := m.RenderApp(
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
				WithNewFile("rendered.yaml", renderedChartFile).
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
