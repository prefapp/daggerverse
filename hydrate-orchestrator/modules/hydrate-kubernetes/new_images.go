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

	if len(imageMatrix.Images) == 0 {

		return dag.Directory().
			WithNewFile("current-images.yaml", "{}").
			File("current-images.yaml"), nil

	}

	mapNewImages := make(map[string]map[string]string)
	seenServices := make(map[string]bool)

	for _, imageData := range imageMatrix.Images {

		if len(imageData.ServiceNameList) == 0 && len(imageData.ImageKeys) == 0 {

			return nil, fmt.Errorf("service_names and image_keys cannot be empty")

		}

		if len(imageData.ServiceNameList) > 0 && len(imageData.ImageKeys) > 0 {

			return nil, fmt.Errorf("service_names and image_keys cannot be used together")

		}

		if len(imageData.ServiceNameList) > 0 {

			for _, serviceName := range imageData.ServiceNameList {

				if seenServices[serviceName] {

					return nil, fmt.Errorf("duplicate service name %s in matrix", serviceName)

				}

				seenServices[serviceName] = true

				mapNewImages[serviceName] = map[string]string{"image": imageData.Image}

			}

		}

		if len(imageData.ImageKeys) > 0 {

			// image_keys are only supported for a single image; if multiple images are present we error

			if len(imageMatrix.Images) > 1 {

				return nil, fmt.Errorf("image_keys not supported for multiple images in the matrix")

			}

			jsonObj := "{}"

			for _, imageKey := range imageData.ImageKeys {

				jsonObj = m.GenerateOjectFromPath(imageKey, imageData.Image, jsonObj)

			}

			yamlObj, err := yamlsigs.JSONToYAML([]byte(jsonObj))

			if err != nil {

				return nil, err

			}

			return dag.Directory().
				WithNewFile("current-images.yaml", string(yamlObj)).
				File("current-images.yaml"), nil

		}

	}

	imagesYaml := "{}"

	if len(mapNewImages) > 0 {

		marshaled, err := yaml.Marshal(mapNewImages)

		if err != nil {

			return nil, err

		}

		imagesYaml = string(marshaled)

	}

	return dag.Directory().
		WithNewFile("current-images.yaml", string(imagesYaml)).
		File("current-images.yaml"), nil
}
