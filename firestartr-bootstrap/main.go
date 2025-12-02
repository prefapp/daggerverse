package main

import (
	"context"
	"dagger/firestartr-bootstrap/internal/dagger"
	"fmt"
	"log"
	"strings"

	"gopkg.in/yaml.v3"
)

type FirestartrBootstrap struct {
	Bootstrap             *Bootstrap
	BootstrapFile         *dagger.File
	CredentialsSecret     *dagger.Secret
	GhOrg                 string
	Creds                 *CredsFile
	CredsFileContent      string
	GeneratedGhToken      *dagger.Secret
	RenderedCrs           []*Cr
	ProvisionedCrs        []*Cr
	FailedCrs             []*Cr
	PreviousCrsDir        *dagger.Directory
	ClaimsDotConfigDir    *dagger.Directory
	CrsDotConfigDir       *dagger.Directory
	IncludeAllGroup       bool
	ExpectedAWSParameters []string
}

// baseTemplates holds the parameter paths with placeholders.
var baseTemplates = []string{
	"/firestartr/<client>/fs-<client>-admin/<github_org>/app-installation-id",
	"/firestartr/<client>/fs-<client>-argocd/<github_org>/app-installation-id",
	"/firestartr/<client>/fs-<client>-state/<github_org>/app-installation-id",
	"/firestartr/<client>/fs-<client>-checks/<github_org>/app-installation-id",
	"/firestartr/<client>/fs-<client>-import/<github_org>/app-installation-id",

	"/firestartr/<client>/fs-<client>-admin/pem",
	"/firestartr/<client>/fs-<client>-argocd/pem",
	"/firestartr/<client>/fs-<client>-state/pem",
	"/firestartr/<client>/fs-<client>-checks/pem",
	"/firestartr/<client>/fs-<client>-import/pem",

	"/firestartr/<client>/fs-<client>-admin/app-id",
	"/firestartr/<client>/fs-<client>-argocd/app-id",
	"/firestartr/<client>/fs-<client>-state/app-id",
	"/firestartr/<client>/fs-<client>-checks/app-id",
	"/firestartr/<client>/fs-<client>-import/app-id",

	"/firestartr/<client>/fs-<client>/pem",
	"/firestartr/<client>/fs-<client>/app-id",
	"/firestartr/<client>/fs-<client>/<github_org>/app-installation-id",
}

func New(
	ctx context.Context,
	// +optional
	// +defaultPath="fixtures/Bootstrapfile.yaml"
	bootstrapFile *dagger.File,
	// +optional
	previousCrsDir *dagger.Directory,
	// +required
	credentialsSecret *dagger.Secret,
) (*FirestartrBootstrap, error) {

	credsFileContent, err := credentialsSecret.Plaintext(ctx)
	if err != nil {
		return nil, err
	}

	creds, err := loadCredsFile(ctx, credentialsSecret)
	if err != nil {
		panic(err)
	}

	// load bootstrap file
	bootstrapContentFile, err := bootstrapFile.Contents(ctx)
	if err != nil {
		panic(err)
	}

	bootstrap := &Bootstrap{}
	err = yaml.Unmarshal([]byte(bootstrapContentFile), bootstrap)
	if err != nil {
		panic(err)
	}

	// ----------------------------------------------------
	// Autocalculate values
	// We need to calculate the webhook params
	// ----------------------------------------------------
	if bootstrap.Env == "pro" {
		bootstrap.WebhookUrl = fmt.Sprintf("https://%s.events.firestartr.dev", bootstrap.Customer)
	} else {
		bootstrap.WebhookUrl = fmt.Sprintf("https://%s.events.%s.firestartr.dev", bootstrap.Customer, bootstrap.Env)
	}
	bootstrap.WebhookSecretRef = fmt.Sprintf("/firestartr/%s/github-webhook/secret", bootstrap.Customer)

	// We need to calculate the bucket (if necessary)
	if creds.CloudProvider.Config.Bucket == nil {
		calculatedBucket := fmt.Sprintf("tfstate-%s", bootstrap.Customer)
		creds.CloudProvider.Config.Bucket = &calculatedBucket
	}

	bootstrap.PrefappBotPatSecretRef = fmt.Sprintf("/firestartr/%s/prefapp-bot-pat", bootstrap.Customer)
	bootstrap.FirestartrCliVersionSecretRef = fmt.Sprintf("/firestartr/%s/firestartr-cli-version", bootstrap.Customer)

	claimsDotConfigDir, err := getClaimsDotConfigDir(ctx, bootstrap)
	if err != nil {
		panic(err)
	}

	// calculate providers
	githubProviderConfigName := fmt.Sprintf("github-%s", bootstrap.Customer)
	backendConfigName := fmt.Sprintf("tfstate-%s", bootstrap.Customer)
	defaultsInterface := CrsDefaultsData{
		GithubAppProviderConfigName:     githubProviderConfigName,
		CloudProviderProviderConfigName: backendConfigName,
		DefaultBranch:                   bootstrap.DefaultBranch,
	}

	creds.CloudProvider.ProviderConfigName = backendConfigName
	creds.GithubApp.ProviderConfigName = githubProviderConfigName

	// calculate store name
	bootstrap.FinalSecretStoreName = fmt.Sprintf("%s-firestartr-secret-store", bootstrap.Customer)

	crsDotConfigDir, err := getCrsDotConfigDir(ctx, bootstrap, defaultsInterface)
	if err != nil {
		panic(err)
	}

	return &FirestartrBootstrap{
		Bootstrap:             bootstrap,
		BootstrapFile:         bootstrapFile,
		CredentialsSecret:     credentialsSecret,
		GhOrg:                 creds.GithubApp.Owner,
		Creds:                 creds,
		CredsFileContent:      credsFileContent,
		PreviousCrsDir:        previousCrsDir,
		ClaimsDotConfigDir:    claimsDotConfigDir,
		CrsDotConfigDir:       crsDotConfigDir,
		ExpectedAWSParameters: calculateParameters(bootstrap.Customer, bootstrap.Org),
	}, nil
}

