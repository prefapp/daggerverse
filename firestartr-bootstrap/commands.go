package main

import (
	"context"
	"dagger/firestartr-bootstrap/internal/dagger"
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
) (*dagger.Container, error) {

	clusterName := "kind"

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
			"set-cluster", fmt.Sprintf("kind-%s", clusterName), fmt.Sprintf("--server=https://localhost:%d", port)},
		).
		WithExec([]string{
			"curl",
			"https://prefapp.github.io/gitops-k8s/index.yaml",
			"-o",
			"/tmp/crds.yaml",
		}).
		WithExec([]string{"kubectl", "apply", "-f", "/tmp/crds.yaml"}).
		Sync(ctx)

	if err != nil {
		return nil, err
	}

	return ctn, nil
}

func (m *FirestartrBootstrap) CmdValidateBootstrap(
	ctx context.Context,
	kubeconfig *dagger.Directory,
	kindSvc *dagger.Service,
) (string, error) {

	err := m.ValidateBootstrap(ctx, kubeconfig, kindSvc)
	if err != nil {
		errorMessage := PrepareAndPrintError(
			ctx,
			"CmdValidateBootstrap",
			"An error ocurred validating the context and bootstrap conditions",
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
) (string, error) {

	kindContainer, err := m.CreateBridgeContainer(ctx, kubeconfig, kindSvc)
	if err != nil {
		errorMessage := PrepareAndPrintError(
			ctx,
			"CmdInitSecretsMachinery",
			"An error ocurred while preparing the external-secrets machinery and the push of secrets to the store",
			err,
		)

		return "", errorMessage
	}

	kindContainer = m.InstallHelmAndExternalSecrets(ctx, kindContainer)
	_, err = m.CreateKubernetesSecrets(ctx, kindContainer)

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

	if err != nil {
		errorMessage := PrepareAndPrintError(
			ctx,
			"CmdInitSecretsMachinery",
			"An error ocurred while preparing the external-secrets machinery and the push of secrets to the store",
			err,
		)

		return "", errorMessage
	}

	return m.UpdateSummaryAndRun(ctx, successMessage), nil

}

func (m *FirestartrBootstrap) CmdInitGithubAppsMachinery(
	ctx context.Context,
	kubeconfig *dagger.Directory,
	kindSvc *dagger.Service,
) (*dagger.Container, error) {

	kindContainer, err := m.CreateBridgeContainer(ctx, kubeconfig, kindSvc)
	if err != nil {
		errorMessage := PrepareAndPrintError(
			ctx,
			"CmdInitGithubAppsMachinery",
			"An error ocurred while preparing the github apps machinery",
			err,
		)

		return nil, errorMessage
	}

	m.PopulateGithubAppCredsFromSecrets(ctx, kindContainer)

	tokenSecret, err := m.GenerateGithubToken(ctx)
	if err != nil {
		errorMessage := PrepareAndPrintError(
			ctx,
			"CmdInitGithubAppsMachinery",
			"An error ocurred while preparing the github apps machinery",
			err,
		)

		return nil, errorMessage
	}

	m.Bootstrap.BotName = m.Creds.GithubApp.BotName
	m.Bootstrap.HasFreePlan, err = m.OrgHasFreePlan(ctx, tokenSecret)
	if err != nil {
		errorMessage := PrepareAndPrintError(
			ctx,
			"CmdInitGithubAppsMachinery",
			"An error ocurred while preparing the github apps machinery",
			err,
		)

		return nil, errorMessage
	}

	err = m.CheckIfOrgAllGroupExists(ctx, tokenSecret)
	if err != nil {
		errorMessage := PrepareAndPrintError(
			ctx,
			"CmdInitGithubAppsMachinery",
			"An error ocurred while preparing the github apps machinery",
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
	cacheVolume *dagger.CacheVolume,
) string {
	kindContainer, err := m.CmdInitGithubAppsMachinery(ctx, kubeconfig, kindSvc)
	if err != nil {
		return err.Error()
	}

	kindContainer = m.InstallInitialCRsAndBuildHelmValues(ctx, kindContainer)

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
		return fmt.Sprintf("Error busting cache volume for resources: %s", err)
	}

	tokenSecret, err := m.GenerateGithubToken(ctx)
	if err != nil {
		return err.Error()
	}

	if m.PreviousCrsDir == nil {
		// if any of the CRs already exist, we skip their creation
		err = m.CheckAlreadyCreatedRepositories(ctx, tokenSecret)
		if err != nil {
			return err.Error()
		}
	}

	kindContainer = m.RunImporter(ctx, kindContainer)
	kindContainer = m.RunOperator(ctx, kindContainer)
	kindContainer = m.UpdateSecretStoreRef(ctx, kindContainer)
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
		return err.Error()
	}

	return m.UpdateSummaryAndRunForImportResourcesStep(ctx, kindContainer)
}

func (m *FirestartrBootstrap) CmdPushResources(
	ctx context.Context,
	kubeconfig *dagger.Directory,
	kindSvc *dagger.Service,
	cacheVolume *dagger.CacheVolume,

) string {

	kindContainer, err := m.CmdInitGithubAppsMachinery(ctx, kubeconfig, kindSvc)
	if err != nil {
		return err.Error()
	}

	kindContainer = kindContainer.WithMountedCache(
		"/mnt/",
		cacheVolume,
	).
		WithExec([]string{
			"cp", "-a", "/mnt/resources", "/",
		})

	m.PushCrsFiles(
		ctx,
		kindContainer,
	)

	return m.UpdateSummaryAndRunForPushResourcesStep(ctx, kindContainer)
}

func (m *FirestartrBootstrap) CmdPushDeployment(
	ctx context.Context,
) string {

	_, err := m.CreateDeployment(ctx)

	if err != nil {
		return err.Error()
	}

	return m.UpdateSummaryAndRunForPushDeploymentStep(
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
}

func (m *FirestartrBootstrap) CmdPushStateSecrets(
	ctx context.Context,
	kubeconfig *dagger.Directory,
	kindSvc *dagger.Service,
	cacheVolume *dagger.CacheVolume,
) string {

	kindContainer, err := m.CmdInitGithubAppsMachinery(ctx, kubeconfig, kindSvc)
	if err != nil {
		return err.Error()
	}

	tokenSecret, err := m.GenerateGithubToken(ctx)
	if err != nil {
		return err.Error()
	}

	m.Bootstrap.BotName = m.Creds.GithubApp.BotName
	m.Bootstrap.HasFreePlan, err = m.OrgHasFreePlan(ctx, tokenSecret)
	if err != nil {
		return err.Error()
	}

	if !m.Bootstrap.HasFreePlan {
		err = m.SetOrgVariables(ctx, tokenSecret, kindContainer)
		if err != nil {
			return err.Error()
		}

		err = m.SetOrgSecrets(ctx, tokenSecret, kindContainer)
		if err != nil {
			return err.Error()
		}
	} else {
		return fmt.Sprintf("%s org has a free plan, org secrets are not available", m.Bootstrap.Org)
	}

	for _, component := range m.Bootstrap.Components {
		if len(component.Labels) > 0 {
			err = m.CreateLabelsInRepo(ctx, component.Name, component.Labels, tokenSecret)

			if err != nil {
				return err.Error()
			}
		}
	}

	successMessage := `
=====================================================
            🔐⚙️ ORG STATE SECRETS PUSHED ⚙️🔐
=====================================================
GitHub access machinery initialized, organization plan
validated, and required state secrets and variables
have been successfully configured.
`

	return m.UpdateSummaryAndRun(ctx, successMessage)
}

func (m *FirestartrBootstrap) CmdPushArgo(
	ctx context.Context,
) string {

	_, err := m.AddArgoCDSecrets(ctx)

	if err != nil {
		return err.Error()
	}

	_, err = m.CreateArgCDApplications(ctx)

	if err != nil {
		return err.Error()
	}

	return m.UpdateSummaryAndRunForPushArgoCDStep(
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
}

func (m *FirestartrBootstrap) CmdRollback(
	ctx context.Context,
	kubeconfig *dagger.Directory,
	kindSvc *dagger.Service,
) string {

	m.CmdValidateBootstrap(ctx, kubeconfig, kindSvc)

	kindContainer, err := m.CreateBridgeContainer(ctx, kubeconfig, kindSvc)
	if err != nil {
		return err.Error()
	}

	output, err := m.ProcessArtifactsByKind(ctx, kindContainer)
	if err != nil {
		return err.Error()
	}

	return m.UpdateSummaryAndRunForRollbackStep(ctx, output)

}

func (m *FirestartrBootstrap) CmdRunBootstrap(
	ctx context.Context,
	kubeconfig *dagger.Directory,
	kindSvc *dagger.Service,
) {

	persistentVolume := m.CmdCreatePersistentVolume(ctx, "firestartr-init")

	m.CmdValidateBootstrap(ctx, kubeconfig, kindSvc)

	m.CmdInitSecretsMachinery(ctx, kubeconfig, kindSvc)

	m.CmdInitGithubAppsMachinery(ctx, kubeconfig, kindSvc)

	m.CmdImportResources(ctx, kubeconfig, kindSvc, persistentVolume)

	m.CmdPushResources(ctx, kubeconfig, kindSvc, persistentVolume)

	m.CmdPushStateSecrets(ctx, kubeconfig, kindSvc, persistentVolume)

	//m.CmdPushDeployment(ctx)

	//m.CmdPushArgo(ctx)

}
