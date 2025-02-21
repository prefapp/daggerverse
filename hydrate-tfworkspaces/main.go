package main

import (
	"context"
	"dagger/hydrate-tfworkspaces/internal/dagger"
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

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

	configContents, err := valuesDir.File(".github/hydrate_tfworkspaces_config.yaml").Contents(ctx)

	if err != nil {

		panic(err)

	}

	config := Config{}

	err = yaml.Unmarshal([]byte(configContents), &config)

	if err != nil {

		panic(err)

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
	platformClaimsDir := m.ValuesDir.Directory("claims/tfworkspaces")

	appClaimsDir := m.ValuesDir.Directory("tfworkspaces")

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

	appClaimsDir, err = m.PatchClaimWithNewImageValues(ctx, matrix, appClaimsDir)

	if err != nil {

		return nil, err

	}

	// Combine platform and app claims directories
	combDirs := dag.Directory().
		WithDirectory("platform", platformClaimsDir).
		WithDirectory("app", appClaimsDir)

	entries, err := appClaimsDir.Glob(ctx, "**.yaml")

	if err != nil {

		return nil, err

	}

	if len(entries) == 0 {

		return nil, fmt.Errorf("no claims found in %s", "platform-claims/claims/tfworkspaces")

	}

	outputDir, err := m.RenderWithFirestartrContainer(ctx, combDirs)

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
			matrix.Images[0].Platform,
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
		WithoutFile(crFileName).
		WithFile(crFileName, crFile)

	return []*dagger.Directory{m.WetRepoDir}, nil

}
