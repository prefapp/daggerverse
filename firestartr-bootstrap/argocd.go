package main

import (
	"context"
	"fmt"

	"dagger/firestartr-bootstrap/internal/dagger"

	"gopkg.in/yaml.v3"
)

func renderDedicatedArgoCDHierarchy(org string) (map[string]string, error) {
	files := make(map[string]string)
	writeManifest := func(file string, manifest interface{}) error {
		content, err := yaml.Marshal(manifest)
		if err != nil {
			return fmt.Errorf("rendering %s: %w", file, err)
		}
		files[file] = string(content)
		return nil
	}
	server := "https://kubernetes.default.svc"
	argoRepo := fmt.Sprintf("https://github.com/%s/state-argocd", org)
	servicesRepo := fmt.Sprintf("https://github.com/%s/state-sys-services.git", org)
	syncPolicy := map[string]interface{}{
		"automated":   map[string]interface{}{"prune": true},
		"syncOptions": []string{"PruneLast=true", "CreateNamespace=true", "ServerSideApply=true"},
	}
	for _, name := range []string{"apps", "sys-services"} {
		manifest := map[string]interface{}{
			"apiVersion": "argoproj.io/v1alpha1",
			"kind":       "Application",
			"metadata":   map[string]interface{}{"name": name, "namespace": "argocd"},
			"spec": map[string]interface{}{
				"project": "default",
				"source": map[string]interface{}{
					"repoURL": argoRepo, "path": "./" + name, "targetRevision": "HEAD",
					"directory": map[string]interface{}{"recurse": true, "exclude": ".gitkeep"},
				},
				"destination": Destination{Namespace: "argocd", Server: server},
				"syncPolicy":  syncPolicy,
			},
		}
		if err := writeManifest(fmt.Sprintf("Application.%s.yaml", name), manifest); err != nil {
			return nil, err
		}
	}
	services := []struct {
		Directory string
		Name      string
		Namespace string
	}{
		{"firestartr", "firestartr", "firestartr"},
		{"nginx", "ingress-nginx", "ingress-nginx"},
		{"cert-manager", "cert-manager", "cert-manager"},
		{"external-dns", "external-dns", "external-dns"},
		{"argocd", "argocd", "argocd"},
		{"argo-configuration-secrets", "argocd-configuration-secrets", "argocd"},
		{"argo-events", "argo-events", "argo-events"},
		{"argo-workflows", "argo-workflows", "argo-workflows"},
	}
	for _, service := range services {
		sourcePath := fmt.Sprintf("kubernetes-sys-services/%s/%s", DeploymentPlatformAKS, service.Directory)
		metadata := map[string]interface{}{"name": service.Name, "namespace": "argocd"}
		project := AppProject{
			APIVersion: "argoproj.io/v1alpha1",
			Kind:       "AppProject",
			Metadata:   metadata,
			Spec: AppProjectSpec{
				Description:  "Project for " + service.Name,
				SourceRepos:  []string{servicesRepo},
				Destinations: []Destination{{Namespace: service.Namespace, Server: server}},
				ClusterResourceWhitelist: []interface{}{
					map[string]string{"group": "rbac.authorization.k8s.io", "kind": "ClusterRole"},
					map[string]string{"group": "rbac.authorization.k8s.io", "kind": "ClusterRoleBinding"},
					map[string]string{"group": "", "kind": "Namespace"},
					map[string]string{"group": "apiextensions.k8s.io", "kind": "CustomResourceDefinition"},
				},
			},
		}
		if service.Name == "ingress-nginx" || service.Name == "cert-manager" || service.Name == "argocd" {
			project.Spec.ClusterResourceWhitelist = append(project.Spec.ClusterResourceWhitelist,
				map[string]string{"group": "admissionregistration.k8s.io", "kind": "ValidatingWebhookConfiguration"},
				map[string]string{"group": "admissionregistration.k8s.io", "kind": "MutatingWebhookConfiguration"},
			)
		}
		if err := writeManifest(fmt.Sprintf("sys-services/%s/AppProject.%s.yaml", service.Directory, service.Name), project); err != nil {
			return nil, err
		}
		applicationSet := map[string]interface{}{
			"apiVersion": "argoproj.io/v1alpha1",
			"kind":       "ApplicationSet",
			"metadata":   metadata,
			"spec": map[string]interface{}{
				"generators": []interface{}{map[string]interface{}{"git": map[string]interface{}{
					"repoURL": servicesRepo, "revision": "deployment",
					"directories": []interface{}{map[string]string{"path": sourcePath}},
				}}},
				"template": map[string]interface{}{
					"metadata": map[string]string{"name": DeploymentPlatformAKS + "-" + service.Name},
					"spec": map[string]interface{}{
						"project":     service.Name,
						"source":      map[string]string{"repoURL": servicesRepo, "path": sourcePath, "targetRevision": "deployment"},
						"destination": Destination{Namespace: service.Namespace, Server: server},
						"syncPolicy":  syncPolicy,
					},
				},
			},
		}
		if err := writeManifest(fmt.Sprintf("sys-services/%s/ApplicationSet.%s.yaml", service.Directory, service.Name), applicationSet); err != nil {
			return nil, err
		}
	}
	return files, nil
}

