package main

import (
	"context"
	"dagger/firestartr-bootstrap/internal/dagger"
	"fmt"
	"path"
	"strings"
	"time"
)

func (m *FirestartrBootstrap) PushDirToRepo(
	ctx context.Context,
	dir *dagger.Directory,
	repoName string,
	ghToken *dagger.Secret,
) error {
	ghCtr, err := m.CloneRepo(ctx, repoName, ghToken)
	if err != nil {
		return err
	}

	entries, err := dir.Glob(ctx, "**")
	if err != nil {
		return err
	}

	for _, entry := range entries {

		if strings.HasSuffix(entry, "/") {
			continue
		}
		ghCtr = ghCtr.WithFile(path.Join("/repo", entry), dir.File(entry))
	}

	ghCtr, err = ghCtr.
		WithWorkdir("/repo").
		WithExec([]string{"git", "add", "."}).
		WithExec([]string{"git", "commit", "-m", "automated commit from firestartr-bootstrap"}).
		WithExec([]string{"git", "push"}).
		Sync(ctx)

	if err != nil {
		return err
	}

	return nil
}

func (m *FirestartrBootstrap) CloneRepo(ctx context.Context, repoName string, ghToken *dagger.Secret) (*dagger.Container, error) {
	alpCtr, err := m.GhContainer(ctx, ghToken).
		WithEnvVariable("BUST_CACHE", time.Now().String()).
		WithExec([]string{
			"gh",
			"repo",
			"clone",
			fmt.Sprintf("%s/%s", m.GhOrg, repoName),
			"/repo",
		}).
		Sync(ctx)

	if err != nil {
		return nil, err
	}

	return alpCtr, nil
}

func (m *FirestartrBootstrap) SetOrgVariable(ctx context.Context, name string, value string, ghToken *dagger.Secret) (*dagger.Container, error) {
	alpCtr, err := m.GhContainer(ctx, ghToken).
		WithEnvVariable("BUST_CACHE", time.Now().String()).
		WithExec([]string{
			"gh",
			"variable",
			"set",
			name,
			"--org",
			m.GhOrg,
			"--body",
			value,
			"--visibility",
			"private",
		}).
		Sync(ctx)

	if err != nil {
		return nil, err
	}

	return alpCtr, nil
}

func (m *FirestartrBootstrap) SetRepoVariable(ctx context.Context, repoName string, name string, value string, ghToken *dagger.Secret) (*dagger.Container, error) {
	alpCtr, err := m.GhContainer(ctx, ghToken).
		WithEnvVariable("BUST_CACHE", time.Now().String()).
		WithExec([]string{
			"gh",
			"variable",
			"set",
			name,
			"--repo",
			fmt.Sprintf("%s/%s", m.GhOrg, repoName),
			"--body",
			value,
		}).
		Sync(ctx)
	if err != nil {
		return nil, err
	}
	return alpCtr, nil
}

