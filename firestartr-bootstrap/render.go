package main

import (
	"bytes"
	"context"
	"dagger/firestartr-bootstrap/internal/dagger"
	"fmt"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

func (m *FirestartrBootstrap) RenderCrs(ctx context.Context) (*dagger.Directory, error) {

	initialCrsTemplate, err := m.RenderInitialCrs(ctx,
		dag.CurrentModule().
			Source().
			File("templates/initial_crs.tmpl"),
	)
	if err != nil {
		return nil, err
	}

	initialCrsDir, err := m.SplitRenderedCrsInFiles(initialCrsTemplate)
	if err != nil {
		return nil, err
	}

	renderedClaims, err := m.RenderBootstrapFile(
		ctx,
		dag.CurrentModule().
			Source().
			File("templates/components.tmpl"),
	)
	if err != nil {
		return nil, err
	}

	claimsDir, err := m.SplitRenderedClaimsInFiles(renderedClaims)
	if err != nil {
		return nil, err
	}

	crsDir := initialCrsDir.WithoutFiles(
		[]string{
			"FirestartrProviderConfig.github-app.yml",
			"FirestartrProviderConfig.firestartr-terraform-state.yml",
		},
	)

	if m.PreviousCrsDir != nil {
		crsDir = m.PreviousCrsDir
	}

	firestartrCrsDir, err := m.RenderWithFirestartrContainer(
		ctx,
		claimsDir,
		crsDir,
	)

	if err != nil {
		return nil, err
	}

	return dag.Directory().
		WithDirectory(
			"firestartr-crs",
			firestartrCrsDir,
		).
		WithDirectory(
			"initial-crs",
			initialCrsDir.WithoutFile(
				fmt.Sprintf("FirestartrGithubGroup.%s-all-c8bc0fd3-78e1-42e0-8f5c-6b0bb13bb669.yaml",
					m.GhOrg,
				)),
		).
		WithDirectory(
			"claims",
			claimsDir,
		), nil
}

func (m *FirestartrBootstrap) RenderInitialCrs(ctx context.Context, templ *dagger.File) (string, error) {

	templateContent, err := templ.Contents(ctx)
	if err != nil {
		return "", err
	}
	return renderTmpl(templateContent, m.Creds)
}

func (m *FirestartrBootstrap) RenderBootstrapFile(ctx context.Context, templ *dagger.File) (string, error) {

	templateContent, err := templ.Contents(ctx)
	if err != nil {
		return "", err
	}

	return renderTmpl(templateContent, m.Bootstrap)
}

func (m *FirestartrBootstrap) RenderClaimsDefaults(ctx context.Context, templ *dagger.File) (string, error) {

	templateContent, err := templ.Contents(ctx)
	if err != nil {
		return "", err
	}
	return renderTmpl(templateContent, m.Bootstrap)
}

func (m *FirestartrBootstrap) RenderWetReposConfig(ctx context.Context, templ *dagger.File) (string, error) {

	templateContent, err := templ.Contents(ctx)
	if err != nil {
		return "", err
	}
	return renderTmpl(templateContent, m.Bootstrap)
}

func renderTmpl(tmpl string, data interface{}) (string, error) {
	t, err := template.New("template").Funcs(sprig.FuncMap()).Parse(tmpl)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)

	err = t.Execute(buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
