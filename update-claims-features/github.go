package main

import (
	"context"
	"dagger/update-claims-features/internal/dagger"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/xeipuuv/gojsonschema"
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

	command := strings.Join([]string{"pr", "merge", "--merge", prLink, "--auto"}, " ")

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

// getReleasesForRepo fetches releases for the given repo (owner/repo) using the
// provided token. features slice should contain feature name prefixes to query
// for (can be empty only if calling code expects all releases).
func (m *UpdateClaimsFeatures) getReleasesForRepo(ctx context.Context, repo string, features []string, token *dagger.Secret) (string, error) {
	ghReleaseListResult := ""
	var err error
	// validate repo format
	parts := strings.SplitN(strings.TrimSpace(repo), "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", fmt.Errorf("invalid repo %q, expected owner/repo", repo)
	}
	owner := parts[0]
	name := parts[1]

	cmd := []string{
		"gh",
		"api",
		"graphql",
		"-F",
		fmt.Sprintf("owner=%s", owner),
		"-F",
		fmt.Sprintf("name=%s", name),
	}

	queryNameTemplate := "feature_query_%d"
	queryVarTemplate := "feature_var_%d"
	queryVarList := "$owner: String!, $name: String!"
	currentQueryIndex := 0

	if len(features) > 0 {
		query := `query GetReleases(%s) {
  repository(owner: $owner, name: $name) {
    %s
  }
}`
		featureQuery := ""

		for _, feature := range features {
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
			return "", fmt.Errorf("no valid features to update specified for repo %s", repo)
		}

		query = fmt.Sprintf(query, queryVarList, featureQuery)
		cmd = append(
			cmd,
			"-f",
			fmt.Sprintf("query=%s", query),
			"--jq",
			".data.repository.[].nodes[].name",
		)

		ghOpts := dagger.GhOpts{Version: m.GhCliVersion}
		ctrOpts := dagger.GhContainerOpts{Token: token, Repo: repo}

		ghReleaseListResult, err = dag.Gh(ghOpts).Container(ctrOpts).
			WithMountedDirectory(m.ClaimsDirPath, m.ClaimsDir).
			WithWorkdir(m.ClaimsDirPath).
			WithEnvVariable("CACHE_BUSTER", time.Now().String()).
			WithExec(cmd).
			Stdout(ctx)
	} else {
		return "", fmt.Errorf("no features to update specified for repo %s", repo)
	}

	return ghReleaseListResult, err
}

func (m *UpdateClaimsFeatures) workflowRun(
	ctx context.Context,
	claimName string,
) (string, error) {
	workflowName := "hydrate-github-claim.yaml"

	ctr := dag.Gh(dagger.GhOpts{
		Version: m.GhCliVersion,
	}).Container(dagger.GhContainerOpts{
		Token: m.GhToken,
	})

	dispatchedAt := time.Now().UTC()
	_, err := ctr.
		WithEnvVariable("CACHE_BUSTER", time.Now().String()).
		WithExec([]string{
			"gh", "workflow", "run",
			"-R", m.Repo,
			workflowName,
			"-f", fmt.Sprintf("name=%s", claimName),
			"-f", "kind=ComponentClaim",
		}).
		Sync(ctx)

	if err != nil {
		return "", err
	}

	type workflowRun struct {
		URL          string    `json:"url"`
		CreatedAt    time.Time `json:"createdAt"`
		Event        string    `json:"event"`
		DisplayTitle string    `json:"displayTitle"`
	}

	const maxAttempts = 8
	const initialPollInterval = 3 * time.Second
	normalizedClaimName := strings.ToLower(strings.TrimSpace(claimName))
	for attempt := 0; attempt < maxAttempts; attempt++ {
		runsJSON, err := ctr.
			WithEnvVariable("BUST_CACHE", time.Now().String()).
			WithExec([]string{
				"gh", "run", "list",
				"-R", m.Repo,
				"--workflow", workflowName,
				"--limit", "20",
				"--json", "url,createdAt,event,displayTitle",
			}).
			Stdout(ctx)
		if err != nil {
			errMsg := extractErrorMessage(err)
			return "", errors.New(errMsg)
		}
		var runs []workflowRun
		if err := json.Unmarshal([]byte(strings.TrimSpace(runsJSON)), &runs); err != nil {
			return "", fmt.Errorf("failed to parse workflow runs for %s: %w", workflowName, err)
		}
		for _, run := range runs {
			if run.Event != "workflow_dispatch" {
				continue
			}
			if run.CreatedAt.Before(dispatchedAt) {
				continue
			}
			if strings.TrimSpace(run.URL) == "" {
				continue
			}
			if normalizedClaimName != "" && !strings.Contains(strings.ToLower(run.DisplayTitle), normalizedClaimName) {
				continue
			}
			return strings.TrimSpace(run.URL), nil
		}
		if attempt < maxAttempts-1 {
			time.Sleep(initialPollInterval * time.Duration(attempt+1))
		}
	}
	return "", fmt.Errorf(
		"failed to find workflow run URL for %s after dispatching %q",
		workflowName,
		claimName,
	)
}

func (m *UpdateClaimsFeatures) getAllValidationSchemas(
	ctx context.Context,
) (*gojsonschema.SchemaLoader, error) {
	ctr, err := dag.Gh(dagger.GhOpts{
		Version: m.GhCliVersion,
	}).Container(dagger.GhContainerOpts{
		Token: m.PrefappGhToken,
		Repo:  "prefapp/features",
	}).WithMountedDirectory(m.ClaimsDirPath, m.ClaimsDir).
		WithWorkdir(m.ClaimsDirPath).
		WithEnvVariable("CACHE_BUSTER", time.Now().String()).
		WithExec([]string{"apk", "add", "curl"}).
		WithExec([]string{
			"curl",
			"--fail",
			"--location",
			"--output",
			"/tmp/schema.json",
			"https://raw.githubusercontent.com/firestartr-pro/docs/main/site/raw/core/claims/claims.schema.json",
		}).
		Sync(ctx)
	if err != nil {
		return nil, err
	}

	schemaContent, err := ctr.File("/tmp/schema.json").Contents(ctx)
	if err != nil {
		return nil, err
	}

	var schemas []interface{}
	if err := json.Unmarshal([]byte(schemaContent), &schemas); err != nil {
		return nil, err
	}

	sl := gojsonschema.NewSchemaLoader()
	err = loadSchemaList(schemas, sl, 0)
	if err != nil {
		return nil, err
	}

	return sl, nil
}

func (m *UpdateClaimsFeatures) getComponentValidationSchema(
	ctx context.Context,
) (*gojsonschema.Schema, error) {
	targetID := "firestartr.dev://common/ComponentClaim"

	schemaLoader, err := m.getAllValidationSchemas(ctx)
	if err != nil {
		return nil, err
	}

	compiledSchema, err := schemaLoader.Compile(
		gojsonschema.NewReferenceLoader(targetID),
	)
	if err != nil {
		return nil, err
	}

	return compiledSchema, nil
}

var releasesChangelog = make(map[string]string)

// getReleaseChangelog fetches the changelog for a release tag in the given
// repo using the provided token. Caches results keyed by "repo|tag".
func (m *UpdateClaimsFeatures) getReleaseChangelog(
	ctx context.Context,
	releaseTag string,
	repo string,
	token *dagger.Secret,
) (string, error) {
	cacheKey := fmt.Sprintf("%s|%s", repo, releaseTag)
	if releasesChangelog[cacheKey] != "" {
		fmt.Printf("☢️ Using cached changelog for %s\n", cacheKey)
		return releasesChangelog[cacheKey], nil
	}

	// validate repo
	parts := strings.SplitN(strings.TrimSpace(repo), "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", fmt.Errorf("invalid repo %q, expected owner/repo", repo)
	}

	if token == nil {
		return "", fmt.Errorf("no token provided for repo %q", repo)
	}

	fmt.Printf("☢️ No cached changelog for %s found, getting it from GitHub\n", cacheKey)
	changelog, err := dag.Gh(dagger.GhOpts{
		Version: m.GhCliVersion,
	}).Container(dagger.GhContainerOpts{
		Token: token,
		Repo:  repo,
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

	if err != nil {
		return "", err
	}

	releasesChangelog[cacheKey] = changelog
	return changelog, nil
}
