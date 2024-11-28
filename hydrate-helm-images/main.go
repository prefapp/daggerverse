// A generated module for HydrateHelmImages functions
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
	"dagger/hydrate-helm-images/internal/dagger"

	"gopkg.in/yaml.v3"
)

type HydrateHelmImages struct{}

type Dp struct {
	Metadata Metadata `yaml:"metadata"`
}

type Metadata struct {
	Annotations Annotations `yaml:"annotations"`
}

type Annotations struct {
	MicroService string `yaml:"firestartr.dev/microservice"`

	Image string `yaml:"firestartr.dev/image"`
}

// Returns a container that echoes whatever string argument is provided
func (m *HydrateHelmImages) BuildPreviousImages(

	ctx context.Context,

	// +optional
	// +description=Directory containing the manifests
	// +defaultPath=manifests
	manifestsDir *dagger.Directory,

) *dagger.File {

	mapImages := make(map[string]map[string]string)

	deploymentManifests, err := manifestsDir.
		Glob(ctx, "Deployment.*.yaml")

	if err != nil {

		panic(err)

	}

	for _, manifest := range deploymentManifests {

		dp := Dp{}

		content, err := manifestsDir.File(manifest).Contents(ctx)

		if err != nil {

			panic(err)

		}

		errUnms := yaml.Unmarshal([]byte(content), &dp)

		if errUnms != nil {

			panic(err)

		}

		mapImages[dp.Metadata.Annotations.MicroService] = map[string]string{"image": dp.Metadata.Annotations.Image}

	}

	marshaled, err := yaml.Marshal(mapImages)

	return dag.Directory().
		WithNewFile("previous-images.yaml", string(marshaled)).
		File("previous-images.yaml")
}
