package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestFirestartrChartVersionAcrossDeploymentModes(t *testing.T) {
	for _, file := range []string{
		"templates/deployment/tenant.tmpl",
		"templates/deployment/azure_tenant.tmpl",
	} {
		t.Run(file, func(t *testing.T) {
			content, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			var descriptor struct {
				Version string `yaml:"version"`
			}
			if err := yaml.Unmarshal(content, &descriptor); err != nil {
				t.Fatal(err)
			}
			if descriptor.Version != "4.3.0" {
				t.Errorf("chart version = %q, want 4.3.0", descriptor.Version)
			}
		})
	}
}

func TestDedicatedFirestartrReleaseNameMatchesInitialApply(t *testing.T) {
	content, err := os.ReadFile("templates/deployment/azure_tenant.tmpl")
	if err != nil {
		t.Fatal(err)
	}
	rendered, err := renderTmpl(string(content), map[string]string{"Customer": "azprefapp"})
	if err != nil {
		t.Fatal(err)
	}
	var descriptor struct {
		ReleaseName string `yaml:"releaseName"`
	}
	if err := yaml.Unmarshal([]byte(rendered), &descriptor); err != nil {
		t.Fatal(err)
	}
	if descriptor.ReleaseName != "firestartr" {
		t.Fatalf("state-sys-services release name = %q, want firestartr", descriptor.ReleaseName)
	}
	if !strings.Contains(sysServicesApplyScript(), fmt.Sprintf("helm template %s firestartr-controller/firestartr", descriptor.ReleaseName)) {
		t.Fatalf("initial sys-services apply does not use release name %q", descriptor.ReleaseName)
	}
}

func TestDedicatedArgoCDHierarchy(t *testing.T) {
	files, err := renderDedicatedArgoCDHierarchy("example-org")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 18 {
		t.Fatalf("expected two root Applications and eight ApplicationSet/AppProject pairs, got %d", len(files))
	}
	namespaces := map[string]string{
		"firestartr":                 "firestartr",
		"nginx":                      "ingress-nginx",
		"cert-manager":               "cert-manager",
		"external-dns":               "external-dns",
		"argocd":                     "argocd",
		"argo-configuration-secrets": "argocd",
		"argo-events":                "argo-events",
		"argo-workflows":             "argo-workflows",
	}
	for _, name := range []string{"apps", "sys-services"} {
		var application struct {
			Kind     string `yaml:"kind"`
			Metadata struct {
				Namespace string `yaml:"namespace"`
			} `yaml:"metadata"`
			Spec struct {
				Source struct {
					RepoURL   string `yaml:"repoURL"`
					Path      string `yaml:"path"`
					Directory struct {
						Recurse bool `yaml:"recurse"`
					} `yaml:"directory"`
				} `yaml:"source"`
				Destination Destination `yaml:"destination"`
			} `yaml:"spec"`
		}
		if err := yaml.Unmarshal([]byte(files["Application."+name+".yaml"]), &application); err != nil {
			t.Fatal(err)
		}
		if application.Kind != "Application" || application.Metadata.Namespace != "argocd" || application.Spec.Destination.Namespace != "argocd" || application.Spec.Source.Path != "./"+name || application.Spec.Source.RepoURL != "https://github.com/example-org/state-argocd" || !application.Spec.Source.Directory.Recurse {
			t.Fatalf("invalid %s root Application: %+v", name, application)
		}
	}
	for file, content := range files {
		if !strings.Contains(file, "/ApplicationSet.") {
			continue
		}
		t.Run(file, func(t *testing.T) {
			var applicationSet struct {
				Kind string `yaml:"kind"`
				Spec struct {
					Generators []struct {
						Git struct {
							RepoURL     string `yaml:"repoURL"`
							Revision    string `yaml:"revision"`
							Directories []struct {
								Path string `yaml:"path"`
							} `yaml:"directories"`
						} `yaml:"git"`
					} `yaml:"generators"`
					Template struct {
						Spec struct {
							Project string `yaml:"project"`
							Source  struct {
								RepoURL        string `yaml:"repoURL"`
								Path           string `yaml:"path"`
								TargetRevision string `yaml:"targetRevision"`
							} `yaml:"source"`
							Destination Destination `yaml:"destination"`
						} `yaml:"spec"`
					} `yaml:"template"`
				} `yaml:"spec"`
			}
			if err := yaml.Unmarshal([]byte(content), &applicationSet); err != nil {
				t.Fatal(err)
			}
			projectFile := strings.Replace(file, "/ApplicationSet.", "/AppProject.", 1)
			var project AppProject
			if err := yaml.Unmarshal([]byte(files[projectFile]), &project); err != nil {
				t.Fatal(err)
			}
			if applicationSet.Kind != "ApplicationSet" || project.Kind != "AppProject" || len(applicationSet.Spec.Generators) != 1 || len(project.Spec.Destinations) != 1 {
				t.Fatalf("invalid ApplicationSet or project pair for %s", file)
			}
			generator := applicationSet.Spec.Generators[0].Git
			template := applicationSet.Spec.Template.Spec
			service := filepath.Base(filepath.Dir(file))
			if wantNamespace, exists := namespaces[service]; !exists || template.Destination.Namespace != wantNamespace {
				t.Fatalf("unexpected namespace %q for service %q", template.Destination.Namespace, service)
			}
			wantPath := "kubernetes-sys-services/firestartr-aks/" + service
			if generator.RepoURL != "https://github.com/example-org/state-sys-services.git" || generator.Revision != "deployment" || len(generator.Directories) != 1 || generator.Directories[0].Path != wantPath || template.Source.Path != wantPath || template.Source.TargetRevision != "deployment" || template.Source.RepoURL != generator.RepoURL {
				t.Fatalf("invalid rendered deployment source for %s", file)
			}
			if template.Project != project.Metadata["name"] || project.Metadata["namespace"] != "argocd" || template.Destination != project.Spec.Destinations[0] || template.Destination.Server != "https://kubernetes.default.svc" || len(project.Spec.SourceRepos) != 1 || project.Spec.SourceRepos[0] != generator.RepoURL || len(project.Spec.ClusterResourceWhitelist) == 0 {
				t.Fatalf("project does not allow its sys-service application: %+v", project)
			}
		})
	}
}

