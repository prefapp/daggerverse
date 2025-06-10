package main

import (
	"context"
	"dagger/hydrate-tfworkspaces/internal/dagger"
	"fmt"
	"strings"
)

func (m *HydrateTfworkspaces) InferSecretsClaimData(
	ctx context.Context,
	app string,
	secretsDir *dagger.Directory,
) (*dagger.Directory, error) {

	entries, err := secretsDir.Glob(ctx, "*/*/*.yaml")
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		splitted := strings.Split(entry, "/")
		tenant := splitted[0]
		env := splitted[1]
		claimData, err := secretsDir.File(entry).Contents(ctx)
		if err != nil {
			return nil, err
		}

		patchedData, err := m.PatchSecretClaimData(app, tenant, env, claimData)
		if err != nil {
			return nil, err
		}

		secretsDir = secretsDir.
			WithoutFile(entry).
			WithNewFile(entry, patchedData)
	}

	return secretsDir, nil
}

func (m *HydrateTfworkspaces) PatchSecretClaimData(
	app string,
	tenant string,
	env string,
	claimData string,
) (string, error) {
	name := fmt.Sprintf(`%s-%s-%s`, app, tenant, env)

	pathsValueMap := map[string]string{
		"/name":                            fmt.Sprintf(`"%s"`, name),
		"/providers/external_secrets/name": fmt.Sprintf(`"%s"`, name),
		"/providers/external_secrets/secretStore": fmt.Sprintf(`{"name": "%s", "kind": "SecretStore"}`, name),
	}

	for path, value := range pathsValueMap {
		var err error
		claimData, err = m.PatchClaim(path, value, claimData)
		if err != nil {
			return "", err
		}
	}

	return claimData, nil
}
