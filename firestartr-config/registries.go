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
	Default      bool      `yaml:"default"`
	AuthStrategy string    `yaml:"auth_strategy"`
	BasePaths    BasePaths `yaml:"base_paths"`
}

type BasePaths struct {
	Services string `yaml:"services"`
	Charts   string `yaml:"charts"`
}

func (m *FirestartrConfig) GetRegistries(ctx context.Context, firestartrDir *dagger.Directory) []Registry {

	registries := []Registry{}

	for _, ext := range []string{".yaml", ".yml"} {

		filePaths, err := firestartrDir.Glob(ctx, "docker_registries/*"+ext)

		if err != nil {

			panic(err)

		}

		for _, filePath := range filePaths {

			fileContent, err := firestartrDir.File(filePath).Contents(ctx)

			if err != nil {

				panic(err)

			}

			reg := Registry{}

			err = yaml.Unmarshal([]byte(fileContent), &reg)

			if err != nil {

				panic(err)

			}

			registries = append(registries, reg)

		}

	}

	return registries

}
