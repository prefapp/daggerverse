package main

import (
	"context"
	"dagger/firestartr-bootstrap/internal/dagger"
	"fmt"

	"gopkg.in/yaml.v3"
)

func (m *FirestartrBootstrap) UpdateSecretStoreRef(
	ctx context.Context,
	kindContainer *dagger.Container,
) *dagger.Container {
	// Claim reference
	secretClaim, err := getYamlDataFromContainerFile(
		ctx,
		kindContainer,
		"/resources/claims/claims/secrets/platform-secrets.yaml",
	)
	if err != nil {
		panic(err)
	}

	secretClaim["providers"].(map[string]interface{})["external_secrets"].(map[string]interface{})["secretStore"].(map[string]interface{})["name"] = m.Bootstrap.FinalSecretStoreName

	kindContainer = saveYamlDataToContainerFile(
		ctx,
		kindContainer,
		"/resources/claims/claims/secrets/platform-secrets.yaml",
		secretClaim,
	)

	// CR reference
	infraDir := kindContainer.Directory("/resources/firestartr-crs/infra")
	secretsCrNameList, err := infraDir.Glob(ctx, "ExternalSecret.*")
	if err != nil {
		panic(err)
	}
	for _, secretName := range secretsCrNameList {
		secretCr, err := getYamlDataFromContainerFile(
			ctx,
			kindContainer,
			fmt.Sprintf("/resources/firestartr-crs/infra/%s", secretName),
		)
		if err != nil {
			panic(err)
		}

		secretCr["spec"].(map[string]interface{})["secretStoreRef"].(map[string]interface{})["name"] = m.Bootstrap.FinalSecretStoreName

		kindContainer = saveYamlDataToContainerFile(
			ctx,
			kindContainer,
			fmt.Sprintf("/resources/firestartr-crs/infra/%s", secretName),
			secretCr,
		)
	}

	return kindContainer
}

func getYamlDataFromContainerFile(
	ctx context.Context,
	kindContainer *dagger.Container,
	path string,
) (map[string]interface{}, error) {
	fileContents, err := kindContainer.
		File(path).
		Contents(ctx)
	if err != nil {
		return nil, err
	}

	var out map[string]interface{}
	err = yaml.Unmarshal([]byte(fileContents), &out)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func saveYamlDataToContainerFile(
	ctx context.Context,
	kindContainer *dagger.Container,
	path string,
	data map[string]interface{},
) *dagger.Container {
	updatedFileContents, err := yaml.Marshal(data)
	if err != nil {
		panic(err)
	}

	return kindContainer.WithNewFile(path, string(updatedFileContents))
}
