package main

import (
	"context"
	"dagger/firestartr-config/internal/dagger"
	"fmt"

	"gopkg.in/yaml.v3"
)

type FirestartrApp struct {
	Name      string    `yaml:"name"`
	StateRepo string    `yaml:"state_repo"`
	Services  []Service `yaml:"image_types"`
}

type Service struct {
	Repo         string   `yaml:"repo"`
	ServiceNames []string `yaml:"service_names"`
}

func loadApps(ctx context.Context, firestartrDir *dagger.Directory) ([]FirestartrApp, error) {
	applications := []FirestartrApp{}

	for _, ext := range []string{".yaml", ".yml"} {
		filePaths, err := firestartrDir.Glob(ctx, "apps/*"+ext)

		if err != nil {
			return nil, err
		}

		for _, filePath := range filePaths {
			fileContent, err := firestartrDir.File(filePath).Contents(ctx)

			if err != nil {
				return nil, err
			}

			app := FirestartrApp{}

			err = yaml.Unmarshal([]byte(fileContent), &app)
			if err != nil {
				return nil, err
			}

			applications = append(applications, app)
		}
	}

	return applications, nil
}

func getAppFromStateRepo(
	ctx context.Context,
	firestartrDir *dagger.Directory,
	stateRepo string,
) (*FirestartrApp, error) {
	applications, err := loadApps(ctx, firestartrDir)

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
