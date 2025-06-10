package main

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

func (m *HydrateKubernetes) BuildPreviousImagesApp(

	ctx context.Context,

	cluster string,

	tenant string,

	env string,

) (string, error) {

	entries, err := m.WetRepoDir.Glob(ctx, "kubernetes/*/*/*")

	if err != nil {
		return "", err
	}

	targetDir := strings.Join([]string{"kubernetes", cluster, tenant, env}, "/") + "/"

	if !slices.Contains(entries, targetDir) {

		return "{}", nil

	}

	manifestsDir := m.WetRepoDir.Directory(targetDir)

	mapImages := make(map[string]map[string]string)

	deploymentManifests := []string{}

	for _, regex := range []string{"*.*.yml", "*.*.yaml"} {

		manifests, err := manifestsDir.Glob(ctx, regex)

		if err != nil {

			return "", err

		}

		deploymentManifests = append(deploymentManifests, manifests...)

	}

	for _, manifest := range deploymentManifests {

		artifact := Artifact{}

		content, err := manifestsDir.File(manifest).Contents(ctx)

		if err != nil {

			return "", err

		}

		err = yaml.Unmarshal([]byte(content), &artifact)

		if err != nil {

			return "", err

		}

		if mapImages[artifact.Metadata.Annotations.MicroService] != nil {

			return "", fmt.Errorf("duplicate microservice found: %s", artifact.Metadata.Annotations.MicroService)

		}

		if artifact.Metadata.Annotations.MicroService != "" {

			mapImages[artifact.Metadata.Annotations.MicroService] = map[string]string{"image": artifact.Metadata.Annotations.Image}

		}

	}

	marshaled, err := yaml.Marshal(mapImages)

	if err != nil {

		return "", err

	}

	return string(marshaled), nil

}
