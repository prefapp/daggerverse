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

	dockerSocket *dagger.Socket,

	kindSvc *dagger.Service,

) *dagger.Container {

	renderedCrsDir, err := m.RenderCrs(ctx)
	if err != nil {
		panic(err)
	}

	kindContainer := GetKind(dockerSocket, kindSvc).
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
		WithDirectory("/resources", renderedCrsDir).
		WithWorkdir("/charts/firestartr-init").
		WithExec([]string{"helm", "upgrade", "--install", "firestartr-init", ".", "--values", "values-file.yaml"}).
		WithExec([]string{
			"kubectl",
			"apply",
			"-f", "/resources/initial-crs",
		})

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
			kindContainer = m.ApplyCrAndWaitForProvisioned(ctx, kindContainer, fmt.Sprintf("/resources/firestartr-crs/%s", entry))
			if err != nil {
				panic(fmt.Sprintf("Failed to apply CR and wait for provisioned: %s", err))
			}
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
