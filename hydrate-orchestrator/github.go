package main

import (
	"context"
	"dagger/hydrate-orchestrator/internal/dagger"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

type Pr struct {
	HeadRefName string `json:"headRefName"`
	Url         string `json:"url"`
	Number      int    `json:"number"`
	State       string `json:"state"`
}

/*
Create or update a PR with the updated contents
*/

func (m *HydrateOrchestrator) upsertPR(
	ctx context.Context,
	// Updated deployment branch name
	// +required
	newBranchName string,
	// Directory containing the updated deployment
	// +required
	contents *dagger.Directory,
	// Labels to be added to the PR
	// +required
	labels []string,
	// PR title
	// +required
	title string,
	// PR body
	// +required
	body string,
	// Clean up the directory
	// +required
	cleanupDir string,
	// PR author
	// +optional
	reviewers []string,

) (string, error) {

	branchWithId := newBranchName

	m.upsertRemoteBranch(ctx, contents, branchWithId)

	contentsDirPath := "/contents"
	_, err := dag.Gh(dagger.GhOpts{
		Version: m.GhCliVersion,
	}).Container(dagger.GhContainerOpts{Token: m.GhToken, Plugins: []string{"prefapp/gh-commit"}}).
		WithDirectory(contentsDirPath, contents, dagger.ContainerWithDirectoryOpts{
			Exclude: []string{".git"},
		}).
		WithWorkdir(contentsDirPath).
		WithEnvVariable("CACHE_BUSTER", time.Now().String()).
		WithExec([]string{
			"gh",
			"commit",
			"-R", m.Repo,
			"-b", branchWithId,
			"-m", "Update deployments",
			"--delete-path", cleanupDir,
		}).Sync(ctx)

	if err != nil {
		return "", err
	}

	cmd := []string{
		"gh",
		"pr",
		"create",
		"-R", m.Repo,
		"--base", m.DeploymentBranch,
		"--title", title,
		"--body", body,
		"--head", branchWithId,
	}

	for _, label := range labels {
		color := m.getColorForLabel(label)
		dag.Gh(dagger.GhOpts{
			Version: m.GhCliVersion,
			Token:   m.GhToken,
		}).Run(
			fmt.Sprintf("label create -R %s --force --color %s %s", m.Repo, color, label), dagger.GhRunOpts{DisableCache: true}).Sync(ctx)
		cmd = append(cmd, "--label", label)
	}

	for _, reviewer := range reviewers {
		cmd = append(cmd, "--reviewer", reviewer)
	}

	// Create a PR for the updated deployment
	stdout, err := dag.Gh().Container(dagger.GhContainerOpts{
		Version: m.GhCliVersion,
		Token:   m.GhToken,
	}).
		WithEnvVariable(
			"CACHE_BUSTER",
			time.Now().String()).WithDirectory(contentsDirPath, contents).
		WithWorkdir(contentsDirPath).
		WithExec(cmd).
		Stdout(ctx)

	if err != nil {
		return "", err
	}

	return stdout, nil

}

func (m *HydrateOrchestrator) prExists(ctx context.Context, branchName string) (*Pr, error) {

	// branch name depends on the deployment kind, the format is <depKindId>-<depKind>-<cluster>-<tenant>-<env>
	//                                                           0-kubernetes-cluster-tenant-env
	//														     code-repo-kubernetes-cluster-tenant-env
	prs, err := m.getRepoPrs(ctx)

	if err != nil {
		return nil, err
	}

	for _, pr := range prs {

		if pr.HeadRefName == branchName && strings.ToLower(pr.State) == "open" {

			return &pr, nil

		}

	}

	return nil, nil
}

func (m *HydrateOrchestrator) AutomergeFileExists(ctx context.Context, globPattern string) bool {

	entries, err := m.ValuesStateDir.Glob(ctx, globPattern+"/*")

	if err != nil {

		panic(err)
	}

	automergeFileFound := false

	for _, entry := range entries {

		if fmt.Sprintf("%s/%s", globPattern, "AUTO_MERGE") == entry {

			fmt.Printf("Automerge file found: %s\n", entry)

			automergeFileFound = true

			break
		}
	}

	return automergeFileFound

}

func (m *HydrateOrchestrator) getRepoPrs(ctx context.Context) ([]Pr, error) {

	command := strings.Join([]string{
		"pr",
		"list",
		"--json", "headRefName",
		"--json", "number,url",
		"--json", "state",
		"-L", "1000",
		"-R", m.Repo},
		" ")

	content, err := dag.Gh().Run(command,
		dagger.GhRunOpts{
			Version:      m.GhCliVersion,
			DisableCache: true,
			Token:        m.GhToken}).
		Stdout(ctx)

	if err != nil {

		panic(err)
	}

	prs := []Pr{}

	json.Unmarshal([]byte(content), &prs)

	return prs, nil
}

func (m *HydrateOrchestrator) upsertRemoteBranch(
	ctx context.Context,
	// Base branch name
	// +required
	gitDir *dagger.Directory,
	// New branch name
	// +required
	newBranch string,
) {
	gitDirPath := "/git_dir"
	_, err := dag.Gh().Container(dagger.GhContainerOpts{
		Token:   m.GhToken,
		Version: m.GhCliVersion,
	}).
		WithDirectory(gitDirPath, gitDir).
		WithWorkdir(gitDirPath).
		WithEnvVariable("CACHE_BUSTER", time.Now().String()).
		WithExec([]string{
			"git",
			"checkout",
			"-b",
			newBranch,
		}, dagger.ContainerWithExecOpts{},
		).WithExec([]string{
		"git",
		"push",
		"--force",
		"origin",
		newBranch,
	}).Sync(ctx)

	if err != nil {
		panic(err)
	}
}

func (m *HydrateOrchestrator) getColorForLabel(label string) string {
	switch {
	case strings.Contains(label, "app/"): // It is currently redundant but may be useful in the future.
		return "AC1D1C"
	case strings.Contains(label, "tenant/"):
		return "234099"
	case strings.Contains(label, "env/"):
		return "33810B"
	case strings.Contains(label, "service/"): // It is currently redundant but may be useful in the future.
		return "F1C232"
	case strings.Contains(label, "cluster/"):
		return "AC1CAA"
	case strings.Contains(label, "type/"):
		return "6C3B2A"
	default:
		return "7E7C7A"
	}
}

func (m *HydrateOrchestrator) MergePullRequest(ctx context.Context, prLink string) error {

	command := strings.Join([]string{"pr", "merge", "--merge", prLink}, " ")

	_, err := dag.Gh().Run(command, dagger.GhRunOpts{
		Version:      m.GhCliVersion,
		Token:        m.GhToken,
		DisableCache: true,
	}).Sync(ctx)

	if err != nil {
		return err
	}

	log.Printf("PR %s merged successfully", prLink)

	return nil
}
