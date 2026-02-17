package main

import (
	"context"
	"dagger/firestartr-bootstrap/internal/dagger"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func (m *FirestartrBootstrap) CmdCreatePersistentVolume(
	ctx context.Context,
	volumeName string,
) *dagger.CacheVolume {
	persistentVolume := dag.CacheVolume(volumeName)

	return persistentVolume
}

func (m *FirestartrBootstrap) CreateBridgeContainer(
	ctx context.Context,
	kubeconfig *dagger.Directory,
	kindSvc *dagger.Service,
	kindClusterName string,
) (*dagger.Container, error) {
	ep, err := kindSvc.Endpoint(ctx)
	if err != nil {
		return nil, err
	}

	port, err := strconv.Atoi(strings.Split(ep, ":")[1])
	if err != nil {
		return nil, err
	}

	ctn, err := dag.Container().
		From("alpine:latest").
		WithExec([]string{"apk", "add", "docker", "kubectl", "k9s", "curl", "helm"}).
		WithMountedDirectory("/root/.kube", kubeconfig).
		WithWorkdir("/workspace").
		WithServiceBinding("localhost", kindSvc).
		WithExec([]string{
			"kubectl", "config",
			"set-cluster", fmt.Sprintf("kind-%s", kindClusterName), fmt.Sprintf("--server=https://localhost:%d", port)},
		).
		WithExec([]string{
			"curl",
			"https://raw.githubusercontent.com/firestartr-pro/docs/refs/heads/main/site/raw/core/crds/latest/index.yaml",
			"-o",
			"/tmp/crds.yaml",
		}).
		WithExec([]string{"kubectl", "apply", "-f", "/tmp/crds.yaml"}).
		Sync(ctx)

	if err != nil {
		errMsg := extractErrorMessage(err, "Failed to create bridge container")
		return nil, errors.New(errMsg)
	}

	return ctn, nil
}

func (m *FirestartrBootstrap) CmdValidateBootstrap(
	ctx context.Context,
	kubeconfig *dagger.Directory,
	kindSvc *dagger.Service,
	kindClusterName string,
) (string, error) {
	err := m.ValidateBootstrap(ctx, kubeconfig, kindSvc, kindClusterName)
	if err != nil {
		errorMessage := PrepareAndPrintError(
			ctx,
			"CmdValidateBootstrap",
			"An error occurred validating the context and bootstrap conditions",
			err,
		)

		return "", errorMessage
	}

	successMessage := `
=====================================================
🎉 ALL VALIDATION CHECKS PASSED 🎉
=====================================================
The pipeline executed without detecting any fatal errors.
The environment, configuration, and state are considered valid.
`

	m.UpdateSummaryAndRun(ctx, successMessage)

	return m.ShowSummaryReport(ctx), nil
}

func (m *FirestartrBootstrap) CmdInitSecretsMachinery(
	ctx context.Context,
	kubeconfig *dagger.Directory,
	kindSvc *dagger.Service,
	kindClusterName string,
) (string, error) {
	kindContainer, err := m.CreateBridgeContainer(ctx, kubeconfig, kindSvc, kindClusterName)
	if err != nil {
		errorMessage := PrepareAndPrintError(
			ctx,
			"CmdInitSecretsMachinery",
			"An error occurred while creating the bridge container",
			err,
		)

		return "", errorMessage
	}

	kindContainer, err = m.InstallHelmAndExternalSecrets(ctx, kindContainer)
	if err != nil {
		errorMessage := PrepareAndPrintError(
			ctx,
			"CmdInitSecretsMachinery",
			"An error occurred while installing Helm and External Secrets",
			err,
		)

		return "", errorMessage
	}

	_, err = m.CreateKubernetesSecrets(ctx, kindContainer)
	if err != nil {
		errorMessage := PrepareAndPrintError(
			ctx,
			"CmdInitSecretsMachinery",
			"An error occurred while creating the Kubernetes secrets",
			err,
		)

		return "", errorMessage
	}

	successMessage := `
=====================================================
🔒 SECRETS MACHINERY INITIALIZED 🔒
=====================================================
Helm, External Secrets Operator, and all required
Kubernetes secrets have been successfully deployed.
List of secrets:
	- Bootstrap Secrets
	- Aws Secrets
List of push secrets:
    - Webhook secret
	- Prefapp bot secret
`

	return m.UpdateSummaryAndRun(ctx, successMessage), nil
}

func (m *FirestartrBootstrap) CmdInitGithubAppsMachinery(
	ctx context.Context,
	kubeconfig *dagger.Directory,
	kindSvc *dagger.Service,
	kindClusterName string,
) (*dagger.Container, error) {
	kindContainer, err := m.CreateBridgeContainer(ctx, kubeconfig, kindSvc, kindClusterName)
	if err != nil {
		errorMessage := PrepareAndPrintError(
			ctx,
			"CmdInitGithubAppsMachinery",
			"An error occurred while creating the bridge container",
			err,
		)

		return nil, errorMessage
	}

	err = m.PopulateGithubAppCredsFromSecrets(ctx, kindContainer)
	if err != nil {
		errorMessage := PrepareAndPrintError(
			ctx,
			"CmdInitGithubAppsMachinery",
			"An error occurred while populating the GitHub App credentials from the Kubernetes secrets",
			err,
		)

		return nil, errorMessage
	}

	tokenSecret, err := m.GenerateGithubToken(ctx)
	if err != nil {
		errorMessage := PrepareAndPrintError(
			ctx,
			"CmdInitGithubAppsMachinery",
			"An error occurred while generating the GitHub token",
			err,
		)

		return nil, errorMessage
	}

	m.Bootstrap.HasFreePlan, err = m.OrgHasFreePlan(ctx, tokenSecret)
	if err != nil {
		errorMessage := PrepareAndPrintError(
			ctx,
			"CmdInitGithubAppsMachinery",
			"An error occurred while getting the organization plan",
			err,
		)

		return nil, errorMessage
	}

	err = m.CheckIfOrgAllGroupExists(ctx, tokenSecret)
	if err != nil {
		errorMessage := PrepareAndPrintError(
			ctx,
			"CmdInitGithubAppsMachinery",
			fmt.Sprintf("An error occurred while checking if the %s-all group exists", m.Bootstrap.Org),
			err,
		)

		return nil, errorMessage
	}

	err = m.CheckIfDefaultGroupExists(ctx, tokenSecret)
	if err != nil {
		errorMessage := PrepareAndPrintError(
			ctx,
			"CmdInitGithubAppsMachinery",
			fmt.Sprintf("An error occurred while checking if the %s group exists", m.Bootstrap.DefaultGroup),
			err,
		)

		return nil, errorMessage
	}

	successMessage := `
=====================================================
🤖 GITHUB APPS MACHINERY VALIDATED 🤖
=====================================================
Access tokens generated, GitHub App credentials loaded,
and organization plans validated successfully.
The system is ready to interact with GitHub.
`
	m.UpdateSummaryAndRun(ctx, successMessage)

	return kindContainer, nil
}

// calls CmdInitGithubAppsMachinery
func (m *FirestartrBootstrap) CmdImportResources(
	ctx context.Context,
	kubeconfig *dagger.Directory,
	kindSvc *dagger.Service,
	kindClusterName string,
	cacheVolume *dagger.CacheVolume,
) (string, error) {
	kindContainer, err := m.CmdInitGithubAppsMachinery(ctx, kubeconfig, kindSvc, kindClusterName)
	if err != nil {
		return "", err
	}

	kindContainer, err = m.InstallInitialCRsAndBuildHelmValues(ctx, kindContainer)
	if err != nil {
		errorMessage := PrepareAndPrintError(
			ctx,
			"CmdImportResources",
			"An error occurred while installing Helm and initial CRs",
			err,
		)

		return "", errorMessage
	}

	// bust cache volume
	kindContainer, err = kindContainer.
		WithMountedCache("/cache", cacheVolume).
		WithExec([]string{
			"rm", "-rf", "/cache/import",
		}).
		WithExec([]string{
			"rm", "-rf", "/cache/resources",
		}).
		Sync(ctx)

	if err != nil {
		errorMessage := PrepareAndPrintError(
			ctx,
			"CmdImportResources",
			"An error occurred while clearing the cache volume",
			errors.New(extractErrorMessage(err, "Failed to clear cache volume")),
		)

		return "", errorMessage
	}

	tokenSecret, err := m.GenerateGithubToken(ctx)
	if err != nil {
		errorMessage := PrepareAndPrintError(
			ctx,
			"CmdImportResources",
			"An error occurred while generating the GitHub token",
			err,
		)

		return "", errorMessage
	}

	if m.PreviousCrsDir == nil {
		// if any of the CRs already exist, we skip their creation
		err = m.CheckAlreadyCreatedRepositories(ctx, tokenSecret)
		if err != nil {
			errorMessage := PrepareAndPrintError(
				ctx,
				"CmdImportResources",
				"A problem was detected while checking if any repository already exists",
				err,
			)

			return "", errorMessage
		}
	}

	err = m.ValidateWebhookNotExists(ctx, tokenSecret, m.Bootstrap.WebhookUrl)
	if err != nil {
		errorMessage := PrepareAndPrintError(
			ctx,
			"CmdImportResources",
			"A problem was detected while checking if the webhook already exists",
			err,
		)

		return "", errorMessage
	}

	err = m.EnableActionsToCreateAndApprovePullRequestsInOrg(ctx, tokenSecret)
	if err != nil {
		errorMessage := PrepareAndPrintError(
			ctx,
			"CmdImportResources",
			"An error occurred while enabling actions to create and approve pull requests in the organization",
			err,
		)

		return "", errorMessage
	}

	err = m.SetLatestFeatureVersionWhenNecessary(ctx, tokenSecret)
	if err != nil {
		errorMessage := PrepareAndPrintError(
			ctx,
			"CmdImportResources",
			"An error occurred while resolving the latest feature versions",
			err,
		)

		return "", errorMessage
	}

	kindContainer, err = m.RunImporter(ctx, kindContainer)
	if err != nil {
		errorMessage := PrepareAndPrintError(
			ctx,
			"CmdImportResources",
			"An error occurred while importing the organization's resources",
			err,
		)

		return "", errorMessage
	}

	kindContainer, err = m.RunOperator(ctx, kindContainer)
	if err != nil {
		errorMessage := PrepareAndPrintError(
			ctx,
			"CmdImportResources",
			"An error occurred while creating the necessary resources using the operator",
			err,
		)

		return "", errorMessage
	}

	kindContainer, err = m.UpdateSecretStoreRef(ctx, kindContainer)
	if err != nil {
		errorMessage := PrepareAndPrintError(
			ctx,
			"CmdImportResources",
			"An error occurred while updating the secret store references",
			err,
		)

		return "", errorMessage
	}

	kindContainer, err = kindContainer.
		WithMountedCache("/cache", cacheVolume).
		WithExec([]string{
			"cp", "-a", "/import", "/cache",
		}).
		WithExec([]string{
			"cp", "-a", "/resources/", "/cache",
		}).
		Sync(ctx)
	if err != nil {
		errorMessage := PrepareAndPrintError(
			ctx,
			"CmdImportResources",
			"An error occurred while updating the cache volume",
			errors.New(extractErrorMessage(err, "Failed to update cache volume")),
		)

		return "", errorMessage
	}

	summary, err := m.UpdateSummaryAndRunForImportResourcesStep(ctx, kindContainer)
	if err != nil {
		errorMessage := PrepareAndPrintError(
			ctx,
			"CmdImportResources",
			"An error occurred while updating the summary for this command",
			err,
		)

		return "", errorMessage
	}

	return summary, nil
}

func (m *FirestartrBootstrap) CmdPushResources(
	ctx context.Context,
	kubeconfig *dagger.Directory,
	kindSvc *dagger.Service,
	kindClusterName string,
	cacheVolume *dagger.CacheVolume,
) (string, error) {
	kindContainer, err := m.CmdInitGithubAppsMachinery(ctx, kubeconfig, kindSvc, kindClusterName)
	if err != nil {
		return "", err
	}

	kindContainer = kindContainer.
		WithMountedCache(
			"/mnt/",
			cacheVolume,
		).
		WithExec([]string{
			"cp", "-a", "/mnt/resources", "/",
		})

	err = m.PushBootstrapFiles(
		ctx,
		kindContainer,
	)
	if err != nil {
		errorMessage := PrepareAndPrintError(
			ctx,
			"CmdPushResources",
			"An error occurred while pushing the claims and CRs to the repositories",
			err,
		)

		return "", errorMessage
	}

	summary, err := m.UpdateSummaryAndRunForPushResourcesStep(ctx, kindContainer)
	if err != nil {
		errorMessage := PrepareAndPrintError(
			ctx,
			"CmdPushResources",
			"An error occurred while updating the summary for this command",
			err,
		)

		return "", errorMessage
	}

	return summary, nil
}

func (m *FirestartrBootstrap) CmdPushDeployment(
	ctx context.Context,
) (string, error) {
	_, err := m.CreateDeployment(ctx)
	if err != nil {
		errorMessage := PrepareAndPrintError(
			ctx,
			"CmdPushDeployment",
			"An error occurred while pushing the deployment to the app-firestartr repository",
			err,
		)

		return "", errorMessage
	}

	summary := m.UpdateSummaryAndRunForPushDeploymentStep(
		ctx,
		fmt.Sprintf(
			"https://github.com/firestartr-%s/app-firestartr",
			m.Bootstrap.Env,
		),
		fmt.Sprintf(
			"firestartr-%s  /  %s  /   %s",
			m.Bootstrap.Env,
			m.Bootstrap.Customer,
			m.Bootstrap.Env,
		),
	)

	return summary, nil
}

func (m *FirestartrBootstrap) CmdPushStateSecrets(
	ctx context.Context,
	kubeconfig *dagger.Directory,
	kindSvc *dagger.Service,
	kindClusterName string,
	cacheVolume *dagger.CacheVolume,
) (string, error) {
	kindContainer, err := m.CmdInitGithubAppsMachinery(ctx, kubeconfig, kindSvc, kindClusterName)
	if err != nil {
		return "", err
	}

	tokenSecret, err := m.GenerateGithubToken(ctx)
	if err != nil {
		errorMessage := PrepareAndPrintError(
			ctx,
			"CmdPushStateSecrets",
			"An error occurred while generating the GitHub token",
			err,
		)

		return "", errorMessage
	}

	for _, component := range m.Bootstrap.Components {
		if len(component.Labels) > 0 {
			err = m.CreateLabelsInRepo(ctx, component.Name, component.Labels, tokenSecret)

			if err != nil {
				errorMessage := PrepareAndPrintError(
					ctx,
					"CmdPushStateSecrets",
					fmt.Sprintf(
						"An error occurred while creating the labels for repo %s",
						component.Name,
					),
					err,
				)

				return "", errorMessage
			}
		}
	}

	m.Bootstrap.HasFreePlan, err = m.OrgHasFreePlan(ctx, tokenSecret)
	if err != nil {
		errorMessage := PrepareAndPrintError(
			ctx,
			"CmdPushStateSecrets",
			"An error occurred while checking if the organization has a free plan",
			err,
		)

		return "", errorMessage
	}

	if !m.Bootstrap.HasFreePlan {
		err = m.SetOrgVariables(ctx, tokenSecret, kindContainer)
		if err != nil {
			errorMessage := PrepareAndPrintError(
				ctx,
				"CmdPushStateSecrets",
				"An error occurred while creating the organization variables",
				err,
			)

			return "", errorMessage
		}

		err = m.SetOrgSecrets(ctx, tokenSecret, kindContainer)
		if err != nil {
			errorMessage := PrepareAndPrintError(
				ctx,
				"CmdPushStateSecrets",
				"An error occurred while creating the organization secrets",
				err,
			)

			return "", errorMessage
		}
	} else {
		return fmt.Sprintf("%s org has a free plan, org secrets are not available", m.Bootstrap.Org), nil
	}

	successMessage := `
=====================================================
            🔐⚙️ ORG STATE SECRETS PUSHED ⚙️🔐
=====================================================
GitHub access machinery initialized, organization plan
validated, and required state secrets and variables
have been successfully configured.
`

	return m.UpdateSummaryAndRun(ctx, successMessage), nil
}

func (m *FirestartrBootstrap) CmdPushArgo(
	ctx context.Context,
) (string, error) {
	_, err := m.AddArgoCDSecrets(ctx)
	if err != nil {
		errorMessage := PrepareAndPrintError(
			ctx,
			"CmdPushArgo",
			"An error occurred while adding ArgoCD secrets",
			err,
		)

		return "", errorMessage
	}

	_, err = m.CreateArgCDApplications(ctx)
	if err != nil {
		errorMessage := PrepareAndPrintError(
			ctx,
			"CmdPushArgo",
			"An error occurred while creating the ArgoCD applications",
			err,
		)

		return "", errorMessage
	}

	summary := m.UpdateSummaryAndRunForPushArgoCDStep(
		ctx,
		fmt.Sprintf(
			"https://github.com/firestartr-%s/state-argocd",
			m.Bootstrap.Env,
		),
		fmt.Sprintf(
			"https://github.com/firestartr-%s/state-sys-services",
			m.Bootstrap.Env,
		),
		fmt.Sprintf(
			"firestartr-%s  /  argo-configuration-secrets  ",
			m.Bootstrap.Env,
		),
	)

	return summary, nil
}

func (m *FirestartrBootstrap) CmdRollback(
	ctx context.Context,
	kubeconfig *dagger.Directory,
	kindSvc *dagger.Service,
	kindClusterName string,
) (string, error) {
	m.CmdValidateBootstrap(ctx, kubeconfig, kindSvc, kindClusterName)

	kindContainer, err := m.CreateBridgeContainer(ctx, kubeconfig, kindSvc, kindClusterName)
	if err != nil {
		errorMessage := PrepareAndPrintError(
			ctx,
			"CmdRollback",
			"An error occurred while creating the bridge container",
			err,
		)

		return "", errorMessage
	}

	output, err := m.ProcessArtifactsByKind(ctx, kindContainer)
	if err != nil {
		errorMessage := PrepareAndPrintError(
			ctx,
			"CmdRollback",
			"An error occurred during the rollback process",
			err,
		)

		return "", errorMessage
	}

	return m.UpdateSummaryAndRunForRollbackStep(ctx, output), nil
}

func (m *FirestartrBootstrap) CmdRunBootstrap(
	ctx context.Context,
	kubeconfig *dagger.Directory,
	kindSvc *dagger.Service,
	kindClusterName string,
) {
	persistentVolume := m.CmdCreatePersistentVolume(ctx, "firestartr-init")

	m.CmdValidateBootstrap(ctx, kubeconfig, kindSvc, kindClusterName)

	m.CmdInitSecretsMachinery(ctx, kubeconfig, kindSvc, kindClusterName)

	m.CmdInitGithubAppsMachinery(ctx, kubeconfig, kindSvc, kindClusterName)

	m.CmdImportResources(ctx, kubeconfig, kindSvc, kindClusterName, persistentVolume)

	m.CmdPushResources(ctx, kubeconfig, kindSvc, kindClusterName, persistentVolume)

	m.CmdPushStateSecrets(ctx, kubeconfig, kindSvc, kindClusterName, persistentVolume)

	//m.CmdPushDeployment(ctx)

	//m.CmdPushArgo(ctx)
}

func (m *FirestartrBootstrap) CmdDebugBootstrap(
	ctx context.Context,
	cacheVolume *dagger.CacheVolume,
) *dagger.Container {


    return dag.Container().
        From("alpine:3.21.3").
        WithMountedCache("/cache", cacheVolume)
}
