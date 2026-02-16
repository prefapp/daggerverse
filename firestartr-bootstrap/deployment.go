package main

import (
	"context"
	"dagger/firestartr-bootstrap/internal/dagger"
	"fmt"
	"regexp"
)

func (m *FirestartrBootstrap) CreateDeployment(
	ctx context.Context,
) (*dagger.Directory, error) {

	deploymentRenderedDir, err := m.RenderDeployment(ctx)

	if err != nil {

		return nil, fmt.Errorf("rendering firestartr-app deployment data: %w", err)
	}

	tokenSecret := dag.SetSecret(
		"token",
		m.Creds.GithubApp.OperatorPat,
	)

	err = m.CreatePR(
		ctx,
		"app-firestartr",
		fmt.Sprintf("firestartr-%s", m.Bootstrap.Env),
		deploymentRenderedDir,
		fmt.Sprintf("automated-create-deployment-%s", m.Bootstrap.Customer),
		fmt.Sprintf("feat: add deployment for %s [automated]", m.Bootstrap.Customer),
		fmt.Sprintf("kubernetes/firestartr-%s/%s", m.Bootstrap.Env, m.Bootstrap.Customer),
		tokenSecret,
	)

	if err != nil {
		return nil, fmt.Errorf("error generating PR for firestartr-app deployment: %w", err)
	}

	return deploymentRenderedDir, nil

}

func (m *FirestartrBootstrap) RenderDeployment(
	ctx context.Context,
) (*dagger.Directory, error) {

	accountID, err := m.ValidateSTSCredentials(ctx)

	if err != nil {
		return nil, fmt.Errorf("obtaining the accountID of aws: %w", err)
	}

	re := regexp.MustCompile("^https://")
	WebhookUri := re.ReplaceAllString(m.Bootstrap.WebhookUrl, "")

	// let's populate the struct
	deploymentData := DeploymentConfig{

		Org:      m.Bootstrap.Org,
		Customer: m.Bootstrap.Customer,
		Webhook: DeploymentWebhook{

			URL:    WebhookUri,
			Secret: m.Bootstrap.WebhookSecretRef,
		},

		ExternalSecrets: DeploymentExternalSecrets{

			RoleARN: fmt.Sprintf("arn:aws:iam::%s:role/FirestartrExternalSecretsStore-%s",

				accountID,

				m.Bootstrap.Customer,
			),
		},

		Controller: DeploymentController{

			Image: fmt.Sprintf("ghcr.io/prefapp/gitops-k8s:%s", fmt.Sprintf(

				"%s_full-%s",
				m.Bootstrap.Firestartr.OperatorVersion,
				m.Creds.CloudProvider.Name,
			)),

			RoleARN: fmt.Sprintf("arn:aws:iam::%s:role/Firestartr-%s",

				accountID,

				m.Bootstrap.Customer,
			),

			GithubApp: DeploymentGithubApp{

				GithubAppId: fmt.Sprintf(

					"/firestartr/%s/fs-%s-admin/app-id",

					m.Bootstrap.Customer,
					m.Bootstrap.Customer,
				),
				GithubAppInstallationId: fmt.Sprintf(

					"/firestartr/%s/fs-%s-admin/%s/app-installation-id",

					m.Bootstrap.Customer,
					m.Bootstrap.Customer,
					m.GhOrgLowerCase,
				),
				GithubAppPem: fmt.Sprintf(

					"/firestartr/%s/fs-%s-admin/pem",

					m.Bootstrap.Customer,
					m.Bootstrap.Customer,
				),
			},
		},

		Aws: DeploymentAws{

			Bucket: *m.Creds.CloudProvider.Config.Bucket,
			Region: m.Creds.CloudProvider.Config.Region,
		},

		Provider: DeploymentGithubApp{

			GithubAppId: fmt.Sprintf(

				"/firestartr/%s/fs-%s-admin/app-id",

				m.Bootstrap.Customer,
				m.Bootstrap.Customer,
			),
			GithubAppInstallationId: fmt.Sprintf(

				"/firestartr/%s/fs-%s-admin/%s/app-installation-id",

				m.Bootstrap.Customer,
				m.Bootstrap.Customer,
				m.GhOrgLowerCase,
			),
			GithubAppPem: fmt.Sprintf(

				"/firestartr/%s/fs-%s-admin/pem",

				m.Bootstrap.Customer,
				m.Bootstrap.Customer,
			),
		},
	}

	deploymentTemplateFile := dag.CurrentModule().
		Source().
		File("templates/deployment/values.tmpl")

	deploymentPreTemplateFile := dag.CurrentModule().
		Source().
		File("templates/deployment/pre.tmpl")

		// deployment values
	templateContent, err := deploymentTemplateFile.Contents(ctx)
	if err != nil {
		return nil, err
	}

	renderedValues, err := renderTmpl(templateContent, deploymentData)
	if err != nil {
		return nil, err
	}

	// deployment master yaml file
	templatePreContent, err := deploymentPreTemplateFile.Contents(ctx)
	if err != nil {
		return nil, err
	}

	renderedPre, err := renderTmpl(templatePreContent, deploymentData)
	if err != nil {
		return nil, err
	}

	deploymentDir := dag.Directory().
		WithNewFile(fmt.Sprintf("%s.yaml", m.Bootstrap.Env), renderedPre).
		WithNewFile(fmt.Sprintf("%s/values.yaml", m.Bootstrap.Env), renderedValues)

	return deploymentDir, nil
}
