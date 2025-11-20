package main

import (
	"context"
	"dagger/firestartr-bootstrap/internal/dagger"
	"fmt"
	"strings"
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

	kindContainer = m.ApplyFirestartrCrs(
		ctx,
		kindContainer,
		"/resources/firestartr-crs/infra",
		[]string{"ExternalSecret.*"},
	)
	kindContainer = m.ApplyFirestartrCrs(
		ctx,
		kindContainer,
		"/resources/firestartr-crs/github",
		[]string{
			"FirestartrGithubGroup.*",
			"FirestartrGithubRepository.*",
			"FirestartrGithubRepositorySecretsSection.*",
			"FirestartrGithubRepositoryFeature.*",
			"FirestartrGithubOrgWebhook.*",
		},
	)

	return kindContainer

}

func (m *FirestartrBootstrap) InstallHelmAndExternalSecrets(
	ctx context.Context,
	kindContainer *dagger.Container,
) *dagger.Container {

	kindContainerWithSecrets, err := kindContainer.
		WithExec([]string{
			"helm", "repo", "add",
			"external-secrets", "https://charts.external-secrets.io",
		}).
		WithExec([]string{
			"helm", "upgrade", "--install", "external-secrets",
			"external-secrets/external-secrets",
			"-n", "external-secrets",
			"--create-namespace",
		}).
		Sync(ctx)

	if err != nil {
		panic(err)
	}

	return kindContainerWithSecrets
}

func (m *FirestartrBootstrap) InstallInitialCRsAndBuildHelmValues(
	ctx context.Context,
	kindContainer *dagger.Container,
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

	return kindContainer.
		WithDirectory("/resources/initial-crs", initialCrsDir).
		WithMountedDirectory("/charts",
			dag.CurrentModule().
				Source().
				Directory("helm"),
		).
		WithExec([]string{
			"kubectl",
			"apply",
			"-f", "/resources/initial-crs",
		}).
		WithNewFile(
			"/charts/firestartr-init/values-file.yaml",
			m.BuildHelmValues(ctx),
		).
		WithWorkdir("/charts/firestartr-init").
		WithExec([]string{"helm", "upgrade", "--install", "firestartr-init", ".", "--values", "values-file.yaml"})
}

func (m *FirestartrBootstrap) ApplyFirestartrCrs(
	ctx context.Context,
	kindContainer *dagger.Container,
	crsDirectoryPath string,
	crsToApplyList []string,
) *dagger.Container {

	for _, kind := range crsToApplyList {
		entries, err := kindContainer.Directory(crsDirectoryPath).Glob(ctx, kind)
		if err != nil {
			panic(fmt.Sprintf("Failed to get glob entries: %s", err))
		}
		for _, entry := range entries {
			kindContainer = m.ApplyCrAndWaitForProvisioned(
				ctx, kindContainer,
				fmt.Sprintf("%s/%s", crsDirectoryPath, entry),
				kind != "ExternalSecret.*",
			)
		}
	}

	// let's patch the all group with the bootstrapped annotation
	err := patchCR(
		ctx,
		kindContainer,
		"githubgroup",
		fmt.Sprintf("%s-all-c8bc0fd3-78e1-42e0-8f5c-6b0bb13bb669", m.GhOrg),
		"default",
		"firestartr.dev/bootstrapped",
		"true",
	)

	if err != nil {
		panic(err)
	}

	return kindContainer
}

func (m *FirestartrBootstrap) ApplyCrAndWaitForProvisioned(
	ctx context.Context,
	kindContainer *dagger.Container,
	entry string,
	waitForProvisioned bool,
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

	if waitForProvisioned {
		kindContainer = kindContainer.
			WithExec([]string{
				"kubectl",
				"wait",
				"--for=condition=PROVISIONED=True",
				fmt.Sprintf("%s/%s", getSingularByKind(cr.Kind), cr.Metadata.Name),
				"--timeout=180s",
			})
	}

	kindContainer, err = kindContainer.Sync(ctx)

	if err != nil {
		m.FailedCrs = append(m.FailedCrs, cr)
	} else {
		m.ProvisionedCrs = append(m.ProvisionedCrs, cr)
	}

	return kindContainer
}

func patchCR(
	ctx context.Context,
	kindContainer *dagger.Container,
	resourceKind string,
	resourceName string,
	namespace string,
	annotationKey string,
	annotationValue string,

) error {

	resourceRef := fmt.Sprintf("%s/%s", resourceKind, resourceName)

	// The JSON string defines the patch: modify the 'metadata.annotations' map.
	patchJSON := fmt.Sprintf(`{"metadata":{"annotations":{"%s":"%s"}}}`, annotationKey, annotationValue)

	patchCommand := []string{
		"kubectl",
		"patch",
		resourceRef,
		"-n",
		namespace,
		"--type=merge", // Use strategic merge patch to safely update only the annotation field
		"-p",           // The patch data flag
		patchJSON,      // The JSON payload
	}

	_, err := kindContainer.
		WithExec(patchCommand).
		Stdout(ctx)

	if err != nil {
		// Capture stderr for better debugging
		errorOutput, _ := kindContainer.Stderr(ctx)
		return fmt.Errorf("kubectl patch failed for %s. Error: %s", resourceRef, strings.TrimSpace(errorOutput))
	}

	return nil
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
		"ExternalSecret":                           "",
		"FirestartrGithubRepository":               "githubrepository",
		"FirestartrGithubGroup":                    "githubgroup",
		"FirestartrTerraformWorkspace":             "terraformworkspace",
		"FirestartrGithubMembership":               "githubmembership",
		"FirestartrGithubRepositoryFeature":        "githubrepositoryfeature",
		"FirestartrGithubOrgWebhook":               "githuborgwebhook",
		"FirestartrGithubRepositorySecretsSection": "githubrepositorysecretssections",
	}

	if singular, ok := mapSingular[kind]; ok {
		return singular
	} else {
		panic(fmt.Sprintf("No singular found for kind: %s", kind))
	}

}
