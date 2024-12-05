package main

import (
	"context"
	"dagger/hydrate-kubernetes/internal/dagger"
	"encoding/json"

	"gopkg.in/yaml.v3"
)

func (m *HydrateKubernetes) BuildPreviousImages(

	ctx context.Context,

	updatedDeployments string,

	repoDir *dagger.Directory,

) *dagger.File {

	unmarshaledDeps := []string{}

	err := json.Unmarshal([]byte(updatedDeployments), &unmarshaledDeps)

	if err != nil {

		panic(err)

	}

	dir := dag.Directory()

	for _, dep := range unmarshaledDeps {

		imagesContent := m.BuildPreviousImagesForApp(ctx, repoDir.Directory(dep), dir, dep)

		dir = dir.
			WithNewDirectory(dep, dagger.DirectoryWithNewDirectoryOpts{
				Permissions: 0755,
			}).
			WithNewFile(
				dep+"/previous_images.yaml",
				imagesContent,
				dagger.DirectoryWithNewFileOpts{Permissions: 0644},
			)

	}

	return dir.File("previous_images.yaml")
}

func (m *HydrateKubernetes) BuildPreviousImagesForApp(

	ctx context.Context,

	manifestsDir *dagger.Directory,

	outputDir *dagger.Directory,

	app string,

) string {

	mapImages := getDeploymentMap(ctx, manifestsDir)

	marshaled, err := yaml.Marshal(mapImages)

	if err != nil {

		panic(err)

	}
	return string(marshaled)

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

func (m *HydrateKubernetes) BuildCurrentImages(

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