func (m *FirestartrBootstrap) SetRepoVariables(ctx context.Context, ghToken *dagger.Secret) error {
	for _, component := range m.Bootstrap.Components {
		for _, variable := range component.Variables {
			repoName := ""
			if component.RepoName != "" {
				repoName = component.RepoName
			} else {
				repoName = component.Name
			}
			if repoName == "" {
				return fmt.Errorf("repoName is empty for component %s", component.Name)
			}
			_, err := m.SetRepoVariable(ctx, repoName, variable.Name, variable.Value, ghToken)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *FirestartrBootstrap) SetOrgVariables(ctx context.Context, ghToken *dagger.Secret) error {

	mappedVars := map[string]string{
		"FIRESTARTER_GITHUB_APP_ID":                      m.Creds.GithubApp.GhAppId,
		"FIRESTARTER_GITHUB_APP_NAME":                    m.Creds.GithubApp.BotName,
		"FIRESTARTER_WORKFLOW_DOCKER_IMAGE_TAG":          fmt.Sprintf("%s_slim", m.Bootstrap.Firestartr.Version),
		"FIRESTARTER_GITHUB_APP_INSTALLATION_ID_PREFAPP": m.Creds.GithubApp.PrefappInstallationId,
		"FIRESTARTER_GITHUB_APP_INSTALLATION_ID":         m.Creds.GithubApp.InstallationId,
		"FIRESTARTR_CLI_VERSION":                         strings.TrimPrefix(m.Bootstrap.Firestartr.Version, "v"),
	}

	for name, value := range mappedVars {
		_, err := m.SetOrgVariable(ctx, name, value, ghToken)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *FirestartrBootstrap) SetOrgSecret(ctx context.Context, name string, value string, ghToken *dagger.Secret) (*dagger.Container, error) {
	alpCtr, err := m.GhContainer(ctx, ghToken).
		WithEnvVariable("BUST_CACHE", time.Now().String()).
		WithExec([]string{
			"gh",
			"secret",
			"set",
			name,
			"--org",
			m.GhOrg,
			"--body",
			value,
			"--visibility",
			"private",
		}).
		Sync(ctx)

	if err != nil {
		return nil, err
	}

	return alpCtr, nil
}

func (m *FirestartrBootstrap) SetOrgSecrets(ctx context.Context, ghToken *dagger.Secret) error {
	mappedVars := map[string]string{
		"FIRESTARTER_GITHUB_APP_PEM_FILE": m.Creds.GithubApp.RawPem,
		"FIRESTARTR_GITHUB_APP_PEM_FILE":  m.Creds.GithubApp.Pem,
	}

	for name, value := range mappedVars {
		_, err := m.SetOrgSecret(ctx, name, value, ghToken)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *FirestartrBootstrap) GhContainer(ctx context.Context, ghToken *dagger.Secret) *dagger.Container {
	tokenRaw, err := ghToken.Plaintext(ctx)
	if err != nil {
		panic(err)
	}

	gitConfig := fmt.Sprintf(
		`[url "https://firestartr:%s@github.com/"]
		 insteadOf = https://github.com/`,
		tokenRaw,
	)

	return dag.Container().
		From("alpine:3.21.3").
		WithExec([]string{"apk", "add", "git", "github-cli"}).
		WithNewFile("/root/.gitconfig", gitConfig).
		WithExec([]string{"git", "config", "--global", "user.name", "firestartr"}).
		WithExec([]string{"git", "config", "--global", "user.email", "info@prefapp.es"}).
		WithExec([]string{"gh", "auth", "login", "--with-token"}, dagger.ContainerWithExecOpts{
			Stdin: tokenRaw,
		})
}

func (m *FirestartrBootstrap) GenerateGithubToken(ctx context.Context) (*dagger.Secret, error) {
	ctr, err := dag.Container().
		From("node:22").
		WithEnvVariable("BUST_CACHE", time.Now().String()).
		WithEnvVariable("GITHUB_APP_ID", m.Creds.GithubApp.GhAppId).
		WithEnvVariable("GITHUB_APP_INSTALLATION_ID", m.Creds.GithubApp.InstallationId).
		WithEnvVariable("GITHUB_APP_PEM_FILE", m.Creds.GithubApp.Pem).
		WithDirectory("/app", dag.CurrentModule().Source().Directory("js")).
		WithWorkdir("/app").
		WithExec([]string{
			"npm", "ci",
		}).
		WithExec([]string{
			"npm", "run", "generate-github-token",
		}).
		Sync(ctx)

	if err != nil {
		return nil, err
	}

	tokenRaw, err := ctr.File("/token").Contents(ctx)

	if err != nil {
		return nil, err
	}

	tokenSecret := dag.SetSecret(
		"token",
		tokenRaw,
	)

	return tokenSecret, nil
}

func (m *FirestartrBootstrap) WorkflowRun(ctx context.Context, jsonInput string, workflowFileName string, repo string, ghToken *dagger.Secret) error {
	_, err := m.GhContainer(ctx, ghToken).
		WithEnvVariable("BUST_CACHE", time.Now().String()).
		WithExec([]string{
			"gh",
			"workflow",
			"run",
			"-R", fmt.Sprintf("%s/%s", m.GhOrg, repo),
			workflowFileName,
			"--json",
		}, dagger.ContainerWithExecOpts{
			Stdin: jsonInput,
		}).
		Sync(ctx)

	if err != nil {
		return err
	}

	return nil
}

func (m *FirestartrBootstrap) RunImportsWorkflow(ctx context.Context, ghToken *dagger.Secret) error {
	err := m.WorkflowRun(
		ctx,
		`{"gh-repo-filter":"SKIP=SKIP","gh-members-filter":"REGEXP=[A-Za-z0-9\\-]+","gh-group-filter":"REGEXP=[A-Za-z0-9\\-]+"}`,
		"github-import.yaml",
		m.Bootstrap.PushFiles.Claims.Repo,
		ghToken,
	)
	if err != nil {
		return err
	}

	return nil
}
