package main

import (
	"context"
	"dagger/firestartr-config/internal/dagger"

	"gopkg.in/yaml.v3"
)

type Registry struct {
	Name         string       `yaml:"name"`
	Url          string       `yaml:"url"`
	ImageTypes   []string     `yaml:"image_types,omitempty"`
	AuthStrategy AuthStrategy `yaml:"auth_strategy,omitempty"`
	BasePaths    BasePaths    `yaml:"base_paths,omitempty"`
}

func (r *Registry) isValid() bool {

	if r.Name == "" || r.Url == "" {
		return false
	}

	if r.AuthStrategy != "" {
		switch r.AuthStrategy {
		case AuthStrategyAWSOIDC, AuthStrategyAzureOIDC, AuthStrategyGeneric, AuthStrategyGHCR, AuthStrategyDockerHub:
			return true
		default:
			return false
		}
	}

	return true

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

			/*
			This is done this way because Unmarshal does not have a way of validating the content while unmarshalling.
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
