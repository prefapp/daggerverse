package main

import (
	"context"
	"dagger/hydrate-orchestrator/internal/dagger"
	"encoding/json"
	"fmt"
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
	// Branch ID
	// +required
	branchId int,
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

) error {

	prExists, err := m.checkPrExists(ctx, newBranchName, branchId)

	if err != nil {
		return err
	}

	branchWithId := fmt.Sprintf("%d-%s", branchId, newBranchName)

	if !prExists {
		m.createRemoteBranch(ctx, contents, branchWithId)
	}

	contentsDirPath := "/contents"
	_, err = dag.Gh().Container(dagger.GhContainerOpts{Token: m.GhToken, Plugins: []string{"prefapp/gh-commit"}}).
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
		return err
	}

	if !prExists {
		labelArgs := ""

		// Create labels and prepare the arguments for the PR creation
		for _, label := range labels {
			dag.Gh(dagger.GhOpts{Token: m.GhToken}).Run(fmt.Sprintf("label create -R %s --force %s", m.Repo, label), dagger.GhRunOpts{DisableCache: true}).Sync(ctx)
			labelArgs += fmt.Sprintf(" --label '%s'", label)
		}

		reviewerArgs := ""
		for _, reviewer := range reviewers {
			reviewerArgs += fmt.Sprintf(" --reviewer '%s'", reviewer)
		}

		// Create a PR for the updated deployment
		_, err := dag.Gh().Container(dagger.GhContainerOpts{
			Token: m.GhToken,
		}).
			WithEnvVariable(
				"CACHE_BUSTER",
				time.Now().String(),
			).
			WithDirectory(contentsDirPath, contents).
			WithWorkdir(contentsDirPath).
			WithExec([]string{
				"gh",
				"pr",
				"create",
				"-R", m.Repo,
				"--base", m.DeploymentBranch,
				"--title", title,
				"--body", body,
				"--head", branchWithId,
				labelArgs,
				reviewerArgs,
			}).Sync(ctx)

		if err != nil {
			return err
		}

	}

	return nil
}

func (m *HydrateOrchestrator) checkPrExists(ctx context.Context, branchName string, branchId int) (bool, error) {

	// branch name depends on the deployment kind, the format is <depKindId>-<depKind>-<cluster>-<tenant>-<env>

	prs, err := m.getRepoPrs(ctx)

	if err != nil {
		return false, err
	}

	for _, pr := range prs {
		if strings.HasSuffix(pr.HeadRefName, branchName) &&
			!strings.HasPrefix(pr.HeadRefName, fmt.Sprintf("%d-", branchId)) &&
			strings.ToLower(pr.State) == "open" {
			return false, fmt.Errorf("Deployment pending (%s) with branch name %s", branchName, pr.HeadRefName)
		} else if strings.HasSuffix(pr.HeadRefName, branchName) &&
			strings.ToLower(pr.State) == "open" {
			return true, nil
		}

	}
	return false, nil
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

	content, err := dag.Gh().Run(command, dagger.GhRunOpts{DisableCache: true, Token: m.GhToken}).Stdout(ctx)

	if err != nil {

		panic(err)
	}

	prs := []Pr{}

	json.Unmarshal([]byte(content), &prs)

	return prs, nil
}

func (m *HydrateOrchestrator) createRemoteBranch(
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
		Token: m.GhToken,
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
		"origin",
		newBranch,
	}).Sync(ctx)

	if err != nil {
		panic(err)
	}
}
