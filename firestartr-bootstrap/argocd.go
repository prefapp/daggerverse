package main

import (
	"context"
	"fmt"

	"dagger/firestartr-bootstrap/internal/dagger"
)

func (m *FirestartrBootstrap) CreateArgCDApplications(
	ctx context.Context,
) (*dagger.Directory, error){

    argoCDRenderedDir, err := m.RenderDeployment(ctx)

    if err != nil {

        return nil, fmt.Errorf("Rendering argcd apps: %s", err)
    }


	tokenSecret := dag.SetSecret(
		"token",
		m.Creds.GithubApp.OperatorPat,
	)

    err = m.CreatePR(
        ctx,
        "state-argocd",
        fmt.Sprintf("firestartr-%s", m.Env),
        argoCDRenderedDir,
        fmt.Sprintf("automated-create-applications-%s", m.Bootstrap.Org),
        fmt.Sprintf("feat: add applications for %s [automated]", m.Bootstrap.Org),
        fmt.Sprintf("apps/firestartr/%s", m.Bootstrap.Customer),
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

	argoCDData := ArgoCDConfig {

		Name: fmt.Sprintf(
			"app-firestartr-firestartr-%s-%s-%s-state-github",
			m.Env,
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
			m.Env,
		),
	}

	applicationStateGithub, err := renderArgoCDApplication(
		ctx,
		&argoCDData,
	)

	if err != nil {
		return nil, fmt.Errorf("Error creating argocd app: %s", err)
	}
	
	argoCDDataInfra := ArgoCDConfig {

		Name: fmt.Sprintf("app-firestartr-firestartr-%s-%s-%s-state-infra",
			m.Env,
			m.Bootstrap.Customer,
			m.Bootstrap.Org,
		),

		App: "state-github",

		Repo: fmt.Sprintf("https://github.com/%s/state-infra",
			m.Bootstrap.Org,
		),

		Namespace: fmt.Sprintf("%s-firestartr-%s", 
			m.Bootstrap.Customer,
			m.Env,
		),
	}

	applicationStateGithub, errGithub := renderArgoCDApplication(
		ctx,
		&argoCDData,
	)
	
	if errGithub != nil {
		return nil, fmt.Errorf("Error creating argocd app: %s", errGithub)
	}

	applicationStateInfra, errInfra := renderArgoCDApplication(
		ctx,
		&argoCDDataInfra,
	)
	if errInfra != nil {
		return nil, fmt.Errorf("Error creating argocd app: %s", errInfra)
	}


	pathAppStateGithub := fmt.Sprintf(

		"argo-firestartr-%s.%s.Application.yaml",
		"state-github",
		m.Bootstrap.Org,
	)

	pathAppStateInfra := fmt.Sprintf(

		"argo-firestartr-%s.%s.Application.yaml",
		"state-infra",
		m.Bootstrap.Org,
	)

	return dag.Directory().WithNewFile(pathAppStateGithub, applicationStateGithub).WithNewFile(pathAppStateInfra, applicationStateInfra), nil
}

func renderArgoCDApplication(
	ctx context.Context,
	argoCDData *ArgoCDConfig,
) (string, error) {

	argoCDTemplate:= dag.CurrentModule().
		Source().
		File("templates/argocd/application.tmpl")

    // deployment values
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

