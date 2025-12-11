package main

type AppProjectSpec struct {
	Description              string        `yaml:"description"`
	SourceRepos              []string      `yaml:"sourceRepos"`
	Destinations             []Destination `yaml:"destinations"`
	ClusterResourceWhitelist []interface{} `yaml:"clusterResourceWhitelist"`
}

type Destination struct {
	Namespace string `yaml:"namespace"`
	Server    string `yaml:"server"`
}

type AppProject struct {
	APIVersion string                 `yaml:"apiVersion"`
	Kind       string                 `yaml:"kind"`
	Metadata   map[string]interface{} `yaml:"metadata"`
	Spec       AppProjectSpec         `yaml:"spec"`
}

type ClientAccess struct {
	GithubAppId             PrivateKeyReference `yaml:"githubAppId"`
	GithubAppInstallationId PrivateKeyReference `yaml:"githubAppInstallationId"`
	GithubAppPrivateKey     PrivateKeyReference `yaml:"githubAppPrivateKey"`
}

type PrivateKeyReference struct {
	RemoteRef string `yaml:"remoteRef"` // e.g., /firestartr/rc-prefapp/fs-rc-prefapp-argocd/pem
}
