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

	coordinatesPath := fmt.Sprintf("app-claims/tfworkspaces/%s/%s/%s", platform, tenant, env)

	platformClaimsDir := m.ValuesDir.Directory(platformClaimsPath)

	appClaimsDir := m.ValuesDir.Directory(coordinatesPath)

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

	fmt.Printf("entries: %v\n", entries)

	if err != nil {

		return nil, err

	}

	if len(entries) == 0 {

		return nil, fmt.Errorf("no claims found in %s", platformClaimsPath)

	}

	cmd := m.CmdContainer().
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
		)

	return []*dagger.Directory{cmd.Directory("/output")}, nil

}

func (m *HydrateTfworkspaces) CmdContainer() *dagger.Container {

	return dag.Container().
		From("ghcr.io/prefapp/gitops-k8s:v1.26.2_slim")

}