func (m *FirestartrBootstrap) CreateArgCDApplications(
	ctx context.Context,
) (*dagger.Directory, error) {
	argoCDRenderedDir, err := m.RenderArgoCDApplications(ctx)

	if err != nil {
		return nil, fmt.Errorf("rendering ArgoCD apps: %w", err)
	}

	tokenSecret := dag.SetSecret(
		"token",
		m.Creds.GithubApp.OperatorPat,
	)

	// For dedicated deployments, target the customer's own org; for SaaS,
	// target the shared firestartr-<env> org.
	var argoOrg string
	if m.isDedicatedDeployment() {
		argoOrg = m.Bootstrap.Org
	} else {
		argoOrg = fmt.Sprintf("firestartr-%s", m.Bootstrap.Env)
	}

	argoCDRepo, err := m.CloneRepo(
		ctx,
		argoOrg,
		"state-argocd",
		tokenSecret,
	)

	if err != nil {
		return nil, fmt.Errorf("cloning ArgoCD repo: %w", err)
	}

	if !m.isDedicatedDeployment() {
		// SaaS path: patch the ArgoCD project to add the new namespace as a destination
		projectDir, err := addProjectDestination(
			ctx,
			argoCDRepo.Directory("/repo"),
			"apps/firestartr/argo-firestartr.Project.yaml",
			fmt.Sprintf("%s-firestartr-%s", m.Bootstrap.Customer, m.Bootstrap.Env),
			"https://kubernetes.default.svc",
		)

		if err != nil {
			return nil, fmt.Errorf("adding project destination to ArgoCD: %w", err)
		}

		argoCDRenderedDir = argoCDRenderedDir.WithFile(
			"apps/firestartr/argo-firestartr.Project.yaml",
			projectDir.File("apps/firestartr/argo-firestartr.Project.yaml"),
		)
	}

	err = m.CreatePR(
		ctx,
		"state-argocd",
		argoOrg,
		argoCDRenderedDir,
		fmt.Sprintf("automated-create-applications-%s", m.Bootstrap.Org),
		fmt.Sprintf("ci: add applications for %s [automated]", m.Bootstrap.Org),
		"",
		tokenSecret,
	)

	if err != nil {
		return nil, fmt.Errorf("Error generating PR for firestartr-app deployment: %s", err)
	}

	return argoCDRenderedDir, nil
}

