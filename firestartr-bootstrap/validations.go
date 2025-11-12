package main

import (
	"context"
	"dagger/firestartr-bootstrap/internal/dagger"
	"fmt"
	"strings"

	"github.com/xeipuuv/gojsonschema"
	"sigs.k8s.io/yaml"
)

func (m *FirestartrBootstrap) ValidateBootstrapFile(ctx context.Context, bootstrapFile *dagger.File) error {
	schema, err := dag.CurrentModule().Source().File("schemas/bootstrap-file.json").Contents(ctx)
	if err != nil {
		return err
	}

	bootstrapFileContents, err := bootstrapFile.Contents(ctx)
	if err != nil {
		return err
	}

	json, err := yaml.YAMLToJSON([]byte(bootstrapFileContents))
	if err != nil {
		return err
	}

	if err := validateDocumentSchema(schema, string(json)); err != nil {
		return fmt.Errorf("failed to validate bootstrap file: %w", err)
	}
	return nil
}

func (m *FirestartrBootstrap) ValidateCredentialsFile(ctx context.Context, credentialsFileContents string) error {
	schema, err := dag.CurrentModule().Source().File("schemas/credentials-file.json").Contents(ctx)
	if err != nil {
		return err
	}

	jsonDoc, err := yaml.YAMLToJSON([]byte(credentialsFileContents))
	if err != nil {
		return err
	}

	if err := validateDocumentSchema(string(jsonDoc), schema); err != nil {
		return fmt.Errorf("failed to validate credentials file: %w", err)
	}
	return nil
}

func validateDocumentSchema(document string, schema string) error {
	schemaLoader := gojsonschema.NewStringLoader(schema)
	documentLoader := gojsonschema.NewStringLoader(document)

	res, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return err
	}

	if !res.Valid() {
		return fmt.Errorf("document is not valid %s", res.Errors())
	}
	return nil
}

func (m *FirestartrBootstrap) GithubRepositoryExists(ctx context.Context, repo string, ghToken *dagger.Secret) (bool, error) {
	ctr, err := m.GhContainer(ctx, ghToken).
		WithExec([]string{
			"gh",
			"repo",
			"view",
			fmt.Sprintf("%s/%s", m.Bootstrap.Org, repo),
		}, dagger.ContainerWithExecOpts{
			RedirectStdout: "/tmp/stdout",
			RedirectStderr: "/tmp/stderr",
			Expect:         "ANY",
		}).
		Sync(ctx)
	if err != nil {
		return false, err
	}

	stderr, err := ctr.File("/tmp/stderr").Contents(ctx)
	if err != nil {
		return false, err
	}
	stdout, err := ctr.File("/tmp/stdout").Contents(ctx)
	if err != nil {
		return false, err
	}

	fmt.Printf("stdout: %s\n", stdout)
	fmt.Printf("stderr: %s\n", stderr)

	eC, err := ctr.ExitCode(ctx)
	if err != nil {
		return false, err
	}

	if eC != 0 {

		if strings.Contains(stderr, "Could not resolve to a Repository with the name") {
			return false, nil
		}

		return false, fmt.Errorf("failed to check if repository exists: %s", stderr)
	}

	return true, nil
}

func (m *FirestartrBootstrap) CheckAlreadyCreatedRepositories(
	ctx context.Context,
	ghToken *dagger.Secret,
) ([]string, error) {
	alreadyCreatedRepos := []string{}

	for idx := range m.Bootstrap.Components {
		component := m.Bootstrap.Components[idx]
		if !component.Skipped {
			repoName := component.Name
			if component.RepoName != "" {
				repoName = component.RepoName
			}
			exists, err := m.GithubRepositoryExists(ctx, repoName, ghToken)
			if err != nil {
				return nil, err
			}
			if exists {
				component.Skipped = true
				m.Bootstrap.Components[idx] = component
				alreadyCreatedRepos = append(alreadyCreatedRepos, repoName)
			}
		}
	}

	return alreadyCreatedRepos, nil
}

