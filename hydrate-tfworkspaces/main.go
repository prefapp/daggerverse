// A generated module for HydrateTfworkspaces functions
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
	"dagger/hydrate-tfworkspaces/internal/dagger"
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
}

func New(
	// The path to the values directory, where the helm values are stored
	valuesDir *dagger.Directory,

	// The path to the wet repo directory, where the wet manifests are stored
	wetRepoDir *dagger.Directory,

	dotFirestartrDir *dagger.Directory,

) *HydrateTfworkspaces {
	return &HydrateTfworkspaces{

		ValuesDir: valuesDir,

		WetRepoDir: wetRepoDir,

		DotFirestartrDir: dotFirestartrDir,
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

	platformClaimsPath := "platform-claims/claims/tfworkspaces"

	coordinatesPath := "app-claims/tfworkspaces"

	platformClaimsDir := m.ValuesDir.Directory(platformClaimsPath)

	appClaimsDir := m.ValuesDir.Directory(coordinatesPath)

	claimNames, err := m.GetAppClaimNames(ctx)

	if err != nil {

		return nil, err

	}

	crsWithMicroserviceAnnotation, err := m.GetCrsWithMicroserviceAnnotation(ctx)

	if err != nil {

		return nil, err

	}

	appClaimsDir, err = m.PatchClaimsWithPreviousImages(
		ctx,
		crsWithMicroserviceAnnotation,
		appClaimsDir,
	)

	if err != nil {

		return nil, err

	}

	appClaimsDir, err = m.PatchClaimsWithNewImageValues(
		ctx,
		newImagesMatrix,
		appClaimsDir,
	)

	if err != nil {

		return nil, err

	}

	combDirs := dag.Directory().
		WithDirectory("platform", platformClaimsDir).
		WithDirectory("app", appClaimsDir)

	platformFound := dag.
		FirestartrConfig(m.DotFirestartrDir).
		FindPlatformByName(platform)

	if platformFound == nil {

		return nil, fmt.Errorf("platform %s not found", platform)

	}

	entries, err := appClaimsDir.Glob(ctx, "**.yaml")

	if err != nil {

		return nil, err

	}

	if len(entries) == 0 {

		return nil, fmt.Errorf("no claims found in %s", platformClaimsPath)

	}

	cmd, err := m.CmdContainer().
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

	outputDir := cmd.Directory("/output")

	entries, err = outputDir.Glob(ctx, "**.yaml")

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

	return []*dagger.Directory{outputDir}, nil

}

func (m *HydrateTfworkspaces) CmdContainer() *dagger.Container {

	return dag.Container().
		From("ghcr.io/prefapp/gitops-k8s:v1.26.2_slim")

}

func (m *HydrateTfworkspaces) GetAppClaimNames(

	ctx context.Context,

) ([]string, error) {

	coordinatesPath := "app-claims/tfworkspaces"

	appClaimsDir := m.ValuesDir.Directory(coordinatesPath)

	claimNamesFromAppDir := []string{}

	entries, err := appClaimsDir.Glob(ctx, "**.yaml")

	if err != nil {

		return nil, err

	}

	for _, entry := range entries {

		claim := Claim{}

		fileContent, err := appClaimsDir.File(entry).Contents(ctx)

		if err != nil {

			return nil, err

		}

		err = yaml.Unmarshal([]byte(fileContent), &claim)

		if err != nil {

			return nil, err

		}

		claimNamesFromAppDir = append(claimNamesFromAppDir, claim.Name)

	}

	return claimNamesFromAppDir, nil

}

func (m *HydrateTfworkspaces) GetCrsWithMicroserviceAnnotation(ctx context.Context) ([]Cr, error) {

	entries, err := m.WetRepoDir.Glob(ctx, "**.yaml")

	if err != nil {

		return nil, err

	}

	crs := []Cr{}

	for _, entry := range entries {

		fileContent, err := m.WetRepoDir.File(entry).Contents(ctx)

		if err != nil {

			return nil, err

		}

		cr := Cr{}

		err = yaml.Unmarshal([]byte(fileContent), &cr)

		if err != nil {

			return nil, err

		}

		if cr.Metadata.Annotations.MicroService != "" {

			crs = append(crs, cr)

		}

	}

	return crs, nil
}
