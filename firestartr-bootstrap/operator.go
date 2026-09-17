package main

import (
	"context"
	"dagger/firestartr-bootstrap/internal/dagger"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"
)

// Chart versions for dedicated sys-services — must stay in sync with the
// release descriptor templates under templates/deployment/sys_services/.
const (
	chartVersionNginx         = "4.10.1"
	chartVersionCertManager   = "v1.15.0"
	chartVersionExternalDns   = "1.14.4"
	chartVersionArgoCD        = "7.6.8"
	chartVersionArgoEvents    = "2.4.10"
	chartVersionArgoWorkflows = "0.42.5"
)

func sysServicesApplyScript() string {
	return "set -euo pipefail\nfor ns in external-secrets ingress-nginx cert-manager external-dns firestartr argocd argo-events argo-workflows; do\n  kubectl create namespace \"$ns\" --dry-run=client -o yaml | kubectl apply -f -\ndone\nhelm template external-secrets external-secrets/external-secrets -n external-secrets --include-crds | kubectl apply --server-side --force-conflicts --field-manager=firestartr-bootstrap -f -\ncurl -fsSL https://raw.githubusercontent.com/firestartr-pro/docs/refs/heads/main/site/raw/core/crds/v2.6.4/index.yaml | kubectl apply --server-side --force-conflicts --field-manager=firestartr-bootstrap -f -\nhelm template ingress-nginx ingress-nginx/ingress-nginx -n ingress-nginx --version " + chartVersionNginx + " --include-crds --values /sys-values/nginx-values.yaml | kubectl apply --server-side --force-conflicts --field-manager=firestartr-bootstrap -f -\nhelm template cert-manager jetstack/cert-manager -n cert-manager --version " + chartVersionCertManager + " --include-crds --values /sys-values/cert-manager-values.yaml | kubectl apply --server-side --force-conflicts --field-manager=firestartr-bootstrap -f -\nhelm template external-dns external-dns/external-dns -n external-dns --version " + chartVersionExternalDns + " --include-crds --values /sys-values/external-dns-values.yaml | kubectl apply --server-side --force-conflicts --field-manager=firestartr-bootstrap -f -\nhelm repo add firestartr-controller https://prefapp.github.io/charts/firestartr-controller\nhelm template firestartr firestartr-controller/firestartr -n firestartr --version 3.5.0 --include-crds --values /sys-values/firestartr-values.yaml | kubectl apply -n firestartr --server-side --force-conflicts --field-manager=firestartr-bootstrap -f -\nhelm template argocd argo/argo-cd -n argocd --version " + chartVersionArgoCD + " --include-crds --values /sys-values/argocd-values.yaml | kubectl apply --server-side --force-conflicts --field-manager=firestartr-bootstrap -f -\nhelm repo add argo-configuration-secrets https://prefapp.github.io/charts/argo-configuration-secrets\nhelm template argo-config-secrets argo-configuration-secrets/argocd-configuration-secrets -n argocd --version 1.1.0 --include-crds --values /sys-values/argo-configuration-secrets-values.yaml | kubectl apply -n argocd --server-side --force-conflicts --field-manager=firestartr-bootstrap -f -\nhelm template argo-events argo/argo-events -n argo-events --version " + chartVersionArgoEvents + " --include-crds --values /sys-values/argo-events-values.yaml | kubectl apply --server-side --force-conflicts --field-manager=firestartr-bootstrap -f -\nhelm template argo-workflows argo/argo-workflows -n argo-workflows --version " + chartVersionArgoWorkflows + " --include-crds --values /sys-values/argo-workflows-values.yaml | kubectl apply --server-side --force-conflicts --field-manager=firestartr-bootstrap -f -"
}

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

	// For dedicated deployments, wait for the TFWorkspace CRs that were
	// already applied by InstallInitialCRsAndBuildHelmValues (they live in
	// /resources/initial-crs).  We apply them again idempotently so that
	// applyCrAndWaitForProvisioned can block until PROVISIONED=True.
	if m.isDedicatedDeployment() {
		kindContainer, err = m.ApplyFirestartrCrs(
			ctx,
			kindContainer,
			"/resources/initial-crs",
			[]string{"FirestartrTerraformWorkspace.*"},
		)
		if err != nil {
			return nil, fmt.Errorf("waiting for dedicated TFWorkspace CRs: %w", err)
		}
	}

	kindContainer, err = m.ApplyFirestartrCrs(
		ctx,
		kindContainer,
		"/resources/firestartr-crs/infra",
		[]string{"ExternalSecret.*"},
	)
	if err != nil {
		return nil, err
	}

	crList := []string{
		"FirestartrGithubGroup.*",
		"FirestartrGithubRepository.*",
		"FirestartrGithubRepositorySecretsSection.*",
		"FirestartrGithubRepositoryFeature.*",
	}
	if m.CreateWebhook {
		crList = append(crList, "FirestartrGithubOrgWebhook.*")
	}

	kindContainer, err = m.ApplyFirestartrCrs(
		ctx,
		kindContainer,
		"/resources/firestartr-crs/github",
		crList,
	)
	if err != nil {
		return nil, err
	}

	return kindContainer, nil

}

