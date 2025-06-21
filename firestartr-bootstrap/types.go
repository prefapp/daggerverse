package main

type Component struct {
	Name          string `yaml:"name"`
	RepoName      string `yaml:"repoName"`
	Description   string `yaml:"description"`
	DefaultBranch string `yaml:"defaultBranch"`
	Features      []Feature
	Variables     []Variable `yaml:"variables"`
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
	Firestartr Firestartr  `yaml:"firestartr"`
	PushFiles  PushFiles   `yaml:"pushFiles"`
	Org        string      `yaml:"org"`
	Components []Component `yaml:"components"`
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
	Github PushFilesRepo `yaml:"github"`
}

type Firestartr struct {
	Version string `yaml:"version"`
}

type CredsFile struct {
	CloudProvider CloudProvider `yaml:"cloudProvider"`
	GithubApp     GithubApp     `yaml:"githubApp"`
}

type CloudProvider struct {
	Config  ConfigProvider `yaml:"config"`
	Source  string         `yaml:"source"`
	Type    string         `yaml:"type"`
	Version string         `yaml:"version"`
	Name    string         `yaml:"name"`
}

type ConfigProvider struct {
	Bucket         string `json:"bucket" yaml:"bucket"`
	Region         string `json:"region" yaml:"region"`
	AccessKey      string `json:"access_key" yaml:"access_key"`
	SecretKey      string `json:"secret_key" yaml:"secret_key"`
	UseAzureADAuth bool   `json:"use_azuread_auth" yaml:"use_azuread_auth"`
	TenantId       string `json:"tenant_id" yaml:"tenant_id"`
	ClientId       string `json:"client_id" yaml:"client_id"`
	ClientSecret   string `json:"client_secret" yaml:"client_secret"`
	StorageAccount string `json:"storage_account_name" yaml:"storage_account_name"`
	ContainerName  string `json:"container_name" yaml:"container_name"`
}

type GithubApp struct {
	Pem                   string `yaml:"pem"`
	RawPem                string
	GhAppId               string `yaml:"id"`
	InstallationId        string `yaml:"installationId"`
	PrefappInstallationId string `yaml:"prefappInstallationId"`
	Owner                 string `yaml:"owner"`
	BotName               string `yaml:"botName"`
}
