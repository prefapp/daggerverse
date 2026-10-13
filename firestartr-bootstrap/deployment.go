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
		fmt.Sprintf("ci: add deployment for %s [automated]", m.Bootstrap.Customer),
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
				m.Creds.CloudProvider.ImageFlavorSuffix(),
			)),

			RoleARN: fmt.Sprintf("arn:aws:iam::%s:role/Firestartr-%s",

				accountID,

				m.Bootstrap.Customer,
			),

			GithubApp: DeploymentGithubApp{

				GithubAppId: fmt.Sprintf(

					"/firestartr/%s/fs-%s/app-id",

					m.Bootstrap.Customer,
					m.Bootstrap.Customer,
				),
				GithubAppInstallationId: fmt.Sprintf(

					"/firestartr/%s/fs-%s/%s/app-installation-id",

					m.Bootstrap.Customer,
					m.Bootstrap.Customer,
					m.GhOrgLowerCase,
				),
				GithubAppPem: fmt.Sprintf(

					"/firestartr/%s/fs-%s/pem",

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
		File("templates/deployment/tenant.tmpl")

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

// CreateDeploymentAzure renders and creates a PR for a dedicated Azure deployment.
// Unlike CreateDeployment, it does not validate STS credentials and targets the
// customer's own state-sys-services repo instead of the shared firestartr-<env>/app-firestartr.
func (m *FirestartrBootstrap) CreateDeploymentAzure(
	ctx context.Context,
) (*dagger.Directory, error) {

	deploymentRenderedDir, err := m.RenderDeploymentAzure(ctx)
	if err != nil {
		return nil, fmt.Errorf("rendering Azure deployment data: %w", err)
	}

	tokenSecret := dag.SetSecret(
		"token",
		m.Creds.GithubApp.OperatorPat,
	)

	err = m.CreatePR(
		ctx,
		"state-sys-services",
		m.Bootstrap.Org,
		deploymentRenderedDir,
		fmt.Sprintf("automated-create-deployment-%s", m.Bootstrap.Customer),
		fmt.Sprintf("ci: add dedicated deployment for %s [automated]", m.Bootstrap.Customer),
		"",
		tokenSecret,
	)

	if err != nil {
		return nil, fmt.Errorf("error generating PR for state-sys-services deployment: %w", err)
	}

	return deploymentRenderedDir, nil
}

// RenderDeploymentAzure renders the full state-sys-services directory structure
// required for a dedicated Azure deployment under firestartr-aks/.
func (m *FirestartrBootstrap) RenderDeploymentAzure(
	ctx context.Context,
) (*dagger.Directory, error) {

	re := regexp.MustCompile("^https://")
	webhookUri := re.ReplaceAllString(m.Bootstrap.WebhookUrl, "")

	azureData := AzureDeploymentConfig{
		Customer:     m.Bootstrap.Customer,
		Org:          m.Bootstrap.Org,
		OrgLowerCase: m.GhOrgLowerCase,
		Domain:       m.Bootstrap.Domain,
		DeploymentPlatform: DeploymentPlatformAKS,
		// ExternalDnsClientId defaults to the main MI here since RenderDeploymentAzure
		// runs at PR-creation time (CmdPushDeployment) when the kind cluster may not
		// be accessible.  Users can re-run or manually update the state-sys-services
		// values once the dedicated external-dns MI client_id is known.
		ExternalDnsClientId: m.Creds.CloudProvider.Config.ClientId,
		Webhook: DeploymentWebhook{
			URL:    webhookUri,
			Secret: m.Bootstrap.WebhookSecretRef,
		},
		CloudProvider: m.Creds.CloudProvider,
		Controller: DeploymentController{
			Image: fmt.Sprintf("ghcr.io/prefapp/gitops-k8s:%s_full-%s",
				m.Bootstrap.Firestartr.OperatorVersion,
				m.Creds.CloudProvider.ImageFlavorSuffix(),
			),
		},
	}

	platform := DeploymentPlatformAKS

	// Render firestartr.yaml (tenant/release descriptor)
	tenantTmpl, err := dag.CurrentModule().Source().File("templates/deployment/azure_tenant.tmpl").Contents(ctx)
	if err != nil {
		return nil, fmt.Errorf("reading azure_tenant.tmpl: %w", err)
	}
	renderedTenant, err := renderTmpl(tenantTmpl, azureData)
	if err != nil {
		return nil, fmt.Errorf("rendering azure_tenant.tmpl: %w", err)
	}

	// Render firestartr/values.yaml
	valuesTmpl, err := dag.CurrentModule().Source().File("templates/deployment/azure_values.tmpl").Contents(ctx)
	if err != nil {
		return nil, fmt.Errorf("reading azure_values.tmpl: %w", err)
	}
	renderedValues, err := renderTmpl(valuesTmpl, azureData)
	if err != nil {
		return nil, fmt.Errorf("rendering azure_values.tmpl: %w", err)
	}

	// Render nginx release descriptor (static — no template variables needed)
	nginxTmpl, err := dag.CurrentModule().Source().File("templates/deployment/sys_services/nginx.tmpl").Contents(ctx)
	if err != nil {
		return nil, fmt.Errorf("reading nginx.tmpl: %w", err)
	}

	// Render nginx/values.yaml (static)
	nginxValuesTmpl, err := dag.CurrentModule().Source().File("templates/deployment/sys_services/nginx_values.tmpl").Contents(ctx)
	if err != nil {
		return nil, fmt.Errorf("reading nginx_values.tmpl: %w", err)
	}

	// Render cert-manager release descriptor (static)
	certManagerTmpl, err := dag.CurrentModule().Source().File("templates/deployment/sys_services/cert_manager.tmpl").Contents(ctx)
	if err != nil {
		return nil, fmt.Errorf("reading cert_manager.tmpl: %w", err)
	}

	// Render cert-manager/values.yaml (static)
	certManagerValuesTmpl, err := dag.CurrentModule().Source().File("templates/deployment/sys_services/cert_manager_values.tmpl").Contents(ctx)
	if err != nil {
		return nil, fmt.Errorf("reading cert_manager_values.tmpl: %w", err)
	}

	// Render external-dns release descriptor (static)
	externalDnsTmpl, err := dag.CurrentModule().Source().File("templates/deployment/sys_services/external_dns.tmpl").Contents(ctx)
	if err != nil {
		return nil, fmt.Errorf("reading external_dns.tmpl: %w", err)
	}

	// Render external-dns/values.yaml (needs Azure identity config)
	externalDnsValuesTmpl, err := dag.CurrentModule().Source().File("templates/deployment/sys_services/external_dns_values.tmpl").Contents(ctx)
	if err != nil {
		return nil, fmt.Errorf("reading external_dns_values.tmpl: %w", err)
	}
	renderedExternalDnsValues, err := renderTmpl(externalDnsValuesTmpl, azureData)
	if err != nil {
		return nil, fmt.Errorf("rendering external_dns_values.tmpl: %w", err)
	}

	// Render ArgoCD release descriptor (static)
	argocdTmpl, err := dag.CurrentModule().Source().File("templates/deployment/sys_services/argocd.tmpl").Contents(ctx)
	if err != nil {
		return nil, fmt.Errorf("reading argocd.tmpl: %w", err)
	}

	// Render ArgoCD/values.yaml (needs domain)
	argocdValuesTmpl, err := dag.CurrentModule().Source().File("templates/deployment/sys_services/argocd_values.tmpl").Contents(ctx)
	if err != nil {
		return nil, fmt.Errorf("reading argocd_values.tmpl: %w", err)
	}
	renderedArgocdValues, err := renderTmpl(argocdValuesTmpl, azureData)
	if err != nil {
		return nil, fmt.Errorf("rendering argocd_values.tmpl: %w", err)
	}

	// Render argo-configuration-secrets release descriptor (static)
	argoConfigSecretsTmpl, err := dag.CurrentModule().Source().File("templates/deployment/sys_services/argo_config_secrets.tmpl").Contents(ctx)
	if err != nil {
		return nil, fmt.Errorf("reading argo_config_secrets.tmpl: %w", err)
	}

	// Render argo-configuration-secrets/values.yaml
	argoConfigSecretsValuesTmpl, err := dag.CurrentModule().Source().File("templates/deployment/sys_services/argo_config_secrets_values.tmpl").Contents(ctx)
	if err != nil {
		return nil, fmt.Errorf("reading argo_config_secrets_values.tmpl: %w", err)
	}
	renderedArgoConfigSecretsValues, err := renderTmpl(argoConfigSecretsValuesTmpl, azureData)
	if err != nil {
		return nil, fmt.Errorf("rendering argo_config_secrets_values.tmpl: %w", err)
	}

	// Render argo-events release descriptor (static)
	argoEventsTmpl, err := dag.CurrentModule().Source().File("templates/deployment/sys_services/argo_events.tmpl").Contents(ctx)
	if err != nil {
		return nil, fmt.Errorf("reading argo_events.tmpl: %w", err)
	}

	// Render argo-events/values.yaml (static)
	argoEventsValuesTmpl, err := dag.CurrentModule().Source().File("templates/deployment/sys_services/argo_events_values.tmpl").Contents(ctx)
	if err != nil {
		return nil, fmt.Errorf("reading argo_events_values.tmpl: %w", err)
	}

	// Render argo-workflows release descriptor (static)
	argoWorkflowsTmpl, err := dag.CurrentModule().Source().File("templates/deployment/sys_services/argo_workflows.tmpl").Contents(ctx)
	if err != nil {
		return nil, fmt.Errorf("reading argo_workflows.tmpl: %w", err)
	}

	// Render argo-workflows/values.yaml (static)
	argoWorkflowsValuesTmpl, err := dag.CurrentModule().Source().File("templates/deployment/sys_services/argo_workflows_values.tmpl").Contents(ctx)
	if err != nil {
		return nil, fmt.Errorf("reading argo_workflows_values.tmpl: %w", err)
	}

	basePath := fmt.Sprintf("kubernetes-sys-services/%s", platform)
	deploymentDir := dag.Directory().
		WithNewFile(fmt.Sprintf("%s/firestartr.yaml", basePath), renderedTenant).
		WithNewFile(fmt.Sprintf("%s/firestartr/values.yaml", basePath), renderedValues).
		WithNewFile(fmt.Sprintf("%s/nginx.yaml", basePath), nginxTmpl).
		WithNewFile(fmt.Sprintf("%s/nginx/values.yaml", basePath), nginxValuesTmpl).
		WithNewFile(fmt.Sprintf("%s/cert-manager.yaml", basePath), certManagerTmpl).
		WithNewFile(fmt.Sprintf("%s/cert-manager/values.yaml", basePath), certManagerValuesTmpl).
		WithNewFile(fmt.Sprintf("%s/external-dns.yaml", basePath), externalDnsTmpl).
		WithNewFile(fmt.Sprintf("%s/external-dns/values.yaml", basePath), renderedExternalDnsValues).
		WithNewFile(fmt.Sprintf("%s/argocd.yaml", basePath), argocdTmpl).
		WithNewFile(fmt.Sprintf("%s/argocd/values.yaml", basePath), renderedArgocdValues).
		WithNewFile(fmt.Sprintf("%s/argo-configuration-secrets.yaml", basePath), argoConfigSecretsTmpl).
		WithNewFile(fmt.Sprintf("%s/argo-configuration-secrets/values.yaml", basePath), renderedArgoConfigSecretsValues).
		WithNewFile(fmt.Sprintf("%s/argo-events.yaml", basePath), argoEventsTmpl).
		WithNewFile(fmt.Sprintf("%s/argo-events/values.yaml", basePath), argoEventsValuesTmpl).
		WithNewFile(fmt.Sprintf("%s/argo-workflows.yaml", basePath), argoWorkflowsTmpl).
		WithNewFile(fmt.Sprintf("%s/argo-workflows/values.yaml", basePath), argoWorkflowsValuesTmpl)

	return deploymentDir, nil
}
