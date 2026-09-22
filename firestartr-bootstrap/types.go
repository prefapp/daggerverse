package main

import "encoding/json"

// ToConfigJSON serialises ConfigProvider to JSON with omitempty semantics,
// emitting only fields that have non-zero values.
//
// We cannot rely on sprig's toJson (json.Marshal) honouring the omitempty
// tags in types.go because dagger develop generates a ConfigProvider.MarshalJSON
// method in dagger.gen.go whose intermediate struct lacks omitempty. That
// generated method takes precedence over struct tags and always serialises every
// field. ToConfigJSON bypasses it by marshalling a local mirror type whose tags
// include omitempty.
func (c ConfigProvider) ToConfigJSON() string {
	type mirror struct {
		Bucket                *string `json:"bucket,omitempty"`
		Region                string  `json:"region,omitempty"`
		AccessKey             string  `json:"access_key,omitempty"`
		SecretKey             string  `json:"secret_key,omitempty"`
		Token                 string  `json:"token,omitempty"`
		TenantId              string  `json:"tenant_id,omitempty"`
		SubscriptionId        string  `json:"subscription_id,omitempty"`
		ClientId              string  `json:"client_id,omitempty"`
		StorageAccountName    string  `json:"storage_account_name,omitempty"`
		ContainerName         string  `json:"container_name,omitempty"`
		ResourceGroupName     string  `json:"resource_group_name,omitempty"`
		KeyVaultName          string  `json:"key_vault_name,omitempty"`
		BootstrapClientId     string  `json:"bootstrap_client_id,omitempty"`
		BootstrapClientSecret string  `json:"bootstrap_client_secret,omitempty"`
		ClientSecret          string  `json:"client_secret,omitempty"`
	}
	b, _ := json.Marshal(mirror{
		Bucket:                c.Bucket,
		Region:                c.Region,
		AccessKey:             c.AccessKey,
		SecretKey:             c.SecretKey,
		Token:                 c.Token,
		TenantId:              c.TenantId,
		SubscriptionId:        c.SubscriptionId,
		ClientId:              c.ClientId,
		StorageAccountName:    c.StorageAccountName,
		ContainerName:         c.ContainerName,
		ResourceGroupName:     c.ResourceGroupName,
		KeyVaultName:          c.KeyVaultName,
		BootstrapClientId:     c.BootstrapClientId,
		BootstrapClientSecret: c.BootstrapClientSecret,
		ClientSecret:          c.ClientSecret,
	})
	return string(b)
}

// ToAzureProviderConfigJSON serialises only the fields that are valid for the
// Terraform azurerm *provider* block (tenant_id, subscription_id, client_id,
// client_secret).  Backend-specific fields (storage_account_name,
// container_name, resource_group_name) are intentionally excluded because the
// Terraform azurerm provider rejects them as "extraneous JSON object
// properties".
//
// Use ToConfigJSON when you need the full config for the backend
// FirestartrProviderConfig, and this method when you need the provider-only
// FirestartrProviderConfig.
func (c ConfigProvider) ToAzureProviderConfigJSON() string {
	type mirror struct {
		TenantId       string                 `json:"tenant_id,omitempty"`
		SubscriptionId string                 `json:"subscription_id,omitempty"`
		ClientId       string                 `json:"client_id,omitempty"`
		ClientSecret   string                 `json:"client_secret,omitempty"`
		Features       map[string]interface{} `json:"features"`
	}
	b, _ := json.Marshal(mirror{
		TenantId:       c.TenantId,
		SubscriptionId: c.SubscriptionId,
		ClientId:       c.ClientId,
		ClientSecret:   c.ClientSecret,
		Features:       map[string]interface{}{},
	})
	return string(b)
}

type Component struct {
	Name          string `yaml:"name"`
	RepoName      string `yaml:"repoName"`
	Description   string `yaml:"description"`
	DefaultBranch string `yaml:"defaultBranch"`
	Features      []Feature
	Variables     []Variable `yaml:"variables"`
	Secrets       []Variable `yaml:"secrets"` // Secrets have the same structure as Variables
	Labels        []string   `yaml:"labels"`
	Skipped       bool       `yaml:"skip"`
}

type Variable struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type Feature struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

// DeploymentModeSaaS and DeploymentModeDedicated are the two supported deployment modes.
const DeploymentModeSaaS = "saas"
const DeploymentModeDedicated = "dedicated"

// DeploymentPlatformAKS is the hardcoded platform name for dedicated Azure deployments.
const DeploymentPlatformAKS = "firestartr-aks"

