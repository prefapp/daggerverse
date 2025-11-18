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
	Firestartr                      Firestartr  `yaml:"firestartr"`
	PushFiles                       PushFiles   `yaml:"pushFiles"`
	Org                             string      `yaml:"org"`
    Customer                        string      `yaml:"customer"`
	Components                      []Component `yaml:"components"`
	DefaultSystemName               string      `yaml:"defaultSystemName"`
	DefaultDomainName               string      `yaml:"defaultDomainName"`
	DefaultFirestartrGroup          string      `yaml:"defaultFirestartrGroup"`
	DefaultBranch                   string      `yaml:"defaultBranch"`
	DefaultBranchStrategy           string      `yaml:"defaultBranchStrategy"`
	DefaultOrgPermissions           string      `yaml:"defaultOrgPermissions"`
	FinalSecretStoreName            string      `yaml:"finalSecretStoreName"`
	WebhookUrl                      string      // Autocalculated
	WebhookSecretRef                string      // Autocalculated
    PrefappBotPatSecretRef          string      // Autocalculated 
    FirestartrCliVersionSecretRef   string      // Autocalculated
	HasFreePlan                     bool        // Autocalculated
	BotName                         string      // Stored in Credentialsfile.yaml, but needed here for templating
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
    CliVersion string   `yaml:"cli"`
}

type CredsFile struct {
	CloudProvider CloudProvider `yaml:"cloudProvider"`
	GithubApp     GithubApp     `yaml:"githubApp"`
}

type CloudProvider struct {
	ProviderConfigName string         `yaml:"providerConfigName"`
	Config             ConfigProvider `yaml:"config"`
	Source             string         `yaml:"source"`
	Type               string         `yaml:"type"`
	Version            string         `yaml:"version"`
	Name               string         `yaml:"name"`
}

type ConfigProvider struct {
	Bucket    *string `json:"bucket" yaml:"bucket"`
	Region    string `json:"region" yaml:"region"`
	AccessKey string `json:"access_key" yaml:"access_key"`
	SecretKey string `json:"secret_key" yaml:"secret_key"`
	Token     string `json:"token" yaml:"token"`
}

type GithubApp struct {
	ProviderConfigName string `yaml:"providerConfigName"`
	Owner              string `yaml:"owner"`
	BotName            string `yaml:"botName"`
    PrefappBotPat      string `yaml:"prefappBotPat"`
	Pem                string
	RawPem             string
	GhAppId            string
	InstallationId     string
	BotPat             string
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


// DeploymentWebhook representa el bloque de la URL y el Secreto del Webhook.
type DeploymentWebhook struct {
	URL    string
	Secret string
}


// DeploymentExternalSecrets representa la referencia al ARN del rol para External Secrets.
type DeploymentExternalSecrets struct {
	RoleARN string
}

// DeploymentController representa la información de la aplicación GitHub usada por el controller.
type DeploymentController struct {
	Image         string
	RoleARN       string
	GithubApp	  DeploymentGithubApp
}

// DeploymentAws representa la configuración de AWS específica (Bucket y Region).
type DeploymentAws struct {
	Bucket string
	Region string
}

type DeploymentGithubApp struct {

	GithubAppId   string
	GithubAppPem  string
	GithubAppInstallationId string
}


// DeploymentConfig contiene solo los campos de nivel superior que son interpolables.
type DeploymentConfig struct {
	Customer        string
	Org				string
	Webhook         DeploymentWebhook
	ExternalSecrets DeploymentExternalSecrets
	Controller      DeploymentController
	Aws             DeploymentAws
	Provider        DeploymentGithubApp
}

type PushSecretElement  struct {
    Name                    string
    SecretStore             string
    KubernetesSecret        string
    KubernetesSecretKey     string
    ParameterName           string
    Value                   string
}
