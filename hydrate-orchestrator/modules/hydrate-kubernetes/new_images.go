package main

import (
	"context"
	"dagger/hydrate-kubernetes/internal/dagger"
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
