package main

import (
	"context"
	"dagger/hydrate-tfworkspaces/internal/dagger"
	"encoding/json"
	"fmt"
	"path"
	"slices"
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

	configContents, err := valuesDir.File("app-claims/.github/hydrate_tfworkspaces_config.yaml").Contents(ctx)

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

	env string,

	platform string,

	tenant string,

	// +optional
	// +default="{\"images\":[]}"
	newImagesMatrix string,

) ([]*dagger.Directory, error) {

	matrix := ImageMatrix{}

	err := json.Unmarshal([]byte(newImagesMatrix), &matrix)

	if err != nil {

		return nil, err

	}

	platformClaimsDir := m.ValuesDir.Directory("platform-claims/claims/tfworkspaces")

	appClaimsDir := m.ValuesDir.Directory("app-claims/tfworkspaces")

	crsWithPreviousImages, err := m.GetPreviousImagesFromCrs(ctx, matrix)

	if err != nil {

		return nil, err

	}

	appClaimsDir, err = m.PatchClaimsWithPreviousImages(
		ctx,
		crsWithPreviousImages,
		appClaimsDir,
	)

	if err != nil {

		return nil, err

	}

	appClaimsDir, err = m.PatchClaimsWithNewImageValues(
		ctx,
		matrix,
		appClaimsDir,
	)

	if err != nil {

		return nil, err

	}

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

	fsCtr, err := dag.Container().
		From(m.Config.Image).
		WithMountedDirectory("claims", combDirs).
		WithMountedDirectory("/crs", m.WetRepoDir).
		WithDirectory("/.config", m.ValuesDir.Directory("platform-claims/.config")).
		WithEnvVariable("DEBUG", "firestartr:*").
		WithExec(
			[]string{
				"./run.sh",
				"cdk8s",
				"--render",
				"--disableRenames",
				"--globals", path.Join("/crs", ".config"),
				"--initializers", path.Join("/crs", ".config"),
				"--claims", "claims",
				"--previousCRs", "/crs",
				"--excludePath", path.Join("/crs", ".github"),
				"--claimsDefaults", "/.config",
				"--outputCrDir", "/output",
				"--provider", "terraform",
			},
		).
		Sync(ctx)

	if err != nil {

		return nil, err

	}

	outputDir := fsCtr.Directory("/output")

	entries, err = outputDir.Glob(ctx, "**.yaml")

	if err != nil {

		return nil, err

	}

	claimNames, err := m.GetAppClaimNames(ctx)

	if err != nil {

		return nil, err

	}

	for _, entry := range entries {

		fileContent, err := outputDir.File(entry).Contents(ctx)

		if err != nil {

			return nil, err

		}

		cr := Cr{}

		err = yaml.Unmarshal([]byte(fileContent), &cr)

		claimName := strings.Split(cr.Metadata.Annotations.ClaimRef, "/")[1]

		if err != nil {

			return nil, err

		}

		if !slices.Contains(claimNames, claimName) {

			outputDir = outputDir.WithoutFile(entry)

		}
	}

	for _, cr := range crsWithPreviousImages {

		outputDir, err = m.AddAnnotationsToCr(
			ctx,
			strings.Split(cr.Metadata.Annotations.ClaimRef, "/")[1],
			cr.Metadata.Annotations.Image,
			cr.Metadata.Annotations.MicroService,
			outputDir,
		)

		if err != nil {

			return nil, err

		}

	}

	if len(matrix.Images) == 1 {

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

	entries, err = outputDir.Glob(ctx, "**.yaml")

	if err != nil {

		return nil, err

	}

	for _, entry := range entries {

		file := outputDir.File(entry)

		fileContent, err := file.Contents(ctx)

		if err != nil {

			return nil, err

		}

		cr := Cr{}

		err = yaml.Unmarshal([]byte(fileContent), &cr)

		if err != nil {

			return nil, err

		}

		if strings.Split(cr.Metadata.Annotations.ClaimRef, "/")[1] == matrix.Images[0].Platform {

			m.WetRepoDir = m.WetRepoDir.
				WithoutFile(entry).
				WithFile(entry, file)

			break
		}

	}

	return []*dagger.Directory{m.WetRepoDir}, nil

}
