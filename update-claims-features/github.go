package main

import (
	"context"
	"dagger/update-claims-features/internal/dagger"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

/*
Create or update a PR with the updated contents
*/

func (m *UpdateClaimsFeatures) upsertPR(
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
	// PR author
	// +optional
	reviewers []string,

) (string, error) {
	return dag.Gh().CommitAndCreatePr(
		ctx,
		contents,
		newBranchName,
		"Update claims' features",
		title,
		body,
		dagger.GhCommitAndCreatePrOpts{
			Version: m.GhCliVersion,
			Token:   m.GhToken,
			Labels:  labels,
		})
}

func (m *UpdateClaimsFeatures) getRepoPrs(ctx context.Context) ([]Pr, error) {

	command := strings.Join([]string{
		"pr",
		"list",
		"--json", "headRefName",
		"--json", "number,url",
		"--json", "state",
		"-L", "1000",
		"-R", m.Repo,
	}, " ")

	content, err := dag.Gh().Run(
		command,
		dagger.GhRunOpts{
			Version:      m.GhCliVersion,
			DisableCache: true,
			Token:        m.GhToken,
		},
	).Stdout(ctx)

	if err != nil {

		panic(err)
	}

	prs := []Pr{}

	json.Unmarshal([]byte(content), &prs)

	return prs, nil
}

func (m *UpdateClaimsFeatures) createRemoteBranch(
	ctx context.Context,
	// Base branch name
	// +required
	gitDir *dagger.Directory,
	// New branch name
	// +required
	newBranch string,
) {
	fmt.Printf("☢️ Creating remote branch %s\n", newBranch)

	gitDirPath := "/git_dir"

	_, err := dag.Gh().Container(dagger.GhContainerOpts{Token: m.GhToken, Version: m.GhCliVersion}).
		WithMountedDirectory(gitDirPath, gitDir).
		WithWorkdir(gitDirPath).
		WithEnvVariable("CACHE_BUSTER", time.Now().String()).
		WithExec([]string{"git", "checkout", "-b", newBranch}, dagger.ContainerWithExecOpts{}).
		WithExec([]string{"git", "push", "--force", "origin", newBranch}).
		Sync(ctx)

	if err != nil {
		panic(err)
	}
}

func (m *UpdateClaimsFeatures) regenerateRemoteBranch(
	ctx context.Context,
	// Base branch name
	// +required
	gitDir *dagger.Directory,
	// New branch name
	// +required
	branchName string,
) {
	fmt.Printf("☢️ Updating remote branch %s\n", branchName)

	gitDirPath := "/git_dir"

	_, err := dag.Gh().Container(dagger.GhContainerOpts{Token: m.GhToken, Version: m.GhCliVersion}).
		WithMountedDirectory(gitDirPath, gitDir).
		WithWorkdir(gitDirPath).
		WithEnvVariable("CACHE_BUSTER", time.Now().String()).
		WithExec([]string{"git", "push", "origin", "--delete", branchName}, dagger.ContainerWithExecOpts{}).
		WithExec([]string{"git", "checkout", "-b", branchName}, dagger.ContainerWithExecOpts{}).
		WithExec([]string{"git", "push", "--force", "origin", branchName}).
		Sync(ctx)

	if err != nil {
		panic(err)
	}
}

func (m *UpdateClaimsFeatures) getColorForLabel(label string) string {
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
	case strings.Contains(label, "tfworkspace/"):
		return "7B42BC"
	case strings.Contains(label, "plan"):
		return "AAE2A0"
	default:
		return "7E7C7A"
	}
}

func (m *UpdateClaimsFeatures) MergePullRequest(ctx context.Context, prLink string) error {

	command := strings.Join([]string{"pr", "merge", "--merge", prLink}, " ")

	_, err := dag.Gh().Run(command, dagger.GhRunOpts{
		Version:      m.GhCliVersion,
		Token:        m.GhToken,
		DisableCache: true,
	}).Sync(ctx)

	if err != nil {
		return err
	}

	fmt.Printf("☢️ PR %s merged successfully\n", prLink)

	return nil
}

func (m *UpdateClaimsFeatures) prExists(ctx context.Context, branchName string) (*Pr, error) {
	fmt.Printf("☢️ Checking if PR exists for branch %s\n", branchName)
	prs, err := m.getRepoPrs(ctx)

	if err != nil {
		return nil, err
	}

	for _, pr := range prs {

		if pr.HeadRefName == branchName && strings.ToLower(pr.State) == "open" {

			fmt.Printf("☢️ PR %s already exists\n", branchName)

			return &pr, nil

		}

	}

	fmt.Printf("☢️ PR %s does not exist\n", branchName)

	return nil, nil
}

func (m *UpdateClaimsFeatures) getReleases(ctx context.Context) (string, error) {
	ghReleaseListResult, err := dag.Gh(dagger.GhOpts{
		Version: m.GhCliVersion,
	}).Container(dagger.GhContainerOpts{
		Token: m.PrefappGhToken,
		Repo:  "prefapp/features",
	}).WithMountedDirectory(m.ClaimsDirPath, m.ClaimsDir).
		WithWorkdir(m.ClaimsDirPath).
		WithEnvVariable("CACHE_BUSTER", time.Now().String()).
		WithExec([]string{
			"gh",
			"release",
			"list",
			"--exclude-pre-releases",
			"--limit",
			"999",
			"--json",
			"tagName",
		}).
		Stdout(ctx)

	return ghReleaseListResult, err
}

var releasesChangelog = make(map[string]string)

func (m *UpdateClaimsFeatures) getReleaseChangelog(
	ctx context.Context,
	releaseTag string,
) (string, error) {
	changelog := ""
	var err error

	if releasesChangelog[releaseTag] == "" {
		fmt.Printf(
			"☢️ No cached changelog for tag %s found, getting it from GitHub\n",
			releaseTag,
		)
		changelog, err = dag.Gh(dagger.GhOpts{
			Version: m.GhCliVersion,
		}).Container(dagger.GhContainerOpts{
			Token: m.PrefappGhToken,
			Repo:  "prefapp/features",
		}).WithMountedDirectory(m.ClaimsDirPath, m.ClaimsDir).
			WithWorkdir(m.ClaimsDirPath).
			WithEnvVariable("CACHE_BUSTER", time.Now().String()).
			WithExec([]string{
				"gh",
				"release",
				"view",
				releaseTag,
				"--json",
				"body",
			}).
			Stdout(ctx)
		releasesChangelog[releaseTag] = changelog
	} else {
		fmt.Printf("☢️ Using cached changelog for tag %s\n", releaseTag)
		changelog = releasesChangelog[releaseTag]
		err = nil
	}

	return changelog, err
}
