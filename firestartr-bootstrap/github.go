package main

import (
	"context"
	"dagger/firestartr-bootstrap/internal/dagger"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"
)

func (m *FirestartrBootstrap) IncludeChanges(
	ctx context.Context,
	dir *dagger.Directory,
	owner string,
	repo string,
	destinyPath string,
	ghToken *dagger.Secret,
) (*dagger.Container, error) {
	ghCtr, err := m.CloneRepo(ctx, owner, repo, ghToken)
	if err != nil {
		return nil, err
	}

	entries, err := dir.Glob(ctx, "**")
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if strings.HasSuffix(entry, "/") {
			continue
		}
		ghCtr = ghCtr.WithFile(
			path.Join("/repo", destinyPath, entry),
			dir.File(entry),
		)
	}

	return ghCtr, nil
}

func (m *FirestartrBootstrap) PushDirToRepo(
	ctx context.Context,
	dir *dagger.Directory,
	repoName string,
	ghToken *dagger.Secret,
) error {
	ghCtr, err := m.IncludeChanges(ctx, dir, m.GhOrg, repoName, "", ghToken)
	if err != nil {
		return err
	}

	_, err = ghCtr.
		WithWorkdir("/repo").
		WithExec([]string{"git", "add", "."}).
		WithExec([]string{"git", "commit", "-m", "automated commit from firestartr-bootstrap"}).
		WithExec([]string{"git", "push"}).
		Sync(ctx)
	if err != nil {
		errMsg := extractErrorMessage(err, "Failed to push changes to repository")
		return errors.New(errMsg)
	}

	return nil
}

func (m *FirestartrBootstrap) CreatePR(
	ctx context.Context,
	repo string,
	owner string,
	dirToPush *dagger.Directory,
	branch string,
	prName string,
	destinyPath string,
	ghToken *dagger.Secret,
) error {
	ghCtr, err := m.IncludeChanges(ctx, dirToPush, owner, repo, destinyPath, ghToken)
	if err != nil {
		return err
	}

	ghCtr, err = ghCtr.
		WithWorkdir("/repo").
		WithExec([]string{"git", "checkout", "-b", branch}).
		WithExec([]string{"git", "add", "."}).
		WithExec([]string{"git", "commit", "-m", "automated commit from firestartr-bootstrap"}).
		WithExec([]string{"git", "push", "origin", branch}).
		WithExec([]string{
			"gh", "pr", "create",
			"--title", prName,
			"--body", "Automated PR created by firestartr-bootstrap",
			"--head", branch,
		}).
		Sync(ctx)
	if err != nil {
		errMsg := extractErrorMessage(err, "Failed to create pull request")
		return errors.New(errMsg)
	}

	return nil
}

func (m *FirestartrBootstrap) CloneRepo(
	ctx context.Context,
	owner string,
	repoName string,
	ghToken *dagger.Secret,
) (*dagger.Container, error) {
	ctr, err := m.GhContainer(ctx, ghToken)
	if err != nil {
		return nil, err
	}

	alpCtr, err := ctr.
		WithEnvVariable("BUST_CACHE", time.Now().String()).
		WithExec([]string{
			"gh",
			"repo",
			"clone",
			fmt.Sprintf("%s/%s", owner, repoName),
			"/repo",
		}).
		Sync(ctx)

	if err != nil {
		errMsg := extractErrorMessage(err, "Failed to clone repository")
		return nil, errors.New(errMsg)
	}

	return alpCtr, nil
}

func (m *FirestartrBootstrap) CreateLabelsInRepo(
	ctx context.Context,
	repoName string,
	labelList []string,
	ghToken *dagger.Secret,
) error {
	ghCtr, err := m.CloneRepo(ctx, m.GhOrg, repoName, ghToken)
	if err != nil {
		return err
	}

	ghCtr = ghCtr.WithWorkdir("/repo")

	for _, label := range labelList {
		ghCtr, err = ghCtr.
			WithExec([]string{
				"gh",
				"label",
				"create",
				label,
			}).
			Sync(ctx)

		if err != nil {
			errMsg := extractErrorMessage(
				err, fmt.Sprintf("Failed to create label in repo %s", repoName),
			)
			return errors.New(errMsg)
		}
	}

	return nil
}

