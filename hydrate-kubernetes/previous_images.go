package main

import (
	"context"
	"dagger/hydrate-kubernetes/internal/dagger"
	"fmt"

	"gopkg.in/yaml.v3"
)

func (m *HydrateKubernetes) BuildPreviousImagesApp(

	ctx context.Context,

	manifestsDir *dagger.Directory,

) string {

	mapImages := make(map[string]map[string]string)

	deploymentManifests := []string{}

	for _, regex := range []string{"*.*.yml", "*.*.yaml"} {

		manifests, err := manifestsDir.Glob(ctx, regex)

		if err != nil {

			panic(err)

		}

		deploymentManifests = append(deploymentManifests, manifests...)

	}

	for _, manifest := range deploymentManifests {

		artifact := Artifact{}

		content, err := manifestsDir.File(manifest).Contents(ctx)

		if err != nil {

			panic(err)

		}

		errUnms := yaml.Unmarshal([]byte(content), &artifact)

		if errUnms != nil {

			panic(err)

		}

		if mapImages[artifact.Metadata.Annotations.MicroService] != nil {

			panic(fmt.Sprintf("Duplicate microservice found: %s", artifact.Metadata.Annotations.MicroService))

		}

		if artifact.Metadata.Annotations.MicroService != "" {

			mapImages[artifact.Metadata.Annotations.MicroService] = map[string]string{"image": artifact.Metadata.Annotations.Image}

		}

	}

	marshaled, err := yaml.Marshal(mapImages)

	if err != nil {

		panic(err)

	}

	return string(marshaled)

}
