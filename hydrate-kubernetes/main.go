// A generated module for HydrateKubernetes functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
	"dagger/hydrate-kubernetes/internal/dagger"
	"time"
)

type HydrateKubernetes struct {
	Container  *dagger.Container
	ValuesDir  *dagger.Directory
	WetRepoDir *dagger.Directory
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
	// +default=[]
	addPackage []string,

) *HydrateKubernetes {

	c := dag.
		Container().
		From(helmfileImage + ":" + helmfileImageTag)

	for _, pkg := range addPackage {

		c = c.WithExec([]string{"apk", "add", pkg})

	}

	return &HydrateKubernetes{

		Container: c,

		ValuesDir: valuesDir,

		WetRepoDir: wetRepoDir,
	}
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

func (m *HydrateKubernetes) RenderApp(

	env string,

	app string,

	cluster string,

	tenant string,

) *dagger.Container {

	m.Container = m.Container.
		WithDirectory("/values", m.ValuesDir).
		WithWorkdir("/values").
		WithMountedFile(
			"/values/helmfile.yaml",
			dag.CurrentModule().Source().File("helm/helmfile.yaml")).
		WithMountedFile(
			"/values/values.yaml.gotmpl",
			dag.CurrentModule().Source().File("helm/values.yaml.gotmpl")).
		WithEnvVariable("BUST", time.Now().String()).
		WithExec([]string{
			"helmfile",
			"-e",
			env,
			"template",
			"--state-values-set-string",
			"tenant=" + tenant + ",app=" + app + ",cluster=" + cluster,
			"--state-values-file",
			"./kubernetes/" + cluster + "/" + tenant + "/" + env + ".yaml",
			"--debug",
		})

	return m.Container
}
