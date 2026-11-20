package main

import (
	"context"
	"dagger/firestartr-config/internal/dagger"
	"fmt"

	"gopkg.in/yaml.v3"
)

func loadApps(
	ctx context.Context,
	firestartrDir *dagger.Directory,
	schemas *schemaProvider,
) ([]FirestartrApp, error) {
	applications := []FirestartrApp{}

	filePaths, err := listConfigFiles(ctx, firestartrDir, "apps")
	if err != nil {
		return nil, err
	}

	for _, filePath := range filePaths {
		fileContent, err := firestartrDir.File(filePath).Contents(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", filePath, err)
		}

		if _, err := schemas.validateDocument(ctx, "apps", filePath, fileContent); err != nil {
			return nil, err
		}

		app := FirestartrApp{}
		if err := yaml.Unmarshal([]byte(fileContent), &app); err != nil {
			return nil, fmt.Errorf("failed to decode %s: %w", filePath, err)
		}

		applications = append(applications, app)
	}

	return applications, nil
}

func getAppFromStateRepo(
	ctx context.Context,
	firestartrDir *dagger.Directory,
	schemas *schemaProvider,
	stateRepo string,
) (*FirestartrApp, error) {
	applications, err := loadApps(ctx, firestartrDir, schemas)
	if err != nil {
		return nil, err
	}

	for _, app := range applications {
		if app.StateRepo == stateRepo {
			return &app, nil
		}
	}

	return nil, fmt.Errorf("No app found for state repo %s", stateRepo)
}
