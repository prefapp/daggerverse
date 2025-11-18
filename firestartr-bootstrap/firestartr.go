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
		fmt.Sprintf(
			"ghcr.io/prefapp/gitops-k8s:%s_slim",
			m.Bootstrap.Firestartr.OperatorVersion,
		),
	).
		WithDirectory("/claims", claimsDir).
		WithDirectory("/crs", crsDir).
		WithDirectory("/config", m.CrsDotConfigDir).
		WithDirectory("/claims_defaults", m.ClaimsDotConfigDir).
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
				"--outputCrDir", "/tmp/rendered_crs/infra",
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
				"--outputCrDir", "/tmp/rendered_crs/github",
				"--provider", "github",
			},
		).Sync(ctx)

	if err != nil {

		return nil, err

	}

	outputDir := fsCtr.Directory("/tmp/rendered_crs")

	return outputDir, nil
}