type Bootstrap struct {
	Env                           string      `yaml:"env"`
	Firestartr                    Firestartr  `yaml:"firestartr"`
	PushFiles                     PushFiles   `yaml:"pushFiles"`
	Org                           string      `yaml:"org"`
	Customer                      string      `yaml:"customer"`
	Components                    []Component `yaml:"components"`
	DefaultSystemName             string      `yaml:"defaultSystemName"`
	DefaultDomainName             string      `yaml:"defaultDomainName"`
	DefaultFirestartrGroup        string      `yaml:"defaultFirestartrGroup"`
	DefaultBranch                 string      `yaml:"defaultBranch"`
	DefaultBranchStrategy         string      `yaml:"defaultBranchStrategy"`
	DefaultOrgPermissions         string      `yaml:"defaultOrgPermissions"`
	DefaultGroup                  string      `yaml:"defaultGroup"`
	CreateWebhook                 bool        `yaml:"createWebhook"`
	FinalSecretStoreName          string      `yaml:"finalSecretStoreName"`
	DeploymentMode                string      `yaml:"deploymentMode"` // "saas" (default) or "dedicated"
	Domain                        string      `yaml:"domain"`         // Base domain for dedicated deployments
	WebhookUrl                    string      // Autocalculated
	WebhookSecretRef              string      // Autocalculated
	PrefappBotPatSecretRef        string      // Autocalculated
	FirestartrCliVersionSecretRef string      // Autocalculated
	HasFreePlan                   bool        // Autocalculated
}

// isDedicatedDeployment returns true when the bootstrap is configured for a
// dedicated (non-SaaS, currently Azure) deployment.
func (b *Bootstrap) isDedicatedDeployment() bool {
	return b.DeploymentMode == DeploymentModeDedicated
}

type PushFiles struct {
	Claims        PushFilesRepo `yaml:"claims"`
	Crs           Crs           `yaml:"crs"`
	DotFirestartr PushFilesRepo `yaml:"dotFirestartr"`
}

type Crs struct {
	Providers Providers `yaml:"providers"`
}

type PushFilesRepo struct {
	Push bool   `yaml:"push"`
	Repo string `yaml:"repo"`
}

type Providers struct {
	Github    PushFilesRepo `yaml:"github"`
	Terraform PushFilesRepo `yaml:"terraform"`
	Secrets   PushFilesRepo `yaml:"secrets"`
}

type Firestartr struct {
	OperatorVersion string `yaml:"operator"`
	CliVersion      string `yaml:"cli"`
}

type CredsFile struct {
	CloudProvider     CloudProvider `yaml:"cloudProvider"`
	GithubApp         GithubApp     `yaml:"github"`
	GithubAppOperator GithubApp     //Autocalculated
}

type CloudProvider struct {
	ProviderConfigName string
	Config             ConfigProvider `yaml:"config"`
	Source             string         `yaml:"source"`
	Type               string         `yaml:"type"`
	Version            string         `yaml:"version"`
	Name               string         `yaml:"name"`
}

// ImageFlavorSuffix returns the suffix used in the gitops-k8s image tag for this
// cloud provider. Azure is published as "az", not "azure".
func (cp CloudProvider) ImageFlavorSuffix() string {
	if cp.Name == "azure" {
		return "az"
	}
	return cp.Name
}