func calculateParameters(customer string, githuborg string) []string {

	results := make([]string, 0, len(baseTemplates))

	clientPlaceholder := "<client>"
	githubOrgPlaceholder := "<github_org>"

	for _, template := range baseTemplates {

		path := strings.ReplaceAll(template, clientPlaceholder, customer)

		path = strings.ReplaceAll(path, githubOrgPlaceholder, githuborg)

		results = append(results, path)
	}

	return results
}

func (m *FirestartrBootstrap) ValidateBootstrap(
	ctx context.Context,
	kubeconfig *dagger.Directory,
	kindSvc *dagger.Service,
) error {

	log.Println("Validating bootstrap parameters")

	err := m.ValidateBootstrapFile(ctx, m.BootstrapFile)
	if err != nil {
		return err
	}

	err = m.ValidateCredentialsFile(ctx, m.CredsFileContent)
	if err != nil {
		return err
	}

	err = m.ValidateCliExistence(ctx)
	if err != nil {
		return err
	}

	err = m.ValidateExistenceOfNeededImages(ctx)
	if err != nil {
		return err
	}

	_, err = m.ValidateSTSCredentials(ctx)
	if err != nil {
		return err
	}

	err = m.ValidateBucket(ctx)
	if err != nil {
		return err
	}

	err = m.ValidateParameters(ctx, fmt.Sprintf("/firestartr/%s", m.Bootstrap.Customer))
	if err != nil {
		return err
	}

	err = m.ValidatePrefappBotPat(ctx)
	if err != nil {
		return err
	}

	err = m.ValidateOperatorPat(ctx)
	if err != nil {
		return err
	}

	err = m.ValidateKindKubernetesConnection(ctx, kubeconfig, kindSvc)
	if err != nil {
		return err
	}

	return nil

}

//func CreateKindContainer(
//    ctx context.Context,
//	dockerSocket *dagger.Socket,
//	kindSvc *dagger.Service,
//) *dagger.Container{
//
//	kindContainer, err := GetKind(dockerSocket, kindSvc).
//		WithExec([]string{"apk", "add", "helm", "curl"}).
//		WithMountedDirectory("/charts",
//			dag.CurrentModule().
//				Source().
//				Directory("helm"),
//		).
//		WithEnvVariable("BUST_CACHE", time.Now().String()).
//		WithExec([]string{
//			"curl",
//			"https://prefapp.github.io/gitops-k8s/index.yaml",
//			"-o",
//			"/tmp/crds.yaml",
//		}).
//		WithExec([]string{"kubectl", "apply", "-f", "/tmp/crds.yaml"}).
//        Sync(ctx)
//
//    if err != nil {
//        panic(err)
//    }
//
//    return kindContainer
//}

