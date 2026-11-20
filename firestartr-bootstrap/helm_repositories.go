package main

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"gopkg.in/yaml.v3"
)

var sysServiceHelmRepositoryTemplates = map[string]string{
	"firestartr-controller":      "templates/deployment/azure_tenant.tmpl",
	"ingress-nginx":              "templates/deployment/sys_services/nginx.tmpl",
	"cert-manager":               "templates/deployment/sys_services/cert_manager.tmpl",
	"external-dns":               "templates/deployment/sys_services/external_dns.tmpl",
	"argo":                       "templates/deployment/sys_services/argocd.tmpl",
	"argo-configuration-secrets": "templates/deployment/sys_services/argo_config_secrets.tmpl",
}

type helmRepository struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

func missingHelmRepositories(required map[string]string, existing map[string]string) (map[string]string, error) {
	allowed := make(map[string]string)
	for file, content := range existing {
		var repository helmRepository
		if err := yaml.Unmarshal([]byte(content), &repository); err != nil {
			return nil, fmt.Errorf("parsing %s: %w", file, err)
		}
		repositoryURL := strings.TrimRight(repository.URL, "/")
		if configuredURL, exists := allowed[repository.Name]; exists && configuredURL != repositoryURL {
			return nil, fmt.Errorf("Helm repository name %q has conflicting URLs; update it manually", repository.Name)
		}
		allowed[repository.Name] = repositoryURL
	}

	missing := make(map[string]string)
	for name, descriptor := range required {
		var release struct {
			Registry string `yaml:"registry"`
		}
		if err := yaml.Unmarshal([]byte(descriptor), &release); err != nil {
			return nil, fmt.Errorf("parsing %s release descriptor: %w", name, err)
		}
		registry, err := url.Parse(release.Registry)
		if err != nil || registry.Host == "" || (registry.Scheme != "https" && registry.Scheme != "http") {
			return nil, fmt.Errorf("unsupported HTTP Helm repository for %s: %q", name, release.Registry)
		}
		if configuredURL, exists := allowed[name]; exists {
			if configuredURL != strings.TrimRight(release.Registry, "/") {
				return nil, fmt.Errorf("Helm repository name %q already exists with a different URL; update it manually", name)
			}
			continue
		}
		file := fmt.Sprintf("helm_repositories/%s.yaml", name)
		if _, exists := existing[file]; exists {
			return nil, fmt.Errorf("%s already exists with a different URL; update it manually", file)
		}
		content, err := yaml.Marshal(helmRepository{Name: name, URL: release.Registry})
		if err != nil {
			return nil, fmt.Errorf("encoding %s: %w", file, err)
		}
		missing[file] = string(content)
	}
	return missing, nil
}

func (m *FirestartrBootstrap) missingSysServiceHelmRepositories(ctx context.Context) (map[string]string, error) {
	token := dag.SetSecret("token", m.Creds.GithubApp.OperatorPat)
	repository, err := m.CloneRepo(ctx, m.Bootstrap.Org, ".firestartr", token)
	if err != nil {
		return nil, fmt.Errorf("cloning .firestartr: %w", err)
	}
	existing := make(map[string]string)
	directory := repository.Directory("/repo")
	files, err := directory.Glob(ctx, "helm_repositories/*.yaml")
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		content, err := directory.File(file).Contents(ctx)
		if err != nil {
			return nil, err
		}
		existing[file] = content
	}
	required := make(map[string]string)
	for name, file := range sysServiceHelmRepositoryTemplates {
		content, err := dag.CurrentModule().Source().File(file).Contents(ctx)
		if err != nil {
			return nil, err
		}
		required[name] = content
	}
	return missingHelmRepositories(required, existing)
}

func (m *FirestartrBootstrap) CmdPushHelmRepositories(ctx context.Context) (string, error) {
	if !m.isDedicatedDeployment() {
		return "Helm repository configuration is only required for dedicated deployments.", nil
	}
	missing, err := m.missingSysServiceHelmRepositories(ctx)
	if err != nil {
		return "", err
	}
	if len(missing) == 0 {
		return "All sys-service Helm repositories are already configured in .firestartr.", nil
	}
	directory := dag.Directory()
	for file, content := range missing {
		directory = directory.WithNewFile(file, content)
	}
	err = m.CreatePR(ctx, ".firestartr", m.Bootstrap.Org, directory,
		fmt.Sprintf("automated-sys-service-helm-repositories-%s", m.Bootstrap.Customer),
		"ci: allow sys-service Helm repositories [automated]", "",
		dag.SetSecret("token", m.Creds.GithubApp.OperatorPat))
	if err != nil {
		return "", fmt.Errorf("creating .firestartr Helm repositories PR: %w", err)
	}
	return fmt.Sprintf("Merge the Helm repositories PR in https://github.com/%s/.firestartr/pulls before running cmd-push-deployment.", m.Bootstrap.Org), nil
}