func (m *FirestartrBootstrap) SetOrgVariable(
	ctx context.Context,
	name string,
	value string,
	ghToken *dagger.Secret,
) (*dagger.Container, error) {
	ctr, err := m.GhContainer(ctx, ghToken)
	if err != nil {
		return nil, err
	}

	alpCtr, err := ctr.
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
		errMsg := extractErrorMessage(err, "Failed to set organization variable")
		return nil, errors.New(errMsg)
	}

	return alpCtr, nil
}

func (m *FirestartrBootstrap) SetRepoVariable(
	ctx context.Context,
	repoName string,
	name string,
	value string,
	ghToken *dagger.Secret,
) (*dagger.Container, error) {
	ctr, err := m.GhContainer(ctx, ghToken)
	if err != nil {
		return nil, err
	}

	alpCtr, err := ctr.
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
		errMsg := extractErrorMessage(
			err, fmt.Sprintf("Failed to set variable in repo %s", repoName),
		)
		return nil, errors.New(errMsg)
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

func (m *FirestartrBootstrap) SetOrgVariables(
	ctx context.Context,
	ghToken *dagger.Secret,
	kindContainer *dagger.Container,
) error {

	mappedVars := map[string]string{
		"FS_STATE_APP_ID":  "ref:secretsclaim:firestartr-secrets:fs-state-appid",
		"FS_CHECKS_APP_ID": "ref:secretsclaim:firestartr-secrets:fs-checks-appid",
	}

	for name, ref := range mappedVars {
		value, err := m.GetKubernetesSecretValue(ctx, kindContainer, ref)
		if err != nil {
			errMsg := extractErrorMessage(err, "Failed to get secret value from Kubernetes")
			return errors.New(errMsg)
		}
		_, err = m.SetOrgVariable(ctx, name, value, ghToken)
		if err != nil {
			return err
		}
	}

	m.SetOrgVariable(ctx, "FIRESTARTR_CLI_VERSION", m.Bootstrap.Firestartr.CliVersion, ghToken)
	return nil
}

func (m *FirestartrBootstrap) SetOrgSecret(
	ctx context.Context,
	name string,
	value string,
	ghToken *dagger.Secret,
) (*dagger.Container, error) {
	ctr, err := m.GhContainer(ctx, ghToken)
	if err != nil {
		return nil, err
	}

	alpCtr, err := ctr.
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
		errMsg := extractErrorMessage(err, "Failed to set organization secret")
		return nil, errors.New(errMsg)
	}

	return alpCtr, nil
}

func (m *FirestartrBootstrap) SetOrgSecrets(
	ctx context.Context,
	ghToken *dagger.Secret,
	kindContainer *dagger.Container,
) error {
	mappedVars := map[string]string{
		"FS_STATE_PEM_FILE":  "ref:secretsclaim:firestartr-secrets:fs-state-pem",
		"FS_CHECKS_PEM_FILE": "ref:secretsclaim:firestartr-secrets:fs-checks-pem",
		"PREFAPP_BOT_PAT":    "ref:secretsclaim:firestartr-secrets:prefapp-bot-pat",
	}

	for name, ref := range mappedVars {
		value, err := m.GetKubernetesSecretValue(ctx, kindContainer, ref)
		if err != nil {
			errMsg := extractErrorMessage(err, "Failed to get secret value from Kubernetes")
			return errors.New(errMsg)
		}

		_, err = m.SetOrgSecret(ctx, name, value, ghToken)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *FirestartrBootstrap) GhContainer(
	ctx context.Context,
	ghToken *dagger.Secret,
) (*dagger.Container, error) {
	tokenRaw, err := ghToken.Plaintext(ctx)
	if err != nil {
		return nil, err
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
		}), nil
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
		errMsg := extractErrorMessage(err, "Failed to create GitHub token")
		return nil, errors.New(errMsg)
	}

	tokenRaw, err := ctr.File("/token").Contents(ctx)
	if err != nil {
		errMsg := extractErrorMessage(err, "Failed to read GitHub token")
		return nil, errors.New(errMsg)
	}

	tokenSecret := dag.SetSecret(
		"token",
		tokenRaw,
	)

	return tokenSecret, nil
}

