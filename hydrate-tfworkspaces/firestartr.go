package main

import (
	"context"
	"dagger/hydrate-tfworkspaces/internal/dagger"
	"path"
)

func (m *HydrateTfworkspaces) RenderWithFirestartrContainer(ctx context.Context, claimsDir *dagger.Directory) (*dagger.Directory, error) {

	fsCtr, err := dag.Container().
		From(m.Config.Image).
		WithMountedDirectory("claims", claimsDir).
		WithMountedDirectory("/crs", m.WetRepoDir).
		WithDirectory("/.config", m.ValuesDir.Directory(".config")).
		WithDirectory("/crs/.config", dag.Directory()).
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
		).Sync(ctx)

	if err != nil {

		return nil, err

	}

	outputDir := fsCtr.Directory("/output")

	return outputDir, nil
}
