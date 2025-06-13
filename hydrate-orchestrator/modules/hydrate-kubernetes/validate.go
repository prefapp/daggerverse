package main

import (
	"context"
	"fmt"

	"gopkg.in/yaml.v3"
)

func (m *HydrateKubernetes) ValidateYamlFiles(
	ctx context.Context,
	files []string,
) error {
	for _, file := range files {
		content, err := m.WetRepoDir.File(file).Contents(ctx)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", file, err)
		}
		if err := ValidateYaml([]byte(content)); err != nil {
			return fmt.Errorf("invalid YAML in file %s: %w", file, err)
		}
	}
	return nil
}

func ValidateYaml(c []byte) error {
	var document interface{}
	err := yaml.Unmarshal(c, &document)
	if err != nil {
		return err
	}
	return nil
}