func TestArgoConfigurationSecretsHTTPRepository(t *testing.T) {
	content, err := os.ReadFile("templates/deployment/sys_services/argo_config_secrets.tmpl")
	if err != nil {
		t.Fatal(err)
	}
	var release map[string]string
	var descriptor struct {
		ReleaseName string `yaml:"releaseName"`
		Registry    string `yaml:"registry"`
		Version     string `yaml:"version"`
		Chart       string `yaml:"chart"`
		Namespace   string `yaml:"namespace"`
	}
	if err := yaml.Unmarshal(content, &descriptor); err != nil {
		t.Fatal(err)
	}
	release = map[string]string{
		"releaseName": descriptor.ReleaseName,
		"registry":    descriptor.Registry,
		"version":     descriptor.Version,
		"chart":       descriptor.Chart,
		"namespace":   descriptor.Namespace,
	}
	for field, want := range map[string]string{
		"releaseName": "argocd-configuration-secrets",
		"registry":    "https://prefapp.github.io/charts/argo-configuration-secrets",
		"version":     "1.0.1",
		"chart":       "argo-configuration-secrets/argocd-configuration-secrets",
		"namespace":   "argocd",
	} {
		if release[field] != want {
			t.Errorf("%s = %q, want %q", field, release[field], want)
		}
	}
	missing, err := missingHelmRepositories(map[string]string{"argo-configuration-secrets": string(content)}, nil)
	if err != nil {
		t.Fatal(err)
	}
	var repository helmRepository
	if err := yaml.Unmarshal([]byte(missing["helm_repositories/argo-configuration-secrets.yaml"]), &repository); err != nil {
		t.Fatal(err)
	}
	if repository.Name != strings.Split(descriptor.Chart, "/")[0] || repository.URL != descriptor.Registry {
		t.Fatalf("HTTP release does not match repository definition: %+v", repository)
	}
}

func TestSysServiceHTTPRepositoryMappings(t *testing.T) {
	required := make(map[string]string)
	for name, file := range sysServiceHelmRepositoryTemplates {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		required[name] = string(content)
	}
	repositories, err := missingHelmRepositories(required, nil)
	if err != nil {
		t.Fatal(err)
	}
	files, err := filepath.Glob("templates/deployment/sys_services/*.tmpl")
	if err != nil {
		t.Fatal(err)
	}
	files = append(files, "templates/deployment/azure_tenant.tmpl")
	for _, file := range files {
		if strings.HasSuffix(file, "_values.tmpl") {
			continue
		}
		t.Run(filepath.Base(file), func(t *testing.T) {
			content, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			var descriptor struct {
				Registry string `yaml:"registry"`
				Chart    string `yaml:"chart"`
			}
			if err := yaml.Unmarshal(content, &descriptor); err != nil {
				t.Fatal(err)
			}
			alias, chart, found := strings.Cut(descriptor.Chart, "/")
			if !found || chart == "" {
				t.Fatalf("expected repository/chart reference, got %q", descriptor.Chart)
			}
			var repository helmRepository
			if err := yaml.Unmarshal([]byte(repositories["helm_repositories/"+alias+".yaml"]), &repository); err != nil {
				t.Fatal(err)
			}
			if repository.Name != alias || strings.TrimRight(repository.URL, "/") != strings.TrimRight(descriptor.Registry, "/") {
				t.Fatalf("HTTP release %q does not match repository definition: %+v", descriptor.Chart, repository)
			}
		})
	}
}

