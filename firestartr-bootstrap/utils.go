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
) (*dagger.Container, error) {
	// Claim reference
	secretClaim, err := getYamlDataFromContainerFile(
		ctx,
		kindContainer,
		"/resources/claims/claims/secrets/firestartr-secrets.yaml",
	)
	if err != nil {
		return nil, err
	}

	secretClaim["providers"].(map[string]interface{})["external_secrets"].(map[string]interface{})["secretStore"].(map[string]interface{})["name"] = m.Bootstrap.FinalSecretStoreName

	kindContainer, err = saveYamlDataToContainerFile(
		ctx,
		kindContainer,
		"/resources/claims/claims/secrets/firestartr-secrets.yaml",
		secretClaim,
	)
	if err != nil {
		return nil, err
	}

	// CR reference
	infraDir := kindContainer.Directory("/resources/firestartr-crs/infra")
	secretsCrNameList, err := infraDir.Glob(ctx, "ExternalSecret.*")
	if err != nil {
		return nil, err
	}
	for _, secretName := range secretsCrNameList {
		secretCr, err := getYamlDataFromContainerFile(
			ctx,
			kindContainer,
			fmt.Sprintf("/resources/firestartr-crs/infra/%s", secretName),
		)
		if err != nil {
			return nil, err
		}

		secretCr["spec"].(map[string]interface{})["secretStoreRef"].(map[string]interface{})["name"] = m.Bootstrap.FinalSecretStoreName

		kindContainer, err = saveYamlDataToContainerFile(
			ctx,
			kindContainer,
			fmt.Sprintf("/resources/firestartr-crs/infra/%s", secretName),
			secretCr,
		)
		if err != nil {
			return nil, err
		}
	}

	return kindContainer, nil
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
) (*dagger.Container, error) {
	updatedFileContents, err := yaml.Marshal(data)
	if err != nil {
		return nil, err
	}

	return kindContainer.WithNewFile(path, string(updatedFileContents)), nil
}

func executeCurlCommand(
	ctx context.Context,
	base *dagger.Container,
	patSecret *dagger.Secret,
	command string,
) (*dagger.Container, error) {
	// The pattern is:
	// 1. Set the secret as an environment variable (still necessary for access).
	// 2. Use /bin/sh -c and 'printf' to create a header file inside the container.
	// 3. Execute the target command using the header file with 'curl -H @<file>'.

	// The shell script must be a single string for /bin/sh -c
	shellScript := fmt.Sprintf(`
		# Create a temporary file with the Authorization header.
		# This uses printf and the environment variable, which is more reliable than direct expansion
		# in the main WithExec string argument.
		printf "Authorization: Bearer $GITHUB_PAT" > auth_header.txt;

		# Execute the main command, using the header file via curl's @ syntax.
		# The 'command' string includes the API URL and jq filter.
		curl -s -H @auth_header.txt %s;

		# Clean up the temporary file.
		rm auth_header.txt;
	`, command)

	// Execute the full script using /bin/sh -c and mount the secret.
	execContainer := base.WithSecretVariable("GITHUB_PAT", patSecret).
		WithExec([]string{"/bin/sh", "-c", shellScript})

	return execContainer, nil
}
