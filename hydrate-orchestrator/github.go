package main

import (
	"context"
	"dagger/hydrate-orchestrator/internal/dagger"
	"fmt"
	"strings"
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
	// Base branch of the PR
	// +optional
	baseBranch string,

) (string, error) {
	labelColors := []string{}

	for _, label := range labels {
		labelColors = append(labelColors, m.getColorForLabel(label))
	}

	return dag.Gh().CommitAndCreatePr(
		ctx,
		contents,
		newBranchName,
		"Update deployments",
		title,
		body,
		dagger.GhCommitAndCreatePrOpts{
			BaseBranch:  baseBranch,
			Version:     m.GhCliVersion,
			Token:       m.GhToken,
			Labels:      labels,
			LabelColors: labelColors,
			Reviewers:   reviewers,
			DeletePath:  cleanupDir,
		},
	)
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
	case strings.Contains(label, "tfworkspace/"):
		return "7B42BC"
	default:
		return "7E7C7A"
	}
}

func (m *HydrateOrchestrator) AutomergeFileExists(ctx context.Context, globPattern string) bool {

	entries, err := m.ValuesStateDir.Glob(ctx, globPattern+"/*")

	if err != nil {

		panic(err)
	}

	automergeFileFound := false

	for _, entry := range entries {

		if fmt.Sprintf("%s/%s", globPattern, "AUTO_MERGE") == entry {

			fmt.Printf("☢️ Automerge file found: %s\n", entry)

			automergeFileFound = true

			break
		}
	}

	fmt.Printf("☢️ Automerge file not found\n")

	return automergeFileFound

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

	fmt.Printf("☢️ PR %s merged successfully\n", prLink)

	return nil
}
