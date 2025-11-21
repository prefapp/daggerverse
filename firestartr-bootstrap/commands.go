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
) *dagger.Container {

	clusterName := "kind"

	ep, err := kindSvc.Endpoint(ctx)
	if err != nil {
		panic(err)
	}

	port, err := strconv.Atoi(strings.Split(ep, ":")[1])
	if err != nil {
		panic(err)
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
		panic(err)
	}

	return ctn
}

func (m *FirestartrBootstrap) CmdValidateBootstrap(
	ctx context.Context,
) error {

	err := m.ValidateBootstrap(ctx)
	if err != nil {
		panic(err)
	}

	return nil
}

func (m *FirestartrBootstrap) CmdInitSecretsMachinery(
	ctx context.Context,
	kubeconfig *dagger.Directory,
	kindSvc *dagger.Service,
) *dagger.Container {

	kindContainer := m.CreateBridgeContainer(ctx, kubeconfig, kindSvc)

	kindContainer = m.InstallHelmAndExternalSecrets(ctx, kindContainer)
	kindContainerWithSecrets, err := m.CreateKubernetesSecrets(ctx, kindContainer)

	if err != nil {
		panic(err)
	}

	return kindContainerWithSecrets
}

func (m *FirestartrBootstrap) CmdInitGithubAppsMachinery(
	ctx context.Context,
	kubeconfig *dagger.Directory,
	kindSvc *dagger.Service,
) *dagger.Container {

	kindContainer := m.CreateBridgeContainer(ctx, kubeconfig, kindSvc)

	m.PopulateGithubAppCredsFromSecrets(ctx, kindContainer)

	tokenSecret, err := m.GenerateGithubToken(ctx)
	if err != nil {
		panic(err)
	}

	m.Bootstrap.BotName = m.Creds.GithubApp.BotName
	m.Bootstrap.HasFreePlan, err = m.OrgHasFreePlan(ctx, tokenSecret)
	if err != nil {
		panic(err)
	}

	err = m.CheckIfOrgAllGroupExists(ctx, tokenSecret)
	if err != nil {
		panic(err)
	}

	return kindContainer
}

// calls CmdInitGithubAppsMachinery
func (m *FirestartrBootstrap) CmdImportResources(
	ctx context.Context,
	kubeconfig *dagger.Directory,
	kindSvc *dagger.Service,
	cacheVolume *dagger.CacheVolume,

) *dagger.Container {

	kindContainer := m.CmdInitGithubAppsMachinery(ctx, kubeconfig, kindSvc)

	kindContainer = m.InstallInitialCRsAndBuildHelmValues(ctx, kindContainer)

	tokenSecret, err := m.GenerateGithubToken(ctx)
	if err != nil {
		panic(err)
	}

	if m.PreviousCrsDir == nil {
		// if any of the CRs already exist, we skip their creation
		err = m.CheckAlreadyCreatedRepositories(ctx, tokenSecret)
		if err != nil {
			panic(err)
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
		panic(err)
	}

	return kindContainer
}

func (m *FirestartrBootstrap) CmdPushResources(
	ctx context.Context,
	kubeconfig *dagger.Directory,
	kindSvc *dagger.Service,
	cacheVolume *dagger.CacheVolume,

) *dagger.Container {

	kindContainer := m.CmdInitGithubAppsMachinery(ctx, kubeconfig, kindSvc)

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

	return kindContainer
}

func (m *FirestartrBootstrap) CmdPushDeployment(
	ctx context.Context,
) *dagger.Container {

	deploymentDir, err := m.CreateDeployment(ctx)

	if err != nil {
		panic(err)
	}

	return dag.Container().
		From("busybox").
		WithMountedDirectory("/deployment", deploymentDir)
}

func (m *FirestartrBootstrap) CmdPushArgo(
	ctx context.Context,
) *dagger.Container {

    argocd, err := m.AddArgoCDSecrets(ctx)

	if err != nil {
		panic(err)
	}

    return dag.Container().
        From("busybox").
        WithMountedDirectory("/argocd", argocd)

	//deploymentDir, err := m.CreateArgCDApplications(ctx)

	//if err != nil {
	//	panic(err)
	//}

	//return dag.Container().
	//	From("busybox").
	//	WithMountedDirectory("/deployment", deploymentDir)
}

func (m *FirestartrBootstrap) CmdRollback(
	ctx context.Context,
	kubeconfig *dagger.Directory,
	kindSvc *dagger.Service,
) string {

	kindContainer := m.CreateBridgeContainer(ctx, kubeconfig, kindSvc)

	output, err := m.ProcessArtifactsByKind(
		ctx,
		kindContainer,
	)

	if err != nil {

		panic(err)
	}

	return output

}

func (m *FirestartrBootstrap) CmdRunBootstrap(
	ctx context.Context,
	kubeconfig *dagger.Directory,
	kindSvc *dagger.Service,
) {

	persistentVolume := m.CmdCreatePersistentVolume(ctx, "firestartr-init")

	m.CmdValidateBootstrap(ctx)

	m.CmdInitSecretsMachinery(ctx, kubeconfig, kindSvc)

	m.CmdInitGithubAppsMachinery(ctx, kubeconfig, kindSvc)

	m.CmdImportResources(ctx, kubeconfig, kindSvc, persistentVolume)

	m.CmdPushResources(ctx, kubeconfig, kindSvc, persistentVolume)

	m.CmdPushDeployment(ctx)

	m.CmdPushArgo(ctx)

}