type ConfigProvider struct {
	// AWS fields
	Bucket    *string `json:"bucket,omitempty" yaml:"bucket"`
	Region    string  `json:"region,omitempty" yaml:"region"`
	AccessKey string  `json:"access_key,omitempty" yaml:"access_key"`
	SecretKey string  `json:"secret_key,omitempty" yaml:"secret_key"`
	Token     string  `json:"token,omitempty" yaml:"token"`

	// Azure fields
	// ClientId is the client ID of the firestartr-mi User-Assigned Managed Identity.
	// Used exclusively in the deployed AKS state via Workload Identity (no secret).
	TenantId           string `json:"tenant_id,omitempty" yaml:"tenant_id"`
	SubscriptionId     string `json:"subscription_id,omitempty" yaml:"subscription_id"`
	ClientId           string `json:"client_id,omitempty" yaml:"client_id"`
	StorageAccountName string `json:"storage_account_name,omitempty" yaml:"storage_account_name"`
	ContainerName      string `json:"container_name,omitempty" yaml:"container_name"`
	ResourceGroupName  string `json:"resource_group_name,omitempty" yaml:"resource_group_name"`
	KeyVaultName       string `json:"key_vault_name,omitempty" yaml:"key_vault_name"`
	// Location is the Azure region for the resource group (e.g. "westeurope").
	// Used when provisioning Azure resources (Managed Identities) via TFWorkspace.
	Location string `json:"location,omitempty" yaml:"location"`
	// AksOidcIssuerUrl is the OIDC issuer URL of the target AKS cluster.
	// Required for configuring Workload Identity federated credentials on
	// dedicated Managed Identities (e.g. the external-dns MI).
	AksOidcIssuerUrl string `json:"aks_oidc_issuer_url,omitempty" yaml:"aks_oidc_issuer_url"`
	// AksClusterName is the name of the target AKS cluster.
	// Used in `az aks get-credentials` during ApplySysServicesWithValues.
	AksClusterName string `json:"aks_cluster_name,omitempty" yaml:"aks_cluster_name"`
	// Bootstrap identity fields — dedicated App Registration (Service Principal).
	// Used only during bootstrap in the local kind cluster by ESO and Terraform.
	// The entire App Registration must be deleted after bootstrap completes.
	BootstrapClientId     string `json:"bootstrap_client_id,omitempty" yaml:"bootstrap_client_id"`
	BootstrapClientSecret string `json:"bootstrap_client_secret,omitempty" yaml:"bootstrap_client_secret"`

	// ClientSecret is an internal computed field populated at render time for dedicated
	// deployments. It is set to BootstrapClientSecret before the FirestartrProviderConfig
	// is serialised so the Terraform azurerm provider in the kind cluster receives the
	// bootstrap SP credentials under the standard client_secret JSON key.
	// It is never read from the credentials file.
	ClientSecret string `json:"client_secret,omitempty" yaml:"-"`
}

type GithubApp struct {
	ProviderConfigName string
	Owner              string // Populated
	PrefappBotPat      string `yaml:"prefappBotPat"`
	OperatorPat        string `yaml:"operatorPat"`
	Pem                string
	RawPem             string
	GhAppId            string
	InstallationId     string
}

type SecretData struct {
	Name  string
	Value string
}

type CrsDefaultsData struct {
	DefaultBranch                   string
	CloudProviderProviderConfigName string
	GithubAppProviderConfigName     string
}

// DeploymentWebhook represents the block containing the URL and Secret of the Webhook.
type DeploymentWebhook struct {
	URL    string
	Secret string
}

// DeploymentExternalSecrets represents the ARN reference of the role for External Secrets.
type DeploymentExternalSecrets struct {
	RoleARN string
}

// DeploymentController represents the GitHub application information used by the controller.
type DeploymentController struct {
	Image     string
	RoleARN   string
	GithubApp DeploymentGithubApp
}

// DeploymentAws represents the specific AWS configuration (Bucket and Region).
type DeploymentAws struct {
	Bucket string
	Region string
}

type DeploymentGithubApp struct {
	GithubAppId             string
	GithubAppPem            string
	GithubAppInstallationId string
}

// DeploymentConfig contains only the top-level fields that are interpolatable.
type DeploymentConfig struct {
	Customer        string
	Org             string
	Webhook         DeploymentWebhook
	ExternalSecrets DeploymentExternalSecrets
	Controller      DeploymentController
	Aws             DeploymentAws
	Provider        DeploymentGithubApp
}

type PushSecretElement struct {
	Name                string
	SecretStore         string
	KubernetesSecret    string
	KubernetesSecretKey string
	ParameterName       string
	Value               string
}

type ArgoCDConfig struct {
	Name      string
	App       string
	Repo      string
	Namespace string
}

// AzureDeploymentConfig holds the data used to render Azure-specific deployment
// templates (azure_values.tmpl, azure_tenant.tmpl, sys-service descriptors).
type AzureDeploymentConfig struct {
	Customer           string
	Org                string
	OrgLowerCase       string
	Domain             string
	DeploymentPlatform string // Always "firestartr-aks" for dedicated
	Webhook            DeploymentWebhook
	CloudProvider      CloudProvider
	Controller         DeploymentController
	// ExternalDnsClientId is the client ID of the dedicated external-dns Managed
	// Identity.  It is read from the kind cluster after the TFWorkspace CR has
	// been provisioned.  If empty, the main firestartr-mi ClientId is used.
	ExternalDnsClientId string
}

// InitialCrsData is the combined data struct passed to initial_crs.tmpl.
// It merges the relevant CredsFile fields (keeping the same template variable
// names) with Bootstrap fields that are needed for dedicated-mode CRs such as
// TFWorkspace claims.
type InitialCrsData struct {
	// Preserved from CredsFile so existing template variables still work.
	CloudProvider     CloudProvider
	GithubApp         GithubApp
	GithubAppOperator GithubApp
	// Bootstrap fields available in dedicated-mode template blocks.
	Customer       string
	Org            string
	DeploymentMode string
	Domain         string
}
