package main

import (
    "context"
	"dagger/firestartr-bootstrap/internal/dagger"
    "strings"
    "strconv"
    "fmt"
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
	kindSvc	*dagger.Service,
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

    ctn,err := dag.Container().
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
	kindSvc	*dagger.Service,
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
	kindSvc	*dagger.Service,
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
	kindSvc	*dagger.Service,
    cacheVolume *dagger.CacheVolume,

) *dagger.Container {

    kindContainer := m.CmdInitGithubAppsMachinery(ctx, kubeconfig, kindSvc)

	kindContainer = m.InstallInitialCRsAndBuildHelmValues(ctx, kindContainer)

    tokenSecret, err := m.GenerateGithubToken(ctx)
    if err != nil {
    	panic(err)
    }

	alreadyCreatedReposList := []string{}
	if m.PreviousCrsDir == nil {
		// if any of the CRs already exist, we skip their creation
        alreadyCreatedReposList, err = m.CheckAlreadyCreatedRepositories(
			ctx,
			tokenSecret,
		)
		if err != nil {
			panic(err)
		}
	}

    fmt.Println(cacheVolume)

	kindContainer = m.RunImporter(ctx, kindContainer, alreadyCreatedReposList)
	kindContainer = m.RunOperator(ctx, kindContainer)
	kindContainer = m.UpdateSecretStoreRef(ctx, kindContainer)

    return kindContainer.
        WithMountedCache("/cache", cacheVolume).
		WithExec([]string{
			"cp", "-a", "/import", "/cache",
		}).
		WithExec([]string{
			"cp", "-a", "/resources", "/cache",
		})
}

func (m *FirestartrBootstrap) CmdRollback(
    ctx context.Context,
	kubeconfig *dagger.Directory,
	kindSvc	*dagger.Service,
){

    kindContainer := m.CreateBridgeContainer(ctx, kubeconfig, kindSvc)

	err := m.ProcessArtifactsByKind(

		ctx,
		kindContainer,
	)

	if err != nil {

		panic(err)
	}

}