func (m *FirestartrBootstrap) InstallClusterServices(
	ctx context.Context,
	kindContainer *dagger.Container,
) (*dagger.Container, error) {

	// External Secrets Operator — required for both SaaS and dedicated paths
	result, err := kindContainer.
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

	// For dedicated deployments the remaining cluster services (nginx, cert-manager,
	// external-dns, ArgoCD, argo-events, argo-workflows) are installed on the
	// target AKS cluster via ApplySysServicesWithValues — NOT on the local kind
	// cluster. Kind only needs ESO to resolve ExternalSecrets during bootstrap.
	return result, nil
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
	var mu sync.Mutex

	for _, kind := range crsToApplyList {
		g, egCtx := errgroup.WithContext(ctx)

		entries, err := kindContainer.Directory(crsDirectoryPath).Glob(egCtx, kind)
		if err != nil {
			return nil, fmt.Errorf("Failed to get glob entries: %s", err)
		}

		for _, entry := range entries {
			entry := entry
			err := err
			g.Go(func() error {
				err = m.applyCrAndWaitForProvisioned(
					egCtx, kindContainer,
					fmt.Sprintf("%s/%s", crsDirectoryPath, entry),
					kind != "ExternalSecret.*",
					&mu,
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

func (m *FirestartrBootstrap) applyCrAndWaitForProvisioned(
	ctx context.Context,
	kindContainer *dagger.Container,
	entry string,
	waitForProvisioned bool,
	mu *sync.Mutex,
) error {
	crFile := kindContainer.File(entry)

	crContent, err := crFile.Contents(ctx)
	if err != nil {
		return fmt.Errorf("Failed to get file contents: %s", err)
	}

	cr := &Cr{}
	err = yaml.Unmarshal([]byte(crContent), cr)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal CR: %s", err)
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
			return err
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

	mu.Lock()
	if err != nil {
		m.FailedCrs = append(m.FailedCrs, cr)
	} else {
		m.ProvisionedCrs = append(m.ProvisionedCrs, cr)
	}
	mu.Unlock()

	return nil
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

// ApplySysServicesWithValues renders the Azure-specific Helm values for every
// dedicated sys-service and applies them to the target AKS cluster via
// `helm upgrade --install --values`.  This fills the gap between
// InstallClusterServices (bare Helm install, no custom values) and the
// state-sys-services PR that ArgoCD will manage going forward.
//
// The AKS kubeconfig is fetched automatically via `az aks get-credentials`
// using the bootstrap App Registration credentials stored in the credentials
// file, so no pre-built kubeconfig directory is required.
//
// The AKS cluster name and resource group are read from
// ConfigProvider.AksClusterName and ConfigProvider.ResourceGroupName.
//
// externalDnsClientId is the client ID of the dedicated external-dns Managed
// Identity provisioned by the TFWorkspace CR.  It must be non-empty; callers
// should obtain it via GetExternalDnsMIClientId before calling this function.
func (m *FirestartrBootstrap) ApplySysServicesWithValues(
	ctx context.Context,
	// Client ID of the dedicated external-dns Managed Identity (from GetExternalDnsMIClientId).
	externalDnsClientId string,
) (*dagger.Container, error) {

	if !m.isDedicatedDeployment() {
		return nil, fmt.Errorf("ApplySysServicesWithValues is only applicable for dedicated deployments")
	}

	cfg := m.Creds.CloudProvider.Config

	re := regexp.MustCompile("^https://")
	webhookUri := re.ReplaceAllString(m.Bootstrap.WebhookUrl, "")

	azureData := AzureDeploymentConfig{
		Customer:            m.Bootstrap.Customer,
		Org:                 m.Bootstrap.Org,
		OrgLowerCase:        m.GhOrgLowerCase,
		Domain:              m.Bootstrap.Domain,
		DeploymentPlatform:  DeploymentPlatformAKS,
		ExternalDnsClientId: externalDnsClientId,
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

	// Render the values file for each service that has dynamic content.
	// Static services (nginx, cert-manager) still get their values applied
	// so the cluster state matches the state-sys-services repo exactly.

	nginxValuesTmpl, err := dag.CurrentModule().Source().
		File("templates/deployment/sys_services/nginx_values.tmpl").Contents(ctx)
	if err != nil {
		return nil, fmt.Errorf("reading nginx_values.tmpl: %w", err)
	}

	certManagerValuesTmpl, err := dag.CurrentModule().Source().
		File("templates/deployment/sys_services/cert_manager_values.tmpl").Contents(ctx)
	if err != nil {
		return nil, fmt.Errorf("reading cert_manager_values.tmpl: %w", err)
	}

	externalDnsValuesTmpl, err := dag.CurrentModule().Source().
		File("templates/deployment/sys_services/external_dns_values.tmpl").Contents(ctx)
	if err != nil {
		return nil, fmt.Errorf("reading external_dns_values.tmpl: %w", err)
	}
	renderedExternalDnsValues, err := renderTmpl(externalDnsValuesTmpl, azureData)
	if err != nil {
		return nil, fmt.Errorf("rendering external_dns_values.tmpl: %w", err)
	}

	argoConfigSecretsValuesTmpl, err := dag.CurrentModule().Source().File("templates/deployment/sys_services/argo_config_secrets_values.tmpl").Contents(ctx)
	if err != nil {
		return nil, fmt.Errorf("reading argo_config_secrets_values.tmpl: %w", err)
	}
	renderedArgoConfigSecretsValues, err := renderTmpl(argoConfigSecretsValuesTmpl, azureData)
	if err != nil {
		return nil, fmt.Errorf("rendering argo_config_secrets_values.tmpl: %w", err)
	}

	firestartrValuesTmpl, err := dag.CurrentModule().Source().File("templates/deployment/azure_values.tmpl").Contents(ctx)
	if err != nil {
		return nil, fmt.Errorf("reading azure_values.tmpl: %w", err)
	}
	renderedValues, err := renderTmpl(firestartrValuesTmpl, azureData)
	if err != nil {
		return nil, fmt.Errorf("rendering azure_values.tmpl: %w", err)
	}

	argocdValuesTmpl, err := dag.CurrentModule().Source().
		File("templates/deployment/sys_services/argocd_values.tmpl").Contents(ctx)
	if err != nil {
		return nil, fmt.Errorf("reading argocd_values.tmpl: %w", err)
	}
	renderedArgocdValues, err := renderTmpl(argocdValuesTmpl, azureData)
	if err != nil {
		return nil, fmt.Errorf("rendering argocd_values.tmpl: %w", err)
	}

	// argo-events — static values (no Azure-specific interpolation needed)
	argoEventsValuesTmpl, err := dag.CurrentModule().Source().
		File("templates/deployment/sys_services/argo_events_values.tmpl").Contents(ctx)
	if err != nil {
		return nil, fmt.Errorf("reading argo_events_values.tmpl: %w", err)
	}

	// argo-workflows — static values
	argoWorkflowsValuesTmpl, err := dag.CurrentModule().Source().
		File("templates/deployment/sys_services/argo_workflows_values.tmpl").Contents(ctx)
	if err != nil {
		return nil, fmt.Errorf("reading argo_workflows_values.tmpl: %w", err)
	}

	// Fetch the AKS admin kubeconfig in an Azure CLI container, then mount it
	// into the render/apply container that applies raw manifests.
	clientSecret := dag.SetSecret("azure-bootstrap-client-secret", cfg.BootstrapClientSecret)

	aksCtr, err := dag.Container().
		From("mcr.microsoft.com/azure-cli:latest").
		WithSecretVariable("AZURE_CLIENT_SECRET", clientSecret).
		// Login with bootstrap Service Principal (password read from env var via shell
		// so the secret value is not logged in Dagger's exec args).
		WithExec([]string{
			"sh", "-c",
			fmt.Sprintf(
				"az login --service-principal --username %s --password $AZURE_CLIENT_SECRET --tenant %s",
				cfg.BootstrapClientId,
				cfg.TenantId,
			),
		}).
		WithExec([]string{
			"az", "account", "set",
			"--subscription", cfg.SubscriptionId,
		}).
		// Fetch AKS admin kubeconfig — no external kubeconfig directory required
		WithExec([]string{
			"az", "aks", "get-credentials",
			"--resource-group", cfg.ResourceGroupName,
			"--name", cfg.AksClusterName,
			"--overwrite-existing",
			"--admin",
		}).
		Sync(ctx)

	if err != nil {
		errMsg := extractErrorMessage(err, "Failed to fetch AKS kubeconfig")
		return nil, errors.New(errMsg)
	}

	kubeconfigDir := aksCtr.Directory("/root/.kube")
	valuesDir := dag.Directory().
		WithNewFile("firestartr-values.yaml", renderedValues).
		WithNewFile("nginx-values.yaml", nginxValuesTmpl).
		WithNewFile("cert-manager-values.yaml", certManagerValuesTmpl).
		WithNewFile("external-dns-values.yaml", renderedExternalDnsValues).
		WithNewFile("argocd-values.yaml", renderedArgocdValues).
		WithNewFile("argo-configuration-secrets-values.yaml", renderedArgoConfigSecretsValues).
		WithNewFile("argo-events-values.yaml", argoEventsValuesTmpl).
		WithNewFile("argo-workflows-values.yaml", argoWorkflowsValuesTmpl)

	ctr, err := dag.Container().
		From("ghcr.io/helmfile/helmfile:latest").
		WithEnvVariable("BUST_CACHE", time.Now().String()).
		WithMountedDirectory("/root/.kube", kubeconfigDir).
		WithEnvVariable("KUBECONFIG", "/root/.kube/config").
		WithDirectory("/sys-values", valuesDir).
		WithExec([]string{"helm", "repo", "add", "external-secrets", "https://charts.external-secrets.io"}).
		WithExec([]string{"helm", "repo", "add", "ingress-nginx", "https://kubernetes.github.io/ingress-nginx"}).
		WithExec([]string{"helm", "repo", "add", "jetstack", "https://charts.jetstack.io"}).
		WithExec([]string{"helm", "repo", "add", "external-dns", "https://kubernetes-sigs.github.io/external-dns/"}).
		WithExec([]string{"helm", "repo", "add", "argo", "https://argoproj.github.io/argo-helm"}).
		WithExec([]string{"helm", "repo", "add", "prefapp", "https://prefapp.github.io/charts"}).
		WithExec([]string{"helm", "repo", "update"}).
		WithExec([]string{"sh", "-c", sysServicesApplyScript()}).
		Sync(ctx)

	if err != nil {
		errMsg := extractErrorMessage(err, "Failed to apply sys-services with values")
		return nil, errors.New(errMsg)
	}

	return ctr, nil
}
