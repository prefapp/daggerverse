package main

import (
	"context"
	"dagger/validate-crds/internal/dagger"
)

type ValidateCrds struct {
	Kind *dagger.Container
	Crds *dagger.Directory
	Crs *dagger.Directory
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

	// The directory with the CRDs to create.
	//+required
	crdsDir *dagger.Directory,

	// A directory that can have nested directories with files that create CRs
	// from the CRDs.
	//+required
	crsDir *dagger.Directory,

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
		Crs: crsDir,
	}
}

func (m *ValidateCrds) CreateCRS(
	ctx context.Context,
	ctr *dagger.Container,
) (string, error) {
	return ctr.
		WithWorkdir("/crs").
		WithMountedDirectory("/crs", m.Crs).
		WithExec([]string{"kubectl", "apply", "-R", "-f", "."}).Stdout(ctx)
}

func (m *ValidateCrds) CreateCRDS(ctx context.Context) *dagger.Container {
	return m.Kind.
		WithWorkdir("/crds").
		WithMountedDirectory("/crds", m.Crds).
		WithExec([]string{"kubectl", "apply", "-f", "."})
}

func (m *ValidateCrds) Validate(ctx context.Context) (string, error) {
	ctr := m.CreateCRDS(ctx)

	_, err := ctr.Stdout(ctx)
	if err != nil {
		return "", err
	}

	return m.CreateCRS(ctx, ctr)
}
