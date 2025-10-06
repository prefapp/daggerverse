package main

import (
	"context"
	"dagger/firestartr-bootstrap/internal/dagger"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type FirestartrBootstrap struct {
	Bootstrap          *Bootstrap
	BootstrapFile      *dagger.File
	CredentialsSecret  *dagger.Secret
	GhOrg              string
	Creds              *CredsFile
	CredsFileContent   string
	GeneratedGhToken   *dagger.Secret
	RenderedCrs        []*Cr
	ProvisionedCrs     []*Cr
	FailedCrs          []*Cr
	PreviousCrsDir     *dagger.Directory
	ClaimsDotConfigDir *dagger.Directory
	CrsDotConfigDir    *dagger.Directory
	IncludeAllGroup    bool
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

	claimsDotConfigDir, err := getClaimsDotConfigDir(ctx, bootstrap)
	if err != nil {
		panic(err)
	}

	defaultsInterface := CrsDefaultsData{
		GithubAppProviderConfigName:     creds.GithubApp.ProviderConfigName,
		CloudProviderProviderConfigName: creds.CloudProvider.ProviderConfigName,
		DefaultBranch:                   bootstrap.DefaultBranch,
	}

	crsDotConfigDir, err := getCrsDotConfigDir(ctx, bootstrap, defaultsInterface)
	if err != nil {
		panic(err)
	}

	return &FirestartrBootstrap{
		Bootstrap:          bootstrap,
		BootstrapFile:      bootstrapFile,
		CredentialsSecret:  credentialsSecret,
		GhOrg:              creds.GithubApp.Owner,
		Creds:              creds,
		CredsFileContent:   credsFileContent,
		PreviousCrsDir:     previousCrsDir,
		ClaimsDotConfigDir: claimsDotConfigDir,
		CrsDotConfigDir:    crsDotConfigDir,
	}, nil
}

func (m *FirestartrBootstrap) RunBootstrap(
	ctx context.Context,
	dockerSocket *dagger.Socket,
	kindSvc *dagger.Service,
) *dagger.Container {

	err := m.ValidateBootstrapFile(ctx, m.BootstrapFile)
	if err != nil {
		panic(err)
	}

	err = m.ValidateCredentialsFile(ctx, m.CredsFileContent)
	if err != nil {
		panic(err)
	}

	kindContainer := m.InstallHelmAndExternalSecrets(ctx, dockerSocket, kindSvc)
	kindContainer, err = m.CreateKubernetesSecrets(ctx, kindContainer)

	if err != nil {
		panic(err)
	}

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
		if strings.Contains(err.Error(), "404") {
			panic(err)
		}
	}

	kindContainer = m.InstallInitialCRsAndBuildHelmValues(ctx, kindContainer)

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

	kindContainer = m.RunImporter(ctx, kindContainer, alreadyCreatedReposList)
	kindContainer = m.RunOperator(ctx, kindContainer)
	kindContainer = m.UpdateSecretStoreRef(ctx, kindContainer)

	if m.Bootstrap.PushFiles.Claims.Push {
		claimsDir := kindContainer.
			Directory("/resources/claims").
			WithoutFile(fmt.Sprintf("claims/groups/%s-all.yaml", m.GhOrg))

		err := m.PushDirToRepo(
			ctx,
			claimsDir,
			m.Bootstrap.PushFiles.Claims.Repo,
			tokenSecret,
		)
		if err != nil {
			panic(err)
		}

		dotConfig := dag.Directory().
			WithDirectory(".config", m.ClaimsDotConfigDir)

		err = m.PushDirToRepo(
			ctx,
			dotConfig,
			m.Bootstrap.PushFiles.Claims.Repo,
			tokenSecret,
		)
		if err != nil {
			panic(err)
		}
	}

	if m.Bootstrap.PushFiles.Crs.Providers.Github.Push {
		crsDir := kindContainer.Directory("/resources/firestartr-crs/github")

		err := m.PushDirToRepo(
			ctx,
			crsDir,
			m.Bootstrap.PushFiles.Crs.Providers.Github.Repo,
			tokenSecret,
		)
		if err != nil {
			panic(err)
		}

		dotConfig := dag.Directory().
			WithDirectory(".config", m.CrsDotConfigDir)

		err = m.PushDirToRepo(
			ctx,
			dotConfig,
			m.Bootstrap.PushFiles.Crs.Providers.Github.Repo,
			tokenSecret,
		)
		if err != nil {
			panic(err)
		}
	}

	if m.Bootstrap.PushFiles.Crs.Providers.Terraform.Push {
		crsDir := kindContainer.Directory("/resources/firestartr-crs/infra")

		err := m.PushDirToRepo(
			ctx,
			crsDir,
			m.Bootstrap.PushFiles.Crs.Providers.Terraform.Repo,
			tokenSecret,
		)
		if err != nil {
			panic(err)
		}
	}

	if !m.Bootstrap.HasFreePlan {
		err = m.SetOrgVariables(ctx, tokenSecret, kindContainer)
		if err != nil {
			panic(err)
		}

		err = m.SetOrgSecrets(ctx, tokenSecret, kindContainer)
		if err != nil {
			panic(err)
		}
	}

	for _, component := range m.Bootstrap.Components {
		if len(component.Labels) > 0 {
			err = m.CreateLabelsInRepo(ctx, component.Name, component.Labels, tokenSecret)

			if err != nil {
				panic(err)
			}
		}
	}

	return kindContainer
}
