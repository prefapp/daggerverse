package main

import (
	"context"
	"dagger/validate-crds/internal/dagger"
)

type ValidateCrds struct {
	Kind *dagger.Container
	Crds *dagger.Directory
}

type Version string

func New(
	ctx context.Context,

	//This is required by the "prefapp/daggerverse/kind" module.
	// +required
	dockerSocket *dagger.Socket,

	//This is required by the "prefapp/daggerverse/kind" module.
	// +required
	kindSvc *dagger.Service,

	//+required
	crdsDir *dagger.Directory,

	// From the "prefapp/daggerverse/kind" module.
	// The Kubernetes version you want to install in the Kind cluster. Has to be
	// one of the available ones in the current Kind version used.
	// +optional
	version dagger.KindVersion,

) *ValidateCrds {

	opts := dagger.KindOpts{}

	if version != "" {
		opts.Version = version
	}

	container := dag.Kind(dockerSocket, kindSvc, opts).Container()

	return &ValidateCrds{
		Kind: container,
		Crds: crdsDir,
	}
}

func (m *ValidateCrds) Validate(ctx context.Context) (string, error) {
	return m.Kind.
		WithWorkdir("/crds").
		WithMountedDirectory("/crds", m.Crds).
		WithExec([]string{"kubectl", "apply", "--dry-run=client", "-f", "."}).
		Stdout(ctx)
}
