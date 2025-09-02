package main

import (
	"context"
	"dagger/firestartr-bootstrap/internal/dagger"
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

func (m *FirestartrBootstrap) RunImporter(
	ctx context.Context,
	dockerSocket *dagger.Socket,
	kindSvc *dagger.Service,
	claimsDir *dagger.Directory,
	crsDir *dagger.Directory,
) *dagger.Container {

	claimsDefaults, err := m.RenderClaimsDefaults(ctx,
		dag.CurrentModule().
			Source().
			File("firestartr_files/claims/.config/claims_defaults.tmpl"),
	)
	if err != nil {
		return nil
	}

	defaultsDir := dag.Directory().
		WithNewDirectory("/claims_defaults").
		WithNewFile("claims_defaults.yaml", claimsDefaults)

	kindContainer := GetKind(dockerSocket, kindSvc).
		WithDirectory("/claims", claimsDir).
		WithDirectory("/crs", crsDir).
		WithDirectory("/config", dag.CurrentModule().Source().Directory("firestartr_files/crs/.config")).
		WithDirectory("/claims_defaults", defaultsDir).
		WithExec([]string{"apk", "add", "nodejs", "npm"}).
		WithExec([]string{"npm", "install", "-g", fmt.Sprintf("@firestartr/cli@%s", m.Bootstrap.Firestartr.Version)}).
		WithExec(
			[]string{
				"firestartr-cli", "importer",
				"--claims", "/claims",
				"--claimsDefaults", "/claims_defaults",
				"--config", "/config",
				"--org", m.GhOrg,
				"--crs", "/crs",
				"--filters", "gh-members,REGEXP=[A-Za-z0-9\\-]+",
				"--filters", "gh-group,REGEXP=[A-Za-z0-9\\-]+",
				"--filters", "gh-repo,SKIP=SKIP",
			},
		)

	kindContainer = m.ApplyFirestartrCrs(ctx, kindContainer)

	return kindContainer

}

func (m *FirestartrBootstrap) ApplyFirestartrCrs(ctx context.Context, kindContainer *dagger.Container) *dagger.Container {

	for _, kind := range []string{
		"FirestartrGithubGroup.*",
		"FirestartrGithubRepository.*",
		"FirestartrGithubRepositoryFeature.*",
	} {
		entries, err := kindContainer.Directory("/resources/firestartr-crs/").Glob(ctx, kind)
		if err != nil {
			panic(fmt.Sprintf("Failed to get glob entries: %s", err))
		}
		for _, entry := range entries {
			kindContainer = m.ApplyCrAndWaitForProvisioned(
				ctx, kindContainer,
				fmt.Sprintf("/resources/firestartr-crs/%s", entry),
			)
		}
	}

	return kindContainer
}

func (m *FirestartrBootstrap) ApplyCrAndWaitForProvisioned(
	ctx context.Context,
	kindContainer *dagger.Container,
	entry string,
) *dagger.Container {

	crFile := kindContainer.File(entry)

	crContent, err := crFile.Contents(ctx)
	if err != nil {
		panic(fmt.Sprintf("Failed to get file contents: %s", err))
	}

	cr := &Cr{}
	err = yaml.Unmarshal([]byte(crContent), cr)
	if err != nil {
		panic(fmt.Sprintf("Failed to unmarshal CR: %s", err))
	}

	kindContainer, err = kindContainer.
		WithEnvVariable("BUST_CACHE", time.Now().String()).
		WithExec([]string{
			"kubectl",
			"apply",
			"-f", entry,
		}).
		WithExec([]string{
			"kubectl",
			"wait",
			"--for=condition=PROVISIONED=True",
			fmt.Sprintf("%s/%s", getSingularByKind(cr.Kind), cr.Metadata.Name),
			"--timeout=180s",
		}).
		Sync(ctx)

	if err != nil {
		m.FailedCrs = append(m.FailedCrs, cr)
	} else {
		m.ProvisionedCrs = append(m.ProvisionedCrs, cr)
	}

	return kindContainer
}

func GetKind(

	dockerSocket *dagger.Socket,

	kindSvc *dagger.Service,

) *dagger.Container {

	return dag.Kind(

		dockerSocket,
		kindSvc,
		dagger.KindOpts{

			ClusterName: "bootstrap-firestartr",
		}).Container()
}

func getSingularByKind(kind string) string {

	mapSingular := map[string]string{
		"FirestartrGithubRepository":        "githubrepository",
		"FirestartrGithubGroup":             "githubgroup",
		"FirestartrTerraformWorkspace":      "terraformworkspace",
		"FirestartrGithubMembership":        "githubmembership",
		"FirestartrGithubRepositoryFeature": "githubrepositoryfeature",
	}

	if singular, ok := mapSingular[kind]; ok {
		return singular
	} else {
		panic(fmt.Sprintf("No singular found for kind: %s", kind))
	}

}
