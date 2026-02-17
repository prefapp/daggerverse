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
) (*dagger.Container, error) {
	claimsDir := dag.Directory().
		WithNewDirectory("/claims")

	renderedClaims, err := m.RenderBootstrapFile(
		ctx,
		dag.CurrentModule().
			Source().
			File("templates/initial_claims.tmpl"),
	)
	if err != nil {
		return nil, err
	}

	claimsDir, err = m.SplitRenderedClaimsInFiles(renderedClaims)
	if err != nil {
		return nil, err
	}

	crsDir := dag.Directory().
		WithNewDirectory("/crs")

	groupFilter := "gh-group,REGEXP=[A-Za-z0-9\\-]+"

	if m.IncludeAllGroup {
		// If the group has been included (that is, created locally by the
		// bootstrap process), we want to exclude it from the import,
		// to avoid duplications.
		groupFilter = fmt.Sprintf("gh-group,REGEXP=^(?!%s-all)[A-Za-z0-9\\-]+$", m.GhOrg)
	}

	importCommand := []string{
		"firestartr-cli", "importer",
		"--org", m.GhOrg,
		"--config", "/config",
		"--crs", "/import/crs",
		"--claims", "/import/claims",
		"--claimsDefaults", "/claims_defaults",
		"--filters", groupFilter,
		"--filters", "gh-members,REGEXP=[A-Za-z0-9\\-]+",
		"--filters", "gh-repo,SKIP=SKIP",
	}

	kindContainer = kindContainer.
		WithDirectory("/import", claimsDir).
		WithDirectory("/import", crsDir).
		WithDirectory("/config", m.CrsDotConfigDir).
		WithDirectory("/claims_defaults", m.ClaimsDotConfigDir).
		WithEnvVariable("BUST_CACHE", time.Now().String()).
		WithEnvVariable("LOG_LEVEL", "debug").
		WithEnvVariable("GITHUB_APP_ID", m.Creds.GithubApp.GhAppId).
		WithEnvVariable("GITHUB_APP_INSTALLATION_ID", m.Creds.GithubApp.InstallationId).
		WithEnvVariable("GITHUB_APP_PEM_FILE", m.Creds.GithubApp.Pem).
		WithEnvVariable("PREFAPP_BOT_PAT", m.Creds.GithubApp.PrefappBotPat).
		WithEnvVariable("ORG", m.GhOrg).
		WithExec([]string{"apk", "add", "nodejs", "npm"}).
		WithExec([]string{
			"npm", "install", "-g",
			fmt.Sprintf("@firestartr/cli@%s", m.Bootstrap.Firestartr.CliVersion),
		}).
		WithExec(importCommand)

    // for debugging purposes
    kindContainer = kindContainer.
        WithExec([]string{"rm", "-rf", "/debug"}).
        WithExec([]string{"mkdir", "-p","/debug/import"}).
        WithExec([]string{"cp","-a", "/import", "/debug/import"})

	kindContainer, err = m.ApplyFirestartrCrs(
		ctx,
		kindContainer,
		"/import/crs",
		[]string{
			"FirestartrGithubMembership.*",
			"FirestartrGithubGroup.*",
			"FirestartrGithubRepository.*",
			"FirestartrGithubRepositoryFeature.*",
		},
	)
	if err != nil {
		return nil, err
	}

	return kindContainer, nil

}
