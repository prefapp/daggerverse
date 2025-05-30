package main

import (
	"context"
	"dagger/hydrate-kubernetes/internal/dagger"
	"fmt"
)

func hydrateConfigFileExists(
	ctx context.Context,
	valuesDir *dagger.Directory,
) (bool, error) {

	entries, err := valuesDir.Glob(ctx, ".github/hydrate_k8s_config.yaml")

	if err != nil {
		return false, fmt.Errorf("failed to check for hydrate_k8s_config.yaml: %w", err)
	}

	entriesLength := len(entries)
	if len(entries) == 0 {
		return false, nil
	} else if entriesLength >= 1 {
		return true, nil
	} else {
		return false, fmt.Errorf(
			"unexpected number of hydrate_k8s_config.yaml files found: %d",
			entriesLength,
		)
	}
}
