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

func (lc *LegacyRegistry) isValid() bool {

	if lc.Name == "" || lc.Registry == "" || len(lc.ImageTypes) == 0 {
		return false
	}

	if lc.AuthStrategy != nil {
		switch *lc.AuthStrategy {
		case AuthStrategyAWSOIDC, AuthStrategyAzureOIDC, AuthStrategyGeneric, AuthStrategyGHCR, AuthStrategyDockerHub:
			return true
		default:
			return false
		}
	}

	return true

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
			if err != nil {
				return nil, err
			}

			/*
			This is done this way because Unmarshal does not has a way of validating the content while unmarshalling.
			We need to make sure that it has all required fields since it can get confused with legacy configuration
			and create a valid object with missing fields, specifically the url and registry.
			We cannot use a custom UnmarshalYAML function because dagger then is not able to export that type in dagger.gen
			*/
			if reg.isValid() {
				registries = append(registries, reg)
			}

		}

	}

	return registries, nil

}
