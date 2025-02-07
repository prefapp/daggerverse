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

) (string, error) {

	prExists, err := m.checkPrExists(ctx, newBranchName, branchId)

	if err != nil {
		return "", err
	}

	branchWithId := fmt.Sprintf("%d-%s", branchId, newBranchName)

	if !prExists {
		m.createRemoteBranch(ctx, contents, branchWithId)
	}

	contentsDirPath := "/contents"
	_, err = dag.Gh(dagger.GhOpts{
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

	if !prExists {
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

		// Create labels and prepare the arguments for the PR creation
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
				time.Now().String(),
			).WithDirectory(contentsDirPath, contents).
			WithWorkdir(contentsDirPath).
			WithExec(cmd).
			Stdout(ctx)

		if err != nil {
			return "", err
		}

		return stdout, nil

	}

	return "", fmt.Errorf("A problem occurred while creating the PR")
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

func (m *HydrateOrchestrator) AutomergeFileExists(ctx context.Context, globPattern string) bool {
	fmt.Println("Glob pattern")
	fmt.Println(globPattern)
	entries, err := m.ValuesStateDir.Glob(ctx, globPattern+"/*")

	if err != nil {

		panic(err)
	}

	automergeFileFound := false

	for _, entry := range entries {
		fmt.Println("Entry")
		fmt.Println(entry)
		if fmt.Sprintf("%s/%s", globPattern, "AUTO_MERGE") == entry {

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

	command := strings.Join([]string{"pr", "merge", prLink}, " ")

	_, err := dag.Gh().Run(command, dagger.GhRunOpts{
		Version:      m.GhCliVersion,
		Token:        m.GhToken,
		DisableCache: true,
	}).Sync(ctx)

	log.Printf("PR %s merged successfully", prLink)

	if err != nil {
		return err
	}

	return nil

}
