package main

import (
	"context"
	"dagger/hydrate-kubernetes/internal/dagger"
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

func (m *HydrateKubernetes) BuildNewImages(
	ctx context.Context,
	matrix string,
) (*dagger.File, error) {
	contents, err := buildNewImagesYAML(matrix)
	if err != nil {
		return nil, err
	}
	return dag.Directory().
		WithNewFile("current-images.yaml", contents).
		File("current-images.yaml"), nil
}

func buildNewImagesYAML(matrix string) (string, error) {
	var imageMatrix ImageMatrix
	if err := json.Unmarshal([]byte(matrix), &imageMatrix); err != nil {
		return "", err
	}
	values := make(map[string]interface{})
	seenServices := make(map[string]bool)
	for _, image := range imageMatrix.Images {
		if len(image.ServiceNameList) == 0 && len(image.ImageKeys) == 0 {
			return "", fmt.Errorf("service_names and image_keys cannot be empty")
		}
		if len(image.ServiceNameList) > 0 && len(image.ImageKeys) > 0 {
			return "", fmt.Errorf("service_names and image_keys cannot be used together")
		}
		for _, service := range image.ServiceNameList {
			if seenServices[service] {
				return "", fmt.Errorf("duplicate service name %s in matrix", service)
			}
			seenServices[service] = true
			if err := insertImageValue(values, []string{service, "image"}, image.Image); err != nil {
				return "", err
			}
		}
		for _, key := range image.ImageKeys {
			if !strings.HasPrefix(key, "/") {
				return "", fmt.Errorf("image key %q must be a JSON pointer starting with /", key)
			}
			parts := strings.Split(key[1:], "/")
			for i, part := range parts {
				for j := 0; j < len(part); j++ {
					if part[j] == '~' {
						if j+1 == len(part) || (part[j+1] != '0' && part[j+1] != '1') {
							return "", fmt.Errorf("invalid JSON pointer escape in image key %q", key)
						}
						j++
					}
				}
				parts[i] = strings.ReplaceAll(strings.ReplaceAll(part, "~1", "/"), "~0", "~")
			}
			if err := insertImageValue(values, parts, image.Image); err != nil {
				return "", fmt.Errorf("image key %q: %w", key, err)
			}
		}
	}
	contents, err := yaml.Marshal(values)
	return string(contents), err
}

func insertImageValue(values map[string]interface{}, path []string, image string) error {
	current := values
	for i, key := range path {
		if i == len(path)-1 {
			if _, exists := current[key]; exists {
				return fmt.Errorf("conflicting image target %q", strings.Join(path, "/"))
			}
			current[key] = image
			return nil
		}
		child, exists := current[key]
		if !exists {
			child = make(map[string]interface{})
			current[key] = child
		}
		nested, ok := child.(map[string]interface{})
		if !ok {
			return fmt.Errorf("conflicting image target %q", strings.Join(path, "/"))
		}
		current = nested
	}
	return nil
}
