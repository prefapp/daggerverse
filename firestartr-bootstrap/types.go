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
	Firestartr             Firestartr  `yaml:"firestartr"`
	PushFiles              PushFiles   `yaml:"pushFiles"`
	Org                    string      `yaml:"org"`
    Customer               string      `yaml:"customer"`
	Components             []Component `yaml:"components"`
	DefaultSystemName      string      `yaml:"defaultSystemName"`
	DefaultDomainName      string      `yaml:"defaultDomainName"`
	DefaultFirestartrGroup string      `yaml:"defaultFirestartrGroup"`
	DefaultBranch          string      `yaml:"defaultBranch"`
	DefaultBranchStrategy  string      `yaml:"defaultBranchStrategy"`
	DefaultOrgPermissions  string      `yaml:"defaultOrgPermissions"`
	FinalSecretStoreName   string      `yaml:"finalSecretStoreName"`
	WebhookUrl             string      `yaml:"webhookUrl"`
	HasFreePlan            bool        // Autocalculated
	BotName                string      // Stored in Credentialsfile.yaml, but needed here for templating
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
	Version string `yaml:"version"`
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
	Bucket    string `json:"bucket" yaml:"bucket"`
	Region    string `json:"region" yaml:"region"`
	AccessKey string `json:"access_key" yaml:"access_key"`
	SecretKey string `json:"secret_key" yaml:"secret_key"`
	Token     string `json:"token" yaml:"token"`
}

type GithubApp struct {
	ProviderConfigName string `yaml:"providerConfigName"`
	Owner              string `yaml:"owner"`
	BotName            string `yaml:"botName"`
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
