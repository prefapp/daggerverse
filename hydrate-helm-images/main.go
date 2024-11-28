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
