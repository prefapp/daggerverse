package main

import (
	"context"
	"dagger/hydrate-helm-images/internal/dagger"

	"encoding/json"

	"gopkg.in/yaml.v3"
)

type HydrateHelmImages struct{}

func (m *HydrateHelmImages) BuildPreviousImages(

	ctx context.Context,

	manifestsDir *dagger.Directory,

) *dagger.File {

	mapImages := getDeploymentMap(ctx, manifestsDir)

	marshaled, err := yaml.Marshal(mapImages)

	if err != nil {

		panic(err)

	}

	return dag.Directory().
		WithNewFile("previous-images.yaml", string(marshaled)).
		File("previous-images.yaml")
}

func getDeploymentMap(

	ctx context.Context,

	manifestsDir *dagger.Directory,

) map[string]map[string]string {

	mapImages := make(map[string]map[string]string)

	deploymentManifests, err := manifestsDir.
		Glob(ctx, "Deployment.*.yml")

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

func (m *HydrateHelmImages) BuildCurrentImages(

	ctx context.Context,

	matrix string,

) *dagger.File {

	var imageMatrix ImageMatrix

	json.Unmarshal([]byte(matrix), &imageMatrix)

	mapNewImages := make(map[string]map[string]string)

	for _, imageData := range imageMatrix.Images {

		for _, serviceName := range imageData.ServiceNameList {

			mapNewImages[serviceName] = map[string]string{"image": imageData.Image}

		}

	}

	marshaled, errMars := yaml.Marshal(mapNewImages)

	if errMars != nil {

		panic(errMars)

	}

	return dag.Directory().
		WithNewFile("current-images.yaml", string(marshaled)).
		File("current-images.yaml")
}
