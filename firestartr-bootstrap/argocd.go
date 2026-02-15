package main

import (
	"context"
	"fmt"

	"dagger/firestartr-bootstrap/internal/dagger"

	"gopkg.in/yaml.v3"
)

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

	argoCDRepo, err := m.CloneRepo(

		ctx,
		fmt.Sprintf("firestartr-%s", m.Bootstrap.Env),
		"state-argocd",
		tokenSecret,
	)

	if err != nil {

		return nil, fmt.Errorf("cloning ArgoCD repo: %w", err)
	}

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
		"firestartr/argo-firestartr.Project.yaml",
		projectDir.File("apps/firestartr/argo-firestartr.Project.yaml"),
	)

	err = m.CreatePR(
		ctx,
		"state-argocd",
		fmt.Sprintf("firestartr-%s", m.Bootstrap.Env),
		argoCDRenderedDir,
		fmt.Sprintf("automated-create-applications-%s", m.Bootstrap.Org),
		fmt.Sprintf("feat: add applications for %s [automated]", m.Bootstrap.Org),
		"apps",
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

		Namespace: fmt.Sprintf("%s-firestartr-%s",
			m.Bootstrap.Customer,
			m.Bootstrap.Env,
		),
	}

	argoCDDataInfra := ArgoCDConfig{

		Name: fmt.Sprintf("app-firestartr-firestartr-%s-%s-%s-state-infra",
			m.Bootstrap.Env,
			m.Bootstrap.Customer,
			m.Bootstrap.Org,
		),

		App: "state-github",

		Repo: fmt.Sprintf("https://github.com/%s/state-infra",
			m.Bootstrap.Org,
		),

		Namespace: fmt.Sprintf("%s-firestartr-%s",
			m.Bootstrap.Customer,
			m.Bootstrap.Env,
		),
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

	return dag.Directory().WithNewFile(pathAppStateGithub, applicationStateGithub).WithNewFile(pathAppStateInfra, applicationStateInfra), nil
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