//func (m *FirestartrBootstrap) RunBootstrap(
//	ctx context.Context,
//	dockerSocket *dagger.Socket,
//	kindSvc *dagger.Service,
//
//) *dagger.Container {
//
//    err := m.ValidateBootstrap(ctx)
//    if err != nil {
//        panic(err)
//    }
//
//	//kindContainer := m.InstallHelmAndExternalSecrets(ctx, dockerSocket, kindSvc)
//	//kindContainer, err = m.CreateKubernetesSecrets(ctx, kindContainer)
//
//	//if err != nil {
//	//	panic(err)
//	//}
//
//	//m.PopulateGithubAppCredsFromSecrets(ctx, kindContainer)
//
//	//tokenSecret, err := m.GenerateGithubToken(ctx)
//	//if err != nil {
//	//	panic(err)
//	//}
//
//	//m.Bootstrap.BotName = m.Creds.GithubApp.BotName
//	//m.Bootstrap.HasFreePlan, err = m.OrgHasFreePlan(ctx, tokenSecret)
//	//if err != nil {
//	//	panic(err)
//	//}
//
//	//err = m.CheckIfOrgAllGroupExists(ctx, tokenSecret)
//	//if err != nil {
//	//	panic(err)
//	//}
//
//	//kindContainer = m.InstallInitialCRsAndBuildHelmValues(ctx, kindContainer)
//
//	//alreadyCreatedReposList := []string{}
//	//if m.PreviousCrsDir == nil {
//	//	// if any of the CRs already exist, we skip their creation
//	//	alreadyCreatedReposList, err = m.CheckAlreadyCreatedRepositories(
//	//		ctx,
//	//		tokenSecret,
//	//	)
//	//	if err != nil {
//	//		panic(err)
//	//	}
//	//}
//
//	//kindContainer = m.RunImporter(ctx, kindContainer, alreadyCreatedReposList)
//	//kindContainer = m.RunOperator(ctx, kindContainer)
//	//kindContainer = m.UpdateSecretStoreRef(ctx, kindContainer)
//
//	//if m.Bootstrap.PushFiles.Claims.Push {
//	//	claimsDir := kindContainer.
//	//		Directory("/resources/claims").
//	//		WithoutFile(fmt.Sprintf("claims/groups/%s-all.yaml", m.GhOrg))
//
//	//	err := m.PushDirToRepo(
//	//		ctx,
//	//		claimsDir,
//	//		m.Bootstrap.PushFiles.Claims.Repo,
//	//		tokenSecret,
//	//	)
//	//	if err != nil {
//	//		panic(err)
//	//	}
//
//	//	dotConfig := dag.Directory().
//	//		WithDirectory(".config", m.ClaimsDotConfigDir)
//
//	//	err = m.PushDirToRepo(
//	//		ctx,
//	//		dotConfig,
//	//		m.Bootstrap.PushFiles.Claims.Repo,
//	//		tokenSecret,
//	//	)
//	//	if err != nil {
//	//		panic(err)
//	//	}
//	//}
//
//	//if m.Bootstrap.PushFiles.Crs.Providers.Github.Push {
//	//	crsDir := kindContainer.Directory("/resources/firestartr-crs/github")
//
//	//	err := m.PushDirToRepo(
//	//		ctx,
//	//		crsDir,
//	//		m.Bootstrap.PushFiles.Crs.Providers.Github.Repo,
//	//		tokenSecret,
//	//	)
//	//	if err != nil {
//	//		panic(err)
//	//	}
//
//	//	dotConfig := dag.Directory().
//	//		WithDirectory(".config", m.CrsDotConfigDir)
//
//	//	err = m.PushDirToRepo(
//	//		ctx,
//	//		dotConfig,
//	//		m.Bootstrap.PushFiles.Crs.Providers.Github.Repo,
//	//		tokenSecret,
//	//	)
//	//	if err != nil {
//	//		panic(err)
//	//	}
//	//}
//
//	//if m.Bootstrap.PushFiles.Crs.Providers.Terraform.Push {
//	//	crsDir := kindContainer.Directory("/resources/firestartr-crs/infra")
//
//	//	err := m.PushDirToRepo(
//	//		ctx,
//	//		crsDir,
//	//		m.Bootstrap.PushFiles.Crs.Providers.Terraform.Repo,
//	//		tokenSecret,
//	//	)
//	//	if err != nil {
//	//		panic(err)
//	//	}
//	//}
//
//	//if !m.Bootstrap.HasFreePlan {
//	//	err = m.SetOrgVariables(ctx, tokenSecret, kindContainer)
//	//	if err != nil {
//	//		panic(err)
//	//	}
//
//	//	err = m.SetOrgSecrets(ctx, tokenSecret, kindContainer)
//	//	if err != nil {
//	//		panic(err)
//	//	}
//	//}
//
//	//for _, component := range m.Bootstrap.Components {
//	//	if len(component.Labels) > 0 {
//	//		err = m.CreateLabelsInRepo(ctx, component.Name, component.Labels, tokenSecret)
//
//	//		if err != nil {
//	//			panic(err)
//	//		}
//	//	}
//	//}
//
//	return kindContainer
//}
