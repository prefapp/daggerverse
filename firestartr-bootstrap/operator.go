package main

import (
	"context"
	"dagger/firestartr-bootstrap/internal/dagger"
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

func (m *FirestartrBootstrap) RunOperator(
	ctx context.Context,
	kindContainer *dagger.Container,
) *dagger.Container {

	renderedCrsDir, err := m.RenderCrs(ctx, kindContainer.Directory("/import"))
	if err != nil {
		panic(err)
	}

	kindContainer = kindContainer.
		WithDirectory("/resources", renderedCrsDir)

	kindContainer = m.ApplyFirestartrCrs(ctx, kindContainer, "/resources/firestartr-crs/")

	return kindContainer

}

func (m *FirestartrBootstrap) InstallCRDsAndInitialCRs(
	ctx context.Context,
	dockerSocket *dagger.Socket,
	kindSvc *dagger.Service,
) *dagger.Container {
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

	kindContainer, err := GetKind(dockerSocket, kindSvc).
		WithExec([]string{"apk", "add", "helm", "curl"}).
		WithMountedDirectory("/charts",
			dag.CurrentModule().
				Source().
				Directory("helm"),
		).
		WithNewFile(
			"/charts/firestartr-init/values-file.yaml",
			m.BuildHelmValues(ctx),
		).
		WithEnvVariable("BUST_CACHE", time.Now().String()).
		WithExec([]string{
			"curl",
			"https://prefapp.github.io/gitops-k8s/index.yaml",
			"-o",
			"/tmp/crds.yaml",
		}).
		WithExec([]string{"kubectl", "apply", "-f", "/tmp/crds.yaml"}).
		WithWorkdir("/charts/firestartr-init").
		WithExec([]string{"helm", "upgrade", "--install", "firestartr-init", ".", "--values", "values-file.yaml"}).
		WithDirectory("/resources/initial-crs", initialCrsDir).
		WithExec([]string{
			"kubectl",
			"apply",
			"-f", "/resources/initial-crs",
		}).
		WithExec([]string{
			"helm", "repo", "add",
			"external-secrets", "https://charts.external-secrets.io",
		}).
		WithExec([]string{
			"helm", "install", "external-secrets",
			"external-secrets/external-secrets",
			"-n", "external-secrets",
			"--create-namespace",
		}).
		Sync(ctx)

	if err != nil {
		panic(err)
	}

	return kindContainer
}

func (m *FirestartrBootstrap) ApplyFirestartrCrs(
	ctx context.Context,
	kindContainer *dagger.Container,
	crsDirectoryPath string,
) *dagger.Container {

	for _, kind := range []string{
		// "ExternalSecret.*",
		"FirestartrGithubMembership.*",
		"FirestartrGithubGroup.*",
		"FirestartrGithubRepository.*",
		"FirestartrGithubRepositoryFeature.*",
	} {
		entries, err := kindContainer.Directory(crsDirectoryPath).Glob(ctx, kind)
		if err != nil {
			panic(fmt.Sprintf("Failed to get glob entries: %s", err))
		}
		for _, entry := range entries {
			kindContainer = m.ApplyCrAndWaitForProvisioned(
				ctx, kindContainer,
				fmt.Sprintf("%s/%s", crsDirectoryPath, entry),
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

	kindContainer = kindContainer.
		WithEnvVariable("BUST_CACHE", time.Now().String()).
		WithExec([]string{
			"kubectl",
			"apply",
			"-f", entry,
		})
		// .
		// WithExec([]string{
		// 	"kubectl",
		// 	"wait",
		// 	"--for=condition=PROVISIONED=True",
		// 	fmt.Sprintf("%s/%s", getSingularByKind(cr.Kind), cr.Metadata.Name),
		// 	"--timeout=180s",
		// }).
		// Sync(ctx)

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
		dagger.KindOpts{ClusterName: "bootstrap-firestartr"},
	).Container()
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
