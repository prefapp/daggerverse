package main

import (
	"context"
	"dagger/hydrate-kubernetes/internal/dagger"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

func (m *HydrateKubernetes) SplitRenderInFiles(

	ctx context.Context,

	renderFile *dagger.File,

) *dagger.Directory {

	content, err := renderFile.Contents(ctx)

	if err != nil {

		panic(err)

	}

	k8sManifests := strings.Split(string(content), "\n---\n")

	dir := dag.Directory()

	for _, manifest := range k8sManifests {

		k8sresource := KubernetesResource{}

		err := yaml.Unmarshal([]byte(manifest), &k8sresource)

		if err != nil {

			panic(err)

		}

		// create a new file for each k8s manifest
		// with <kind>.<metadata.name>.yml as the file name

		fileName := fmt.Sprintf("%s.%s.yml", k8sresource.Kind, k8sresource.Metadata.Name)

		dir = dir.WithNewFile(fileName, manifest)

	}

	return dir
}
