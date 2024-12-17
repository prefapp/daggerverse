package main

import (
	"context"
	"dagger/hydrate-kubernetes/internal/dagger"
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

	depsFileContent, err := valuesDir.File(".github/hydrate_deps.yaml").Contents(ctx)

	if err == nil {

		c = installDeps(depsFileContent, c)
	}

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

func (m *HydrateKubernetes) Render(

	ctx context.Context,

	app string,

	cluster string,

	tenant string,

	env string,

	// +optional
	// +default="{\"images\":[]}"
	newImagesMatrix string,

) *dagger.Directory {

	renderedChartFile, renderErr := m.RenderApp(
		ctx,
		env,
		app,
		cluster,
		tenant,
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
		WithoutDirectory("kubernetes/"+cluster+"/"+tenant+"/"+env).
		WithDirectory(
			"kubernetes/"+cluster+"/"+tenant+"/"+env,
			tmpDir,
		)

	return m.WetRepoDir
}
