package main

import (
	"context"
	"dagger/firestartr-config/internal/dagger"

	"gopkg.in/yaml.v3"
)

type Repository struct {
	Name string `yaml:"name"`
	Url  string `yaml:"url"`
}

func loadRepositories(ctx context.Context, firestartrDir *dagger.Directory) ([]Repository, error) {

	repositories := []Repository{}

	for _, ext := range []string{".yaml", ".yml"} {

		filePaths, err := firestartrDir.Glob(ctx, "helm_repositories/*"+ext)

		if err != nil {

			return nil, err

		}

		for _, filePath := range filePaths {

			fileContent, err := firestartrDir.File(filePath).Contents(ctx)

			if err != nil {

				return nil, err

			}

			reg := Repository{}

			err = yaml.Unmarshal([]byte(fileContent), &reg)

			if err != nil {

				return nil, err

			}

			repositories = append(repositories, reg)

		}

	}

	return repositories, nil

}
