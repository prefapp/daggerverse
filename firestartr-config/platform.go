package main

import (
	"context"
	"dagger/firestartr-config/internal/dagger"
	"fmt"

	"gopkg.in/yaml.v3"
)

func loadPlatforms(ctx context.Context, firestartrDir *dagger.Directory, schemas *schemaProvider) ([]Platform, error) {
	platforms := []Platform{}

	filePaths, err := listConfigFiles(ctx, firestartrDir, "platforms")
	if err != nil {
		return nil, err
	}

	for _, filePath := range filePaths {
		fileContent, err := firestartrDir.File(filePath).Contents(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", filePath, err)
		}

		if _, err := schemas.validateDocument(ctx, "platforms", filePath, fileContent); err != nil {
			return nil, err
		}

		p := Platform{}
		if err := yaml.Unmarshal([]byte(fileContent), &p); err != nil {
			return nil, fmt.Errorf("failed to decode %s: %w", filePath, err)
		}

		platforms = append(platforms, p)
	}

	return platforms, nil
}

func (m *FirestartrConfig) FindPlatformByName(name string) *Platform {
	for _, platform := range m.Platforms {
		if platform.Name == name {
			return &platform
		}
	}
	return nil
}
