package main

import (
	"context"
	"dagger/firestartr-config/internal/dagger"

	"gopkg.in/yaml.v3"
)

type Registry struct {
	Name         string    `yaml:"name"`
	Registry     string    `yaml:"registry"`
	ImageTypes   []string  `yaml:"image_types"`
	IsDefault    bool      `yaml:"default"`
	AuthStrategy string    `yaml:"auth_strategy"`
	BasePaths    BasePaths `yaml:"base_paths"`
}

type BasePaths struct {
	Services string `yaml:"services"`
	Charts   string `yaml:"charts"`
}

func loadRegistries(ctx context.Context, firestartrDir *dagger.Directory) ([]Registry, error) {

	registries := []Registry{}

	for _, ext := range []string{".yaml", ".yml"} {

		filePaths, err := firestartrDir.Glob(ctx, "docker_registries/*"+ext)

		if err != nil {

			return nil, err

		}

		for _, filePath := range filePaths {

			fileContent, err := firestartrDir.File(filePath).Contents(ctx)

			if err != nil {

				return nil, err

			}

			reg := Registry{}

			err = yaml.Unmarshal([]byte(fileContent), &reg)

			if err != nil {

				return nil, err

			}

			registries = append(registries, reg)

		}

	}

	return registries, nil

}
