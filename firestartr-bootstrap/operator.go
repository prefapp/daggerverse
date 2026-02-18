package main

import (
	"context"
	"dagger/firestartr-bootstrap/internal/dagger"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"
)

func (m *FirestartrBootstrap) RunOperator(
	ctx context.Context,
	kindContainer *dagger.Container,
) (*dagger.Container, error) {

	renderedCrsDir, err := m.RenderCrs(ctx, kindContainer.Directory("/import"))
	if err != nil {
		return nil, err
	}

	kindContainer = kindContainer.
		WithDirectory("/resources", renderedCrsDir)

	kindContainer, err = m.ApplyFirestartrCrs(
		ctx,
		kindContainer,
		"/resources/firestartr-crs/infra",
		[]string{"ExternalSecret.*"},
	)
	if err != nil {
		return nil, err
	}

	kindContainer, err = m.ApplyFirestartrCrs(
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
	if err != nil {
		return nil, err
	}

	return kindContainer, nil

}

func (m *FirestartrBootstrap) InstallHelmAndExternalSecrets(
	ctx context.Context,
	kindContainer *dagger.Container,
) (*dagger.Container, error) {

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
		errMsg := extractErrorMessage(err, "Failed to install Helm chart for External Secrets")
		return nil, errors.New(errMsg)
	}

	return kindContainerWithSecrets, nil
}

func (m *FirestartrBootstrap) InstallInitialCRsAndBuildHelmValues(
	ctx context.Context,
	kindContainer *dagger.Container,
) (*dagger.Container, error) {
	initialCrsTemplate, err := m.RenderInitialCrs(ctx,
		dag.CurrentModule().
			Source().
			File("templates/initial_crs.tmpl"),
	)
	if err != nil {
		return nil, err
	}

	initialCrsDir, err := m.SplitRenderedCrsInFiles(initialCrsTemplate)
	if err != nil {
		return nil, err
	}

	helmValues, err := m.BuildHelmValues(ctx)
	if err != nil {
		return nil, err
	}

	kindContainer, err = kindContainer.
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
			helmValues,
		).
		WithWorkdir("/charts/firestartr-init").
		WithExec([]string{"helm", "upgrade", "--install", "firestartr-init", ".", "--values", "values-file.yaml"}).
		Sync(ctx)

	if err != nil {
		errMsg := extractErrorMessage(err, "Failed to install Helm and initial CRs")
		return nil, errors.New(errMsg)
	}

	return kindContainer, nil
}

func (m *FirestartrBootstrap) ApplyFirestartrCrs(
	ctx context.Context,
	kindContainer *dagger.Container,
	crsDirectoryPath string,
	crsToApplyList []string,
) (*dagger.Container, error) {

	for _, kind := range crsToApplyList {
		g, ctx := errgroup.WithContext(ctx)

		entries, err := kindContainer.Directory(crsDirectoryPath).Glob(ctx, kind)
		if err != nil {
			return nil, fmt.Errorf("Failed to get glob entries: %s", err)
		}

		for _, entry := range entries {
			g.Go(func() error {
				kindContainer, err = m.ApplyCrAndWaitForProvisioned(
					ctx, kindContainer,
					fmt.Sprintf("%s/%s", crsDirectoryPath, entry),
					kind != "ExternalSecret.*",
				)

				return err
			})
		}

		err = g.Wait()
		if err != nil {
			return nil, fmt.Errorf("Failed to apply CRs of kind %s: %w", kind, err)
		}
	}

	allGroupGetExitCode, err := kindContainer.
		WithExec([]string{
			"kubectl",
			"get",
			"githubgroup",
			fmt.Sprintf("%s-all-c8bc0fd3-78e1-42e0-8f5c-6b0bb13bb669", m.GhOrg),
		}, dagger.ContainerWithExecOpts{
			RedirectStdout: "/tmp/stdout",
			RedirectStderr: "/tmp/stderr",
			Expect:         "ANY",
		}).
		ExitCode(ctx)

	if err != nil {
		return nil, err
	}

	if allGroupGetExitCode == 0 {
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
			return nil, err
		}
	}

	return kindContainer, nil
}

func (m *FirestartrBootstrap) ApplyCrAndWaitForProvisioned(
	ctx context.Context,
	kindContainer *dagger.Container,
	entry string,
	waitForProvisioned bool,
) (*dagger.Container, error) {

	crFile := kindContainer.File(entry)

	crContent, err := crFile.Contents(ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to get file contents: %s", err)
	}

	cr := &Cr{}
	err = yaml.Unmarshal([]byte(crContent), cr)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal CR: %s", err)
	}

	kindContainer = kindContainer.
		WithEnvVariable("BUST_CACHE", time.Now().String()).
		WithExec([]string{
			"kubectl",
			"apply",
			"-f", entry,
		})

	if waitForProvisioned {
		singularKind, err := getSingularByKind(cr.Kind)
		if err != nil {
			return nil, err
		}

		kindContainer = kindContainer.
			WithExec([]string{
				"kubectl",
				"wait",
				"--for=condition=PROVISIONED=True",
				fmt.Sprintf("%s/%s", singularKind, cr.Metadata.Name),
				"--timeout=10h",
			})
	}

	kindContainer, err = kindContainer.Sync(ctx)

	if err != nil {
		m.FailedCrs = append(m.FailedCrs, cr)
	} else {
		m.ProvisionedCrs = append(m.ProvisionedCrs, cr)
	}

	return kindContainer, nil
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

func getSingularByKind(kind string) (string, error) {

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
		return singular, nil
	} else {
		return "", fmt.Errorf("No singular found for kind: %s", kind)
	}

}
