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
)

type HydrateKubernetes struct {
	HelmfileImageTag string
	Container        *dagger.Container
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

) *HydrateKubernetes {

	return &HydrateKubernetes{

		Container: dag.
			Container().
			From(helmfileImage + ":" + helmfileImageTag),
	}
}

// Returns a container that echoes whatever string argument is provided
func (m *HydrateKubernetes) Render(

	affectedPaths string,

	repoDir *dagger.Directory,

	wetRepoDir *dagger.Directory,

	authDir *dagger.Directory,

) *dagger.Directory {

	return dag.Directory()
}
