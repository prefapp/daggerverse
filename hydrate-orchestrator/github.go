package main

import (
	"context"
	"dagger/hydrate-orchestrator/internal/dagger"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func (m *HydrateOrchestrator) upsertPR(
	ctx context.Context,
	// +required
	newBranchName string,
	// +required
	contents *dagger.Directory,
) {
	contentsDirPath := "/contents"
	dag.Gh().Container(dagger.GhContainerOpts{Token: m.GhToken, Plugins: []string{"prefapp/gh-commit"}}).
		WithDirectory(contentsDirPath, contents, dagger.ContainerWithDirectoryOpts{
			Exclude: []string{".git"},
		}).
		WithWorkdir(contentsDirPath).
		WithEnvVariable("CACHE_BUSTER", time.Now().String()).
		WithExec([]string{
			"gh",
			"commit",
			"-R", m.Repo,
			"-b", newBranchName,
		}).Sync(ctx)

	prExists := m.checkPrExists(ctx, newBranchName)

	if !prExists {
		// Create a PR for the updated deployment
		dag.Gh().Run(fmt.Sprintf("pr create -R '%s' --base '%s' --title 'Update deployment' --body 'Update deployment' --head %s", m.Repo, m.DeploymentBranch, newBranchName),
			dagger.GhRunOpts{
				DisableCache: true,
				Token:        m.GhToken,
			},
		).Sync(ctx)
	}
}

type Pr struct {
	HeadRefName string `json:"headRefName"`
	Url         string `json:"url"`
	Number      int    `json:"number"`
}

func (m *HydrateOrchestrator) checkPrExists(ctx context.Context, branchName string) bool {

	prs, err := m.getRepoPrs(ctx)

	if err != nil {
		panic(err)
	}

	for _, pr := range prs {
		if pr.HeadRefName == branchName {
			return true
		}
	}
	return false
}

func (m *HydrateOrchestrator) getRepoPrs(ctx context.Context) ([]Pr, error) {

	command := strings.Join([]string{
		"pr",
		"list",
		"--json",
		"headRefName",
		"--json",
		"number,url",
		"-L",
		"1000",
		"-R",
		m.Repo},
		" ")

	content, err := dag.Gh().Run(command, dagger.GhRunOpts{DisableCache: true, Token: m.GhToken}).Stdout(ctx)

	if err != nil {

		return nil, err
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
	dag.Gh().Container(dagger.GhContainerOpts{
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
}
