package main

import (
	"context"
	"dagger/hydrate-helm-images/internal/dagger"

	"encoding/json"

	"gopkg.in/yaml.v3"
)

// YAML Types
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

// JSON Types
type ImageData struct {
	Tenant           string
	App              string
	Env              string
	ServiceNameList  []string
	Image            string
	Reviewers        []string
	BaseFolder       string
	RepositoryCaller string
}

type ImageMatrix struct {
	Images []ImageData
}

func (m *HydrateHelmImages) GetDeploymentMap(

	ctx context.Context,

	manifestsDir *dagger.Directory,

) map[string]map[string]string {

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

	return mapImages

}

func (m *HydrateHelmImages) BuildPreviousImages(

	ctx context.Context,

	// +optional
	// +description=Directory containing the manifests
	// +defaultPath=manifests
	manifestsDir *dagger.Directory,

) *dagger.File {

	mapImages := m.GetDeploymentMap(ctx, manifestsDir)

	marshaled, err := yaml.Marshal(mapImages)

	if err != nil {

		panic(err)

	}

	return dag.Directory().
		WithNewFile("previous-images.yaml", string(marshaled)).
		File("previous-images.yaml")
}

func (m *HydrateHelmImages) BuildCurrentImages(

	ctx context.Context,

	matrix string,

	// +optional
	// +description=Directory containing the manifests
	// +defaultPath=manifests
	manifestsDir *dagger.Directory,

) *dagger.File {
	var imageMatrix ImageMatrix

	json.Unmarshal([]byte(matrix), &imageMatrix)

	mapNewImages := make(map[string]map[string]string)

	for _, imageData := range imageMatrix.Images {

		for _, serviceName := range imageData.ServiceNameList {

			mapNewImages[serviceName] = map[string]string{"image": imageData.Image}

		}

	}

	mapOldImages := m.GetDeploymentMap(ctx, manifestsDir)

	for key, value := range mapNewImages {

		mapOldImages[key] = value

	}

	marshaled, errMars := yaml.Marshal(mapOldImages)

	if errMars != nil {

		panic(errMars)

	}

	return dag.Directory().
		WithNewFile("current-images.yaml", string(marshaled)).
		File("current-images.yaml")
}