func (m *FirestartrBootstrap) RenderArgoCDApplications(
	ctx context.Context,
) (*dagger.Directory, error) {

	// For dedicated deployments ArgoCD runs in the "argocd" namespace;
	// for SaaS it uses a customer+env-scoped namespace.
	var namespace string
	if m.isDedicatedDeployment() {
		namespace = "firestartr"
	} else {
		namespace = fmt.Sprintf("%s-firestartr-%s", m.Bootstrap.Customer, m.Bootstrap.Env)
	}

	argoCDData := ArgoCDConfig{

		Name: fmt.Sprintf(
			"app-firestartr-firestartr-%s-%s-%s-state-github",
			m.Bootstrap.Env,
			m.Bootstrap.Customer,
			m.Bootstrap.Org,
		),

		App: "state-github",

		Repo: fmt.Sprintf(
			"https://github.com/%s/state-github",
			m.Bootstrap.Org,
		),

		Namespace: namespace,
	}

	argoCDDataInfra := ArgoCDConfig{

		Name: fmt.Sprintf("app-firestartr-firestartr-%s-%s-%s-state-infra",
			m.Bootstrap.Env,
			m.Bootstrap.Customer,
			m.Bootstrap.Org,
		),

		App: "state-infra",

		Repo: fmt.Sprintf("https://github.com/%s/state-infra",
			m.Bootstrap.Org,
		),

		Namespace: namespace,
	}

	argoCDDataSecrets := ArgoCDConfig{

		Name: fmt.Sprintf("app-firestartr-firestartr-%s-%s-%s-state-secrets",
			m.Bootstrap.Env,
			m.Bootstrap.Customer,
			m.Bootstrap.Org,
		),

		App: "state-secrets",

		Repo: fmt.Sprintf("https://github.com/%s/state-secrets",
			m.Bootstrap.Org,
		),

		Namespace: namespace,
	}

	applicationStateGithub, errGithub := renderArgoCDApplication(
		ctx,
		&argoCDData,
	)
	if errGithub != nil {
		return nil, fmt.Errorf("error creating ArgoCD GitHub app: %w", errGithub)
	}

	applicationStateInfra, errInfra := renderArgoCDApplication(
		ctx,
		&argoCDDataInfra,
	)
	if errInfra != nil {
		return nil, fmt.Errorf("error creating ArgoCD infra app: %w", errInfra)
	}

	applicationStateSecrets, errSecrets := renderArgoCDApplication(
		ctx,
		&argoCDDataSecrets,
	)
	if errSecrets != nil {
		return nil, fmt.Errorf("error creating ArgoCD secrets app: %w", errSecrets)
	}

	pathAppStateGithub := fmt.Sprintf(
		"/apps/firestartr/%s/argo-firestartr-%s.%s.Application.yaml",
		m.Bootstrap.Customer,
		"state-github",
		m.Bootstrap.Org,
	)

	pathAppStateInfra := fmt.Sprintf(
		"/apps/firestartr/%s/argo-firestartr-%s.%s.Application.yaml",
		m.Bootstrap.Customer,
		"state-infra",
		m.Bootstrap.Org,
	)

	pathAppStateSecrets := fmt.Sprintf(
		"/apps/firestartr/%s/argo-firestartr-%s.%s.Application.yaml",
		m.Bootstrap.Customer,
		"state-secrets",
		m.Bootstrap.Org,
	)

	applications := dag.Directory().
		WithNewFile(pathAppStateGithub, applicationStateGithub).
		WithNewFile(pathAppStateInfra, applicationStateInfra).
		WithNewFile(pathAppStateSecrets, applicationStateSecrets)
	if m.isDedicatedDeployment() {
		hierarchy, err := renderDedicatedArgoCDHierarchy(m.Bootstrap.Org)
		if err != nil {
			return nil, err
		}
		for file, content := range hierarchy {
			applications = applications.WithNewFile(file, content)
		}
	}
	return applications, nil
}

func renderArgoCDApplication(
	ctx context.Context,
	argoCDData *ArgoCDConfig,
) (string, error) {

	argoCDTemplate := dag.CurrentModule().
		Source().
		File("templates/argocd/application.tmpl")

	templateContent, err := argoCDTemplate.Contents(ctx)
	if err != nil {
		return "", err
	}

	renderedApplication, err := renderTmpl(templateContent, argoCDData)
	if err != nil {
		return "", err
	}

	return renderedApplication, nil
}

func addProjectDestination(
	ctx context.Context,
	sourceDirectory *dagger.Directory,
	fileName string,
	newNamespace string,
	targetServer string,
) (*dagger.Directory, error) {

	yamlContent, err := sourceDirectory.File(fileName).Contents(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", fileName, err)
	}

	var project AppProject

	if err := yaml.Unmarshal([]byte(yamlContent), &project); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML content: %w", err)
	}

	alreadyExists := false

	for _, dest := range project.Spec.Destinations {
		if dest.Namespace == newNamespace {
			alreadyExists = true
			break
		}
	}

	if !alreadyExists {

		newDestination := Destination{
			Namespace: newNamespace,
			Server:    targetServer,
		}
		project.Spec.Destinations = append(project.Spec.Destinations, newDestination)
	}

	modifiedYAMLBytes, err := yaml.Marshal(&project)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal modified object back to YAML: %w", err)
	}

	modifiedYAMLString := string(modifiedYAMLBytes)

	outputDirectory := dag.Directory().
		WithNewFile(fileName, modifiedYAMLString)

	return outputDirectory, nil
}
