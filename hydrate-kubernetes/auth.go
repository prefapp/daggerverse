package main

import (
	"context"
	"dagger/hydrate-kubernetes/internal/dagger"
)

func prepareHelmLogin(

	ctx context.Context,

	ctr *dagger.Container,

	helmConfigDir *dagger.Directory,

) *dagger.Container {

	return ctr.WithDirectory("/root/.config/helm", helmConfigDir)
}
