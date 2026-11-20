package main

import (
	"context"
	"dagger/firestartr-config/internal/dagger"
	"fmt"

	"gopkg.in/yaml.v3"
)

func loadRegistries(ctx context.Context, firestartrDir *dagger.Directory, schemas *schemaProvider) ([]Registry, error) {
	registries := []Registry{}

	filePaths, err := listConfigFiles(ctx, firestartrDir, "docker_registries")
	if err != nil {
		return nil, err
	}

	for _, filePath := range filePaths {
		fileContent, err := firestartrDir.File(filePath).Contents(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", filePath, err)
		}

		if _, err := schemas.validateDocument(ctx, "docker_registries", filePath, fileContent); err != nil {
			return nil, err
		}

		reg := Registry{}
		if err := yaml.Unmarshal([]byte(fileContent), &reg); err != nil {
			return nil, fmt.Errorf("failed to decode %s: %w", filePath, err)
		}

		registries = append(registries, reg)
	}

	return registries, nil
}
