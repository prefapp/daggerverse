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
	alreadyCreatedReposList []string,
) *dagger.Container {
	claimsDir := dag.Directory().
		WithNewDirectory("/claims")

	renderedClaims, err := m.RenderBootstrapFile(
		ctx,
		dag.CurrentModule().
			Source().
			File("templates/initial_claims.tmpl"),
	)
	if err != nil {
		panic(err)
	}

	claimsDir, err = m.SplitRenderedClaimsInFiles(renderedClaims)
	if err != nil {
		panic(err)
	}

	crsDir := dag.Directory().
		WithNewDirectory("/crs")

	importCommand := []string{
		"firestartr-cli", "importer",
		"--org", m.GhOrg,
		"--config", "/config",
		"--crs", "/import/crs",
		"--claims", "/import/claims",
		"--claimsDefaults", "/claims_defaults",
		"--filters", "gh-group,REGEXP=[A-Za-z0-9\\-]+",
		"--filters", "gh-members,REGEXP=[A-Za-z0-9\\-]+",
	}
	if len(alreadyCreatedReposList) > 0 {
		for _, repoName := range alreadyCreatedReposList {
			importCommand = append(
				importCommand,
				"--filters",
				fmt.Sprintf("gh-repo,NAME=%s", repoName),
			)
		}
	} else {
		importCommand = append(
			importCommand,
			"--filters",
			"gh-repo,SKIP=SKIP",
		)
	}

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
			"npm", "install", "-g",
			fmt.Sprintf("@firestartr/cli@v%s", "1.50.1-snapshot-32"),
			// fmt.Sprintf("@firestartr/cli@v%s", m.Bootstrap.Firestartr.Version),
		}).
		WithExec(importCommand)

	kindContainer = m.ApplyFirestartrCrs(ctx, kindContainer, "/import/crs")

	return kindContainer

}
