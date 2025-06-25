// A generated module for HydrateSecrets functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
	"dagger/hydrate-secrets/internal/dagger"
	"fmt"

	"gopkg.in/yaml.v3"
)

var FIRESTARTR_DOCKER_IMAGE = "ghcr.io/prefapp/gitops-k8s:v1.43.1_slim"

type HydrateSecrets struct {
	ValuesDir        *dagger.Directory
	WetRepoDir       *dagger.Directory
	DotFirestartrDir *dagger.Directory
	Config           Config
}

func New(
	ctx context.Context,

	valuesDir *dagger.Directory,

	wetRepoDir *dagger.Directory,

	dotFirestartrDir *dagger.Directory,
) *HydrateSecrets {

	configFileExists, err := hydrateConfigFileExists(ctx, valuesDir)
	if err != nil {
		panic(fmt.Errorf("failed to check for hydrate_tfworkspaces_config.yaml: %w", err))
	}

	config := Config{
		Image: FIRESTARTR_DOCKER_IMAGE,
	}

	if configFileExists {

		loadedConfigFromFile := Config{}
		configContents, err := valuesDir.
			File(".github/hydrate_tfworkspaces_config.yaml").
			Contents(ctx)

		if err != nil {
			panic(err)
		}

		err = yaml.Unmarshal([]byte(configContents), &loadedConfigFromFile)
		if err != nil {
			panic(err)
		}

		if loadedConfigFromFile.Image != "" {
			config.Image = loadedConfigFromFile.Image
		}
	}

	return &HydrateSecrets{

		ValuesDir: valuesDir,

		WetRepoDir: wetRepoDir,

		DotFirestartrDir: dotFirestartrDir,

		Config: config,
	}
}

func (m *HydrateSecrets) Render(ctx context.Context, app string, tenant string, env string) ([]*dagger.Directory, error) {

	claimName := fmt.Sprintf(`%s-%s-%s`, app, tenant, env)

	targetEntry := fmt.Sprintf("secrets/%s/%s/", tenant, env)
	_, err := m.ValuesDir.Entries(ctx, dagger.DirectoryEntriesOpts{
		Path: targetEntry,
	})
	if err != nil {
		return nil, fmt.Errorf(
			"claim not found in tenant: %s env: %s",
			tenant,
			env,
		)
	}

	secretsDir, err := m.InferSecretsClaimData(
		ctx,
		app,
		m.ValuesDir.Directory("secrets"),
	)
	if err != nil {
		return nil, err
	}

	outputDir, err := m.RenderWithFirestartrContainer(
		ctx,
		secretsDir,
        claimName
	)
	if err != nil {
		return nil, err
	}

	crFiles, err := m.GetCrsFileByClaimName(
		ctx,
		claimName,
		outputDir,
	)
	if err != nil {
		return nil, err
	}

	for _, crFile := range crFiles {
		crFileName, err := crFile.Name(ctx)
		if err != nil {
			return nil, err
		}

		m.WetRepoDir = m.WetRepoDir.
			WithoutFile(fmt.Sprintf("secrets/%s", crFileName)).
			WithFile(fmt.Sprintf("secrets/%s", crFileName), crFile)
	}

	return []*dagger.Directory{m.WetRepoDir}, nil

}
