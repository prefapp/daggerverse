package main

import (
	"context"
	"dagger/hydrate-tfworkspaces/internal/dagger"
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

var FIRESTARTR_DOCKER_IMAGE = "ghcr.io/prefapp/gitops-k8s:v1.43.2_slim"

type HydrateTfworkspaces struct {
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
) *HydrateTfworkspaces {

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

	return &HydrateTfworkspaces{

		ValuesDir: valuesDir,

		WetRepoDir: wetRepoDir,

		DotFirestartrDir: dotFirestartrDir,

		Config: config,
	}
}

func (m *HydrateTfworkspaces) Render(

	ctx context.Context,

	claimName string,

	app string,

	// +optional
	// +default="{\"images\":[]}"
	newImagesMatrix string,

) ([]*dagger.Directory, error) {

	matrix := ImageMatrix{}

	err := json.Unmarshal([]byte(newImagesMatrix), &matrix)

	if err != nil {

		return nil, err

	}

	// Prepare directories
	platformClaimsDir := m.ValuesDir.Directory("claims")

	appClaimsDir := m.ValuesDir.Directory("tfworkspaces")

	err = dag.Opa(app).ValidateClaims(
		ctx,
		appClaimsDir,
		m.DotFirestartrDir.Directory("validations"),
		m.DotFirestartrDir.Directory("validations/policies"),
	)

	if err != nil {

		return nil, err

	}

	// Patch claim with previous images
	previousCr, err := m.GetPreviousCr(ctx, claimName)

	if err != nil {

		return nil, err

	}

	if previousCr != nil {

		fmt.Printf("☢️ Patching claim %s with previous images\n", claimName)

		appClaimsDir, err = m.PatchClaimWithPreviousImages(
			ctx,
			previousCr,
			appClaimsDir,
		)

		if err != nil {

			return nil, err

		}

	} else {

		fmt.Printf("☢️ Skipping patching claim %s with previous images\n", claimName)

	}

	appClaimsDir, err = m.PatchClaimWithNewImageValues(
		ctx,
		matrix,
		appClaimsDir,
	)

	if err != nil {

		return nil, err

	}

	appClaimsDir, err = m.PatchClaimWithInferredProviders(
		ctx,
		claimName,
		appClaimsDir,
	)
	if err != nil {
		return nil, err
	}

	secretsDir, err := m.InferSecretsClaimData(
		ctx,
		app,
		m.ValuesDir.Directory("secrets"),
	)
	if err != nil {
		return nil, err
	}

	// Combine platform and app claims directories
	combDirs := dag.Directory().
		WithDirectory("platform", platformClaimsDir).
		WithDirectory("app", appClaimsDir).
		WithDirectory("app/secrets", secretsDir)

	entries, err := appClaimsDir.Glob(ctx, "**.yaml")

	if err != nil {

		return nil, err

	}

	if len(entries) == 0 {

		return nil, fmt.Errorf("no claims found in %s", "platform-claims/claims/tfworkspaces")

	}

	outputDir, err := m.RenderWithFirestartrContainer(ctx, combDirs, claimName)

	if err != nil {

		return nil, err

	}

	// Add annotations to cr if no new dispatch is present
	if previousCr != nil && len(matrix.Images) == 0 {

		fmt.Printf("☢️ Adding annotations to cr %s from previous images\n", claimName)

		outputDir, err = m.AddAnnotationsToCr(
			ctx,
			strings.Split(previousCr.Metadata.Annotations.ClaimRef, "/")[1],
			previousCr.Metadata.Annotations.Image,
			previousCr.Metadata.Annotations.MicroServicePointer,
			outputDir,
		)

		if err != nil {

			return nil, err

		}
	}

	// Add annotations to cr if new dispatch is present
	if len(matrix.Images) == 1 {

		fmt.Printf("☢️ Adding annotations to cr %s from new image\n", claimName)

		outputDir, err = m.AddAnnotationsToCr(
			ctx,
			matrix.Images[0].Claim,
			matrix.Images[0].Image,
			matrix.Images[0].ImageKeys[0],
			outputDir,
		)

		if err != nil {

			return nil, err

		}

	}

	// Extract the specific cr file that was pointed from the input
	crFile, err := m.GetCrFileByClaimName(ctx, claimName, outputDir)

	if err != nil {

		return nil, err

	}

	crFileName, err := crFile.Name(ctx)

	if err != nil {

		return nil, err

	}

	m.WetRepoDir = m.WetRepoDir.
		WithoutFile(
			fmt.Sprintf("tfworkspaces/%s", crFileName),
		).
		WithFile(
			fmt.Sprintf("tfworkspaces/%s", crFileName),
			crFile,
		)

	return []*dagger.Directory{m.WetRepoDir}, nil

}
