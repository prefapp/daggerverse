package main

import (
	"context"
	"dagger/hydrate-kubernetes/internal/dagger"
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
	yamlsigs "sigs.k8s.io/yaml"
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

	} else if len(imageMatrix.Images) == 0 {

		return dag.Directory().
			WithNewFile("current-images.yaml", "{}").
			File("current-images.yaml"), nil

	}

	imagesYaml := "{}"

	imageData := imageMatrix.Images[0]

	if len(imageData.ServiceNameList) == 0 && len(imageData.ImageKeys) == 0 {

		return nil, fmt.Errorf("service_names and image_keys cannot be empty")

	}

	if len(imageData.ServiceNameList) > 0 && len(imageData.ImageKeys) > 0 {

		return nil, fmt.Errorf("service_names and image_keys cannot be used together")
	}

	if len(imageData.ServiceNameList) > 0 {

		for _, serviceName := range imageData.ServiceNameList {

			mapNewImages[serviceName] = map[string]string{"image": imageData.Image}

		}

		marshaled, err := yaml.Marshal(mapNewImages)

		if err != nil {

			return nil, err

		}

		imagesYaml = string(marshaled)
	}

	if len(imageData.ImageKeys) > 0 {

		jsonObj := "{}"

		for _, imageKey := range imageData.ImageKeys {

			jsonObj = m.GenerateOjectFromPath(imageKey, imageData.Image, jsonObj)

		}

		yamlObj, err := yamlsigs.JSONToYAML([]byte(jsonObj))

		if err != nil {

			return nil, err

		}

		imagesYaml = string(yamlObj)

	}

	return dag.Directory().
		WithNewFile("current-images.yaml", string(imagesYaml)).
		File("current-images.yaml"), nil
}
