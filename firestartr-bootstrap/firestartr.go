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

	claimsDefaults, err := m.RenderClaimsDefaults(ctx,
		dag.CurrentModule().
			Source().
			File("firestartr_files/claims/.config/claims_defaults.tmpl"),
	)
	if err != nil {
		return nil, err
	}

	wetReposConfig, err := m.RenderWetReposConfig(ctx,
		dag.CurrentModule().
			Source().
			File("firestartr_files/claims/.config/wet-repositories-config.tmpl"),
	)
	if err != nil {
		return nil, err
	}

	defaultsDir := dag.Directory().
		WithNewDirectory("/claims_defaults").
		WithNewFile("claims_defaults.yaml", claimsDefaults).
		WithNewFile("wet-repositories-config.yaml", wetReposConfig)

	fsCtr, err := dag.Container().From(
		fmt.Sprintf(
			"ghcr.io/prefapp/gitops-k8s:%s_slim",
			m.Bootstrap.Firestartr.Version,
		),
	).
		WithDirectory("/claims", claimsDir).
		WithDirectory("/crs", crsDir).
		WithDirectory("/config", dag.CurrentModule().Source().Directory("firestartr_files/crs/.config")).
		WithDirectory("/claims_defaults", defaultsDir).
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
				"--outputCrDir", "/output",
				"--provider", "github",
			},
		).Sync(ctx)

	if err != nil {

		return nil, err

	}

	outputDir := fsCtr.Directory("/output")

	return outputDir, nil
}
