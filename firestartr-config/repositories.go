package main

import (
	"context"
	"dagger/firestartr-config/internal/dagger"
	"fmt"

	"gopkg.in/yaml.v3"
)

func loadRepositories(ctx context.Context, firestartrDir *dagger.Directory, schemas *schemaProvider) ([]Repository, error) {
	repositories := []Repository{}

	filePaths, err := listConfigFiles(ctx, firestartrDir, "helm_repositories")
	if err != nil {
		return nil, err
	}

	for _, filePath := range filePaths {
		fileContent, err := firestartrDir.File(filePath).Contents(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", filePath, err)
		}

		if _, err := schemas.validateDocument(ctx, "helm_repositories", filePath, fileContent); err != nil {
			return nil, err
		}

		repo := Repository{}
		if err := yaml.Unmarshal([]byte(fileContent), &repo); err != nil {
			return nil, fmt.Errorf("failed to decode %s: %w", filePath, err)
		}

		repositories = append(repositories, repo)
	}

	return repositories, nil
}
