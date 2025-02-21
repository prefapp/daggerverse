package main

import (
	"context"
	"dagger/firestartr-config/internal/dagger"

	"gopkg.in/yaml.v3"
)

type Platform struct {
	// Cannot use `Type` as a field name, it seems to be a reserved keyword for Dagger.io
	PlatformType string   `yaml:"type"`
	Name         string   `yaml:"name"`
	Tenants      []string `yaml:"tenants"`
	Envs         []string `yaml:"envs"`
}

func loadPlatforms(ctx context.Context, firestartrDir *dagger.Directory) ([]Platform, error) {

	platforms := []Platform{}

	for _, ext := range []string{".yaml", ".yml"} {

		filePaths, err := firestartrDir.Glob(ctx, "platforms/*"+ext)

		if err != nil {

			return nil, err

		}

		for _, filePath := range filePaths {

			fileContent, err := firestartrDir.File(filePath).Contents(ctx)

			if err != nil {

				return nil, err

			}

			reg := Platform{}

			err = yaml.Unmarshal([]byte(fileContent), &reg)

			if err != nil {

				return nil, err

			}

			platforms = append(platforms, reg)

		}

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
