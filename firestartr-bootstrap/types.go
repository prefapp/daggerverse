package main

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
	FinalSecretStoreName          string      `yaml:"finalSecretStoreName"`
	WebhookUrl                    string      // Autocalculated
	WebhookSecretRef              string      // Autocalculated
	PrefappBotPatSecretRef        string      // Autocalculated
	FirestartrCliVersionSecretRef string      // Autocalculated
	HasFreePlan                   bool        // Autocalculated
}

type PushFiles struct {
	Claims PushFilesRepo `yaml:"claims"`
	Crs    Crs           `yaml:"crs"`
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
}

type Firestartr struct {
	OperatorVersion string `yaml:"operator"`
	CliVersion      string `yaml:"cli"`
}

type CredsFile struct {
	CloudProvider       CloudProvider `yaml:"cloudProvider"`
	GithubApp           GithubApp     `yaml:"github"`
    GithubAppOperator   GithubApp     //Autocalculated
}

type CloudProvider struct {
	ProviderConfigName string
	Config             ConfigProvider `yaml:"config"`
	Source             string         `yaml:"source"`
	Type               string         `yaml:"type"`
	Version            string         `yaml:"version"`
	Name               string         `yaml:"name"`
}

type ConfigProvider struct {
	Bucket    *string `json:"bucket" yaml:"bucket"`
	Region    string  `json:"region" yaml:"region"`
	AccessKey string  `json:"access_key" yaml:"access_key"`
	SecretKey string  `json:"secret_key" yaml:"secret_key"`
	Token     string  `json:"token" yaml:"token"`
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
