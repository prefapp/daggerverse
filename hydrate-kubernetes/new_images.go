package main

import (
	"context"
	"dagger/hydrate-kubernetes/internal/dagger"
	"encoding/json"

	"gopkg.in/yaml.v3"
)

func (m *HydrateKubernetes) BuildNewImages(

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
