package main

import (
	"context"
	"dagger/hydrate-secrets/internal/dagger"
	"path"
)

func (m *HydrateSecrets) RenderWithFirestartrContainer(ctx context.Context, claimsDir *dagger.Directory, claimName string) (*dagger.Directory, error) {

	fsCtr, err := dag.Container().
		From(m.Config.Image).
		WithMountedDirectory("claims", claimsDir).
		WithMountedDirectory("/crs", m.WetRepoDir).
		WithDirectory("/.config", m.ValuesDir.Directory(".config")).
		WithDirectory("/crs/.config", dag.Directory()).
		WithNewFile("/crs/secrets/.gitkeep", "").
		WithEnvVariable("DEBUG", "NONE").
		WithExec(
			[]string{
				"./run.sh",
				"cdk8s",
				"--render",
				"--disableRenames",
				"--globals", path.Join("/crs", ".config"),
				"--initializers", path.Join("/crs", ".config"),
				"--claims", "claims",
				"--previousCRs", "/crs/secrets",
				"--excludePath", path.Join("/crs", ".github"),
				"--claimsDefaults", "/.config",
				"--outputCrDir", "/output",
				"--claimRefsList", "SecretsClaim-" + claimName,
				"--provider", "externalSecrets",
			},
		).Sync(ctx)

	if err != nil {

		return nil, err

	}

	outputDir := fsCtr.Directory("/output")

	return outputDir, nil
}
