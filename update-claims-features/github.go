package main

import (
	"context"
	"dagger/update-claims-features/internal/dagger"
	"errors"
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
) (string, error) {
	return dag.Gh().CommitAndCreatePr(
		ctx,
		contents,
		newBranchName,
		"ci: Update claims' features",
		title,
		body,
		dagger.GhCommitAndCreatePrOpts{
			Version:        m.GhCliVersion,
			Token:          m.GhToken,
			Labels:         labels,
			LocalGhCliPath: m.LocalGhCliPath,
		},
	)
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

func (m *UpdateClaimsFeatures) getReleases(ctx context.Context) (string, error) {
	ghReleaseListResult := ""
	var err error
	cmd := []string{
		"gh",
		"api",
		"graphql",
		"-F",
		"owner=prefapp",
		"-F",
		"name=features",
	}

	queryNameTemplate := "feature_query_%d"
	queryVarTemplate := "feature_var_%d"
	queryVarList := "$owner: String!, $name: String!"
	currentQueryIndex := 0

	if len(m.FeaturesToUpdate) > 0 {
		query := `query GetReleases(%s) {
  repository(owner: $owner, name: $name) {
	%s
  }
}`
		featureQuery := ""

		for _, feature := range m.FeaturesToUpdate {
			if feature == "" {
				continue
			}

			varName := fmt.Sprintf(queryVarTemplate, currentQueryIndex)

			featureQuery = fmt.Sprintf(`%s
%s: refs(refPrefix: "refs/tags/", last: 100, query: $%s) {
  nodes {
	name
  }
}`, featureQuery, fmt.Sprintf(queryNameTemplate, currentQueryIndex), varName)

			queryVarList = fmt.Sprintf("%s, $%s: String!", queryVarList, varName)
			cmd = append(cmd, "-F", fmt.Sprintf("%s=%s-", varName, feature))

			currentQueryIndex++
		}

		if featureQuery == "" {
			return "", fmt.Errorf("no valid features to update specified")
		}

		query = fmt.Sprintf(query, queryVarList, featureQuery)
		cmd = append(
			cmd,
			"-f",
			fmt.Sprintf("query=%s", query),
			"--jq",
			".data.repository.[].nodes[].name",
		)

		ghReleaseListResult, err = dag.Gh(dagger.GhOpts{
			Version: m.GhCliVersion,
		}).Container(dagger.GhContainerOpts{
			Token: m.PrefappGhToken,
			Repo:  "prefapp/features",
		}).WithMountedDirectory(m.ClaimsDirPath, m.ClaimsDir).
			WithWorkdir(m.ClaimsDirPath).
			WithEnvVariable("CACHE_BUSTER", time.Now().String()).
			WithExec(cmd).
			Stdout(ctx)
	} else {
		return "", fmt.Errorf("no features to update specified")
	}

	return ghReleaseListResult, err
}

func (m *UpdateClaimsFeatures) WorkflowRun(
	ctx context.Context,
	claimName string,
) error {
	ctr := dag.Gh(dagger.GhOpts{
		Version: m.GhCliVersion,
	}).Container(dagger.GhContainerOpts{
		Token: m.GhToken,
	})

	_, err := ctr.
		WithEnvVariable("BUST_CACHE", time.Now().String()).
		WithExec([]string{
			"gh", "workflow", "run",
			"-R", fmt.Sprintf("%s/%s", m.Org, m.Repo),
			"hydrate-github-claim.yaml",
			"-f", fmt.Sprintf("claim=%s", claimName),
			"-f", "kind=ComponentClaim",
		}).
		Sync(ctx)

	if err != nil {
		errMsg := extractErrorMessage(err)
		return errors.New(errMsg)
	}

	return nil
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
