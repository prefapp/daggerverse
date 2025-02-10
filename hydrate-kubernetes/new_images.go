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

) (*dagger.File, error) {

	var imageMatrix ImageMatrix

	json.Unmarshal([]byte(matrix), &imageMatrix)

	mapNewImages := make(map[string]map[string]string)

	if len(imageMatrix.Images) > 1 {

		panic("Only one image per service is allowed")

	}

	for _, imageData := range imageMatrix.Images {

		for _, serviceName := range imageData.ServiceNameList {

			mapNewImages[serviceName] = map[string]string{"image": imageData.Image}

		}

	}

	marshaled, err := yaml.Marshal(mapNewImages)

	if err != nil {

		return nil, err

	}

	return dag.Directory().
		WithNewFile("current-images.yaml", string(marshaled)).
		File("current-images.yaml"), nil
}
