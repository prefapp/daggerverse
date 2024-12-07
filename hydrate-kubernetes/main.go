package main

import (
	"context"
	"dagger/hydrate-kubernetes/internal/dagger"

	"gopkg.in/yaml.v3"
)

type HydrateKubernetes struct {
	Container    *dagger.Container
	ValuesDir    *dagger.Directory
	WetRepoDir   *dagger.Directory
	Helmfile     *dagger.File
	ValuesGoTmpl *dagger.File
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
	}
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

// HydrateKubernetes hydrates the wet manifests with the helm values
func (m *HydrateKubernetes) Render(

	// Json string of the affected paths
	// Format: ["path/to/file1", "path/to/file2"]
	// optional
	// +default="[]"
	affectedPaths string,

	// The path to the values repo, where helm values are stored
	// +required
	valuesRepoDir *dagger.Directory,

	// The path to the wet repo, where the wet manifests are stored
	// +required
	wetRepoDir *dagger.Directory,

	// The path to auth files, which will contain the helm login credentials
	//
	// For azure:
	//	<authDir>/az/helmfile.user
	//	<authDir>/az/helmfile.password
	//	<authDir>/az/helmfile.repository
	//
	// For aws:
	//	<authDir>/aws/helmfile.user
	//	<authDir>/aws/helmfile.password
	//	<authDir>/aws/helmfile.repository
	//
	// +required
	authDir *dagger.Directory,

) *dagger.Directory {

	return dag.Directory()
}