func (m *FirestartrBootstrap) WorkflowRun(
	ctx context.Context,
	jsonInput string,
	workflowFileName string,
	repo string,
	ghToken *dagger.Secret,
) error {
	ctr, err := m.GhContainer(ctx, ghToken)
	if err != nil {
		return err
	}

	_, err = ctr.
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
		errMsg := extractErrorMessage(err, "Failed to run workflow")
		return errors.New(errMsg)
	}

	return nil
}

func (m *FirestartrBootstrap) CheckIfDefaultGroupExists(
	ctx context.Context,
	ghToken *dagger.Secret,
) error {
	ctr, err := m.GhContainer(ctx, ghToken)
	if err != nil {
		return err
	}

	_, err = ctr.
		WithEnvVariable("BUST_CACHE", time.Now().String()).
		WithExec([]string{
			"gh", "api", fmt.Sprintf("/orgs/%s/teams/%s", m.GhOrg, m.Bootstrap.DefaultGroup),
		}).
		Sync(ctx)

	switch err := err.(type) {
	case nil:
		return nil
	case *dagger.ExecError:
		errMsg := extractErrorMessage(
			err,
			fmt.Sprintf(
				"Failed to check if %s group exists in the organization",
				m.Bootstrap.DefaultGroup,
			),
		)
		return errors.New(errMsg)
	default:
		return err
	}
}

func (m *FirestartrBootstrap) CheckIfOrgAllGroupExists(
	ctx context.Context,
	ghToken *dagger.Secret,
) error {
	ctr, err := m.GhContainer(ctx, ghToken)
	if err != nil {
		return err
	}

	_, err = ctr.
		WithEnvVariable("BUST_CACHE", time.Now().String()).
		WithExec([]string{
			"gh", "api", fmt.Sprintf("/orgs/%s/teams/%s-all", m.GhOrg, m.GhOrg),
		}).
		Sync(ctx)

	switch err := err.(type) {
	case nil:
		m.IncludeAllGroup = false
		return nil
	case *dagger.ExecError:
		if strings.Contains(err.Stderr, "404") {
			m.IncludeAllGroup = true
			return nil
		} else {
			errMsg := extractErrorMessage(
				err,
				fmt.Sprintf(
					"Failed to check if %s-all group exists in the organization",
					m.Bootstrap.Org,
				),
			)
			return errors.New(errMsg)
		}
	default:
		return err
	}
}

func (m *FirestartrBootstrap) GetOrganizationPlanName(
	ctx context.Context,
	ghToken *dagger.Secret,
) (string, error) {
	ctr, err := m.GhContainer(ctx, ghToken)
	if err != nil {
		return "", err
	}

	planName, err := ctr.
		WithEnvVariable("BUST_CACHE", time.Now().String()).
		WithExec([]string{
			"gh", "api", fmt.Sprintf("/orgs/%s", m.GhOrg), "--jq", ".plan.name",
		}).
		Stdout(ctx)

	if err != nil {
		errMsg := extractErrorMessage(err, "Failed to get organization plan name")
		return "", errors.New(errMsg)
	}

	return strings.Trim(planName, "\n"), nil
}

func (m *FirestartrBootstrap) OrgHasFreePlan(
	ctx context.Context,
	ghToken *dagger.Secret,
) (bool, error) {
	planName, err := m.GetOrganizationPlanName(ctx, ghToken)
	if err != nil {
		return false, err
	}

	return strings.EqualFold(planName, "free"), nil
}

// Functions to set and get latest feature version info
// Done like this because functions cannot return maps
var latestVersionMap = make(map[string]string)

func clonePrefappRepo(
	ctx context.Context,
	destinationPath string,
	repoName string,
	patValue string,
) (*dagger.Container, error) {

	authURL := fmt.Sprintf("https://%s@github.com/%s/%s.git", patValue, "prefapp", repoName)

	gitArgs := []string{
		"git",
		"clone",
		"--depth", "1",
		"--single-branch", // Only clone one branch/tag
		"--branch", "main",
		authURL,
		destinationPath,
	}

	ctr := dag.Container().
		From("alpine/git:latest").
		WithExec(gitArgs)

	_, err := ctr.Stdout(ctx)
	if err != nil {
		// If the command fails, it indicates an authentication or access issue.
		errorOutput, _ := ctr.Stderr(ctx)

		// Clean up sensitive data from the output for security
		safeOutput := strings.ReplaceAll(errorOutput, patValue, "[REDACTED_PAT]")

		return nil, fmt.Errorf(
			"access check failed. Cannot clone repository: %s", safeOutput,
		)
	}

	return ctr, nil
}