func TestMissingHelmRepositoryRequiresMatchingAlias(t *testing.T) {
	required := map[string]string{"ingress-nginx": "registry: https://kubernetes.github.io/ingress-nginx\n"}
	existing := map[string]string{
		"helm_repositories/other.yaml": "name: other\nurl: https://kubernetes.github.io/ingress-nginx\n",
	}
	missing, err := missingHelmRepositories(required, existing)
	if err != nil {
		t.Fatal(err)
	}
	if _, exists := missing["helm_repositories/ingress-nginx.yaml"]; !exists {
		t.Fatal("a matching URL under another alias must not satisfy the release repository")
	}
	existing["helm_repositories/other.yaml"] = "name: ingress-nginx\nurl: https://example.com/other\n"
	if _, err := missingHelmRepositories(required, existing); err == nil {
		t.Fatal("expected a conflicting repository name to be reported regardless of filename")
	}
}

func TestMissingSysServiceHelmRepositories(t *testing.T) {
	required := make(map[string]string)
	for name, file := range sysServiceHelmRepositoryTemplates {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		required[name] = string(content)
	}
	existing := map[string]string{
		"helm_repositories/firestartr-controller.yaml":      "name: firestartr-controller\nurl: https://prefapp.github.io/charts/firestartr-controller\n",
		"helm_repositories/argo.yaml":                       "name: argo\nurl: https://argoproj.github.io/argo-helm\n",
		"helm_repositories/external-dns.yaml":               "name: external-dns\nurl: https://kubernetes-sigs.github.io/external-dns\n",
		"helm_repositories/argo-configuration-secrets.yaml": "name: argo-configuration-secrets\nurl: https://prefapp.github.io/charts/argo-configuration-secrets\n",
	}
	missing, err := missingHelmRepositories(required, existing)
	if err != nil {
		t.Fatal(err)
	}
	if len(missing) != 2 {
		t.Fatalf("expected only nginx and cert-manager to be missing, got %v", missing)
	}
	for _, name := range []string{"ingress-nginx", "cert-manager"} {
		file := "helm_repositories/" + name + ".yaml"
		var repository helmRepository
		if err := yaml.Unmarshal([]byte(missing[file]), &repository); err != nil {
			t.Fatal(err)
		}
		if repository.Name != name || repository.URL == "" {
			t.Fatalf("invalid repository definition: %+v", repository)
		}
		existing[file] = missing[file]
	}
	missing, err = missingHelmRepositories(required, existing)
	if err != nil || len(missing) != 0 {
		t.Fatalf("merged configuration should have no missing repositories: %v, %v", missing, err)
	}
	existing["helm_repositories/ingress-nginx.yaml"] = "name: ingress-nginx\nurl: https://example.com/other\n"
	if _, err := missingHelmRepositories(required, existing); err == nil {
		t.Fatal("expected conflicting configuration to be preserved and reported")
	}
}

func TestApplySysServicesUsesNamespaceBootstrapAndServerSideApply(t *testing.T) {
	cmd := sysServicesApplyScript()

	for _, want := range []string{
		"kubectl create namespace",
		"external-secrets",
		"firestartr",
		"CustomResourceDefinitions.yaml",
		"argo-configuration-secrets",
		"firestartr-values.yaml",
		"kubernetes-sys-services/firestartr-aks",
		"kubectl apply -n firestartr --server-side --force-conflicts --field-manager=firestartr-bootstrap -f -",
	} {
		if !strings.Contains(cmd, want) {
			t.Fatalf("expected command to contain %q", want)
		}
	}
	if strings.Contains(cmd, "kubectl apply -f -") && !strings.Contains(cmd, "--server-side") {
		t.Fatalf("expected server-side apply for rendered manifests")
	}
}

func TestExternalDnsFederatedCredentialUsesExplicitSubject(t *testing.T) {
	if !strings.Contains("system:serviceaccount:external-dns:external-dns", "system:serviceaccount:external-dns:external-dns") {
		t.Fatal("expected explicit external-dns workload identity subject")
	}
}
