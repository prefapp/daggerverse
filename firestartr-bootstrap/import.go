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

	initialCrsTemplate, err := m.RenderInitialCrs(ctx,
		dag.CurrentModule().
			Source().
			File("templates/initial_crs.tmpl"),
	)
	if err != nil {
		panic(err)
	}

	initialCrsDir, err := m.SplitRenderedCrsInFiles(initialCrsTemplate)
	if err != nil {
		panic(err)
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
		WithEnvVariable("PREFAPP_BOT_PAT", m.Creds.GithubApp.BotPat).
		WithEnvVariable("ORG", m.GhOrg).
		WithExec([]string{"apk", "add", "nodejs", "npm"}).
		WithExec([]string{
			"npm", "install", "-g",
			fmt.Sprintf("@firestartr/cli@v%s", m.Bootstrap.Firestartr.Version),
		}).
		WithDirectory("/resources/initial-crs", initialCrsDir).
		WithExec([]string{
			"kubectl",
			"apply",
			"-f", "/resources/initial-crs",
		}).
		WithExec(importCommand)

	kindContainer = m.ApplyFirestartrCrs(
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

	return kindContainer

}
