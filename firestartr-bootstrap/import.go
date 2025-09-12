package main

import (
	"context"
	"dagger/firestartr-bootstrap/internal/dagger"
	"fmt"
	"time"
)

func (m *FirestartrBootstrap) RunImporter(
	ctx context.Context,
	kindContainer *dagger.Container,
) *dagger.Container {
	claimsDir := dag.Directory().
		WithNewDirectory("/claims")

	crsDir := dag.Directory().
		WithNewDirectory("/crs")

	kindContainer = kindContainer.
		WithDirectory("/import", claimsDir).
		WithDirectory("/import", crsDir).
		WithDirectory("/config", m.CrsDotConfigDir).
		WithDirectory("/claims_defaults", m.ClaimsDotConfigDir).
		WithEnvVariable("BUST_CACHE", time.Now().String()).
		WithEnvVariable("DEBUG", "*").
		WithEnvVariable("GITHUB_APP_ID", m.Creds.GithubApp.GhAppId).
		WithEnvVariable("GITHUB_APP_INSTALLATION_ID", m.Creds.GithubApp.InstallationId).
		WithEnvVariable("GITHUB_APP_PEM_FILE", m.Creds.GithubApp.Pem).
		WithEnvVariable("PREFAPP_BOT_PAT", m.Creds.GithubApp.BotPat).
		WithEnvVariable("ORG", m.GhOrg).
		WithExec([]string{"apk", "add", "nodejs", "npm"}).
		WithExec([]string{
			"npm",
			"install",
			"-g",
			fmt.Sprintf("@firestartr/cli@v%s", m.Bootstrap.Firestartr.Version),
		}).
		WithExec(
			[]string{
				"firestartr-cli", "importer",
				"--org", m.GhOrg,
				"--config", "/config",
				"--crs", "/import/crs",
				"--claims", "/import/claims",
				"--claimsDefaults", "/claims_defaults",
				"--filters", "gh-repo,SKIP=SKIP",
				"--filters", "gh-group,REGEXP=[A-Za-z0-9\\-]+",
				"--filters", "gh-members,REGEXP=[A-Za-z0-9\\-]+",
			},
		)

	kindContainer = m.ApplyFirestartrCrs(ctx, kindContainer, "/import/crs")

	return kindContainer

}
