package main

import (
	"context"
	"dagger/firestartr-bootstrap/internal/dagger"
	"fmt"
	"time"
)

func (m *FirestartrBootstrap) RenderWithFirestartrContainer(
	ctx context.Context,
	claimsDir *dagger.Directory,
	crsDir *dagger.Directory,
) (*dagger.Directory, error) {

	entries, err := claimsDir.Glob(ctx, "**")
	if err != nil {
		return nil, err
	}
	fmt.Printf("💡 💡 Claims Entries: %v\n", entries)

	fsCtr, err := dag.Container().From(
		// fmt.Sprintf(
		// 	"ghcr.io/prefapp/gitops-k8s:v%s_slim",
		// 	m.Bootstrap.Firestartr.Version,
		// ),
		"ghcr.io/prefapp/gitops-k8s:18ce79a_full-aws",
	).
		WithDirectory("/claims", claimsDir).
		WithDirectory("/crs", crsDir).
		WithDirectory("/config", dag.CurrentModule().Source().Directory("firestartr_files/crs/.config")).
		WithDirectory("/claims_defaults", m.DotConfigDir).
		WithEnvVariable("BUST_CACHE", time.Now().String()).
		WithEnvVariable("DEBUG", "*").
		WithEnvVariable("GITHUB_APP_ID", m.Creds.GithubApp.GhAppId).
		WithEnvVariable("GITHUB_APP_INSTALLATION_ID", m.Creds.GithubApp.InstallationId).
		WithEnvVariable("GITHUB_APP_PEM_FILE", m.Creds.GithubApp.Pem).
		WithEnvVariable("PREFAPP_BOT_PAT", m.Creds.GithubApp.BotPat).
		WithEnvVariable("ORG", m.GhOrg).
		WithExec(
			[]string{
				"./run.sh",
				"cdk8s",
				"--render",
				"--disableRenames",
				"--globals", "/config",
				"--initializers", "/config",
				"--claims", "/claims",
				"--previousCRs", "/crs",
				"--excludePath", "/.config",
				"--claimsDefaults", "/claims_defaults",
				"--outputCrDir", "/tmp/rendered_crs",
				"--provider", "terraform",
			},
		).
		WithExec(
			[]string{
				"./run.sh",
				"cdk8s",
				"--render",
				"--disableRenames",
				"--globals", "/config",
				"--initializers", "/config",
				"--claims", "/claims",
				"--previousCRs", "/crs",
				"--excludePath", "/.config",
				"--claimsDefaults", "/claims_defaults",
				"--outputCrDir", "/tmp/rendered_crs",
				"--provider", "github",
			},
		).Sync(ctx)

	if err != nil {

		return nil, err

	}

	outputDir := fsCtr.Directory("/tmp/rendered_crs")

	return outputDir, nil
}
