package main

import (
	"context"
	"dagger/firestartr-config/internal/dagger"

	"gopkg.in/yaml.v3"
)

type AuthStrategy string

const (
	AuthStrategyAWSOIDC    AuthStrategy = "aws_oidc"
	AuthStrategyAzureOIDC  AuthStrategy = "azure_oidc"
	AuthStrategyGeneric    AuthStrategy = "generic"
	AuthStrategyGHCR       AuthStrategy = "ghcr"
	AuthStrategyDockerHub  AuthStrategy = "dockerhub"
)

type LegacyRegistry struct {
	Name         string          `yaml:"name"`
	Registry     string          `yaml:"registry"`
	ImageTypes   []string        `yaml:"image_types"`
	AuthStrategy *AuthStrategy   `yaml:"auth_strategy,omitempty"`
	BasePaths    LegacyBasePaths `yaml:"base_paths"`
}

type LegacyBasePaths struct {
	Services string `yaml:"services"`
	Charts   string `yaml:"charts"`
}

func loadLegacyRegistries(ctx context.Context, firestartrDir *dagger.Directory) ([]LegacyRegistry, error) {

	registries := []LegacyRegistry{}

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

			reg := LegacyRegistry{}

			err = yaml.Unmarshal([]byte(fileContent), &reg)

			registries = append(registries, reg)

		}

	}

	return registries, nil

}