func (m *FirestartrBootstrap) SetLatestFeatureVersionInfo(
	ctx context.Context,
	ghToken *dagger.Secret,
) error {
	destinationPath := "/tmp/features-repo"
	patValue := m.Creds.GithubApp.PrefappBotPat

	ctr, err := clonePrefappRepo(ctx, destinationPath, "features", patValue)
	if err != nil {
		return err
	}

	clonedDir := ctr.Directory(destinationPath)
	featuresVersionJSON, err := clonedDir.
		File(".release-please-manifest.json").
		Contents(ctx)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(featuresVersionJSON), &latestVersionMap)
	if err != nil {
		return err
	}

	return nil
}

func (m *FirestartrBootstrap) GetLatestFeatureVersion(
	ctx context.Context,
	featureName string,
) (string, error) {
	// Inside the latestVersionMap, feature names are prefixed with "packages/"
	latestVersion, ok := latestVersionMap[fmt.Sprintf(
		"packages/%s", featureName,
	)]
	if !ok {
		return "", fmt.Errorf(
			"could not find latest version for feature: %s",
			featureName,
		)
	}

	return latestVersion, nil
}

func getLatestOperatorVersion(
	ctx context.Context,
	pat string,
) (string, error) {
	destinationPath := "/tmp/gitopsk8s-repo"
	operatorVersionMap := make(map[string]string)

	ctr, err := clonePrefappRepo(ctx, destinationPath, "gitops-k8s", pat)
	if err != nil {
		return "", err
	}

	clonedDir := ctr.Directory(destinationPath)
	operatorVersionJSON, err := clonedDir.
		File(".release-please-manifest.json").
		Contents(ctx)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal([]byte(operatorVersionJSON), &operatorVersionMap)
	if err != nil {
		return "", err
	}

	latestVersion, ok := operatorVersionMap["."]
	if !ok {
		return "", fmt.Errorf("could not find latest version for operator")
	}

	return latestVersion, nil
}

func getLatestCliVersion(
	ctx context.Context,
) (string, error) {
	versionsJson, err := dag.Container().
		From("node:20-alpine").
		WithExec([]string{
			"npm",
			"view",
			"@firestartr/cli",
			"versions",
			"--json",
		}).
		Stdout(ctx)

	if err != nil {
		return "", fmt.Errorf("error getting the latest CLI version: %w", err)
	}

	var versions []string
	err = json.Unmarshal([]byte(versionsJson), &versions)
	if err != nil {
		return "", fmt.Errorf("error parsing the CLI version list: %w", err)
	}

	filteredVersions, err := filterStringSlice(versions, `.+snapshot.+`)
	if err != nil {
		return "", fmt.Errorf("error filtering the CLI version list: %w", err)
	}
	if len(filteredVersions) == 0 {
		return "", fmt.Errorf("No CLI versions remaining after filtering snapshots")
	}

	// Return the last version in the list, which is the latest stable version
	return filteredVersions[len(filteredVersions)-1], nil
}

func (m *FirestartrBootstrap) EnableActionsToCreateAndApprovePullRequestsInOrg(
	ctx context.Context,
	ghToken *dagger.Secret,
) error {
	ctr, err := m.GhContainer(ctx, ghToken)
	if err != nil {
		return err
	}

	_, err = ctr.
		WithEnvVariable("BUST_CACHE", time.Now().String()).
		WithExec([]string{
			"gh", "api",
			"--method", "PUT",
			"-H", "Accept: application/vnd.github+json",
			fmt.Sprintf("/orgs/%s/actions/permissions/workflow", m.GhOrg),
			"-f", "default_workflow_permissions=write",
			"-F", "can_approve_pull_request_reviews=true",
		}).
		Sync(ctx)

	if err != nil {
		errorMsg := extractErrorMessage(
			err,
			"Failed to enable actions to create and approve pull requests in organization",
		)
		return errors.New(errorMsg)
	}

	return nil
}
