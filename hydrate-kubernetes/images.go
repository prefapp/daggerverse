package main

import (
	"context"
	"dagger/hydrate-kubernetes/internal/dagger"
	"slices"
	"strings"
)

func (m *HydrateKubernetes) GetImagesFile(ctx context.Context, cluster string, tenant string, env string) *dagger.File {

	entries, err := m.ValuesDir.Glob(ctx, "kubernetes/*/*/*/*")

	if err != nil {
		panic(err)
	}

	targetDir := strings.Join([]string{"kubernetes", cluster, tenant, env}, "/")

	for _, ext := range []string{".yaml", ".yml"} {

		if slices.Contains(entries, targetDir+"/images"+ext) {

			return m.ValuesDir.File(targetDir + "/images" + ext)

		}

	}

	return dag.Directory().
		WithNewFile("images.yaml", "{}").
		File("images.yaml")

}