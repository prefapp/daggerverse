package main

import (
	"dagger/hydrate-kubernetes/internal/dagger"
)

func prepareHelmLogin(

	ctr *dagger.Container,

	helmConfigDir *dagger.Directory,

) *dagger.Container {

	return ctr.WithDirectory("/root/.config/helm", helmConfigDir)
}
