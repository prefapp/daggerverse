package main

import (
	"context"
	"dagger/update-claims-features/internal/dagger"
	"fmt"
	"regexp"
	"slices"
	"strings"
)

func (m *UpdateClaimsFeatures) New(
	ctx context.Context,
	claimsDir *dagger.Directory,

	// Claims dir path
	// +required
	claimsDirPath string,

	// GitHub token
	// +required
	ghToken *dagger.Secret,

	// Prefapp org GitHub token
	// +required
	prefappGhToken *dagger.Secret,

	// Gh CLI Version
	// +optional
	// +default="v2.74.2"
	ghCliVersion string,

	// Claims repo name
	// +required
	repo string,

	// Name of the default branch of the claims repo
	// +optional
	// +default="main"
	defaultBranch string,

	// Name of the claim to be updated
	// +optional
	// +default=""
	claimsToUpdate string,

	// Name of the feature to be updated
	// +optional
	// +default=""
	featuresToUpdate string,

	// Check for the version we want to install
	// +optional
	// +default=""
	versionConstraint string,

	// Whether or not to automerge
	// +optional
	// +default=false
	automerge bool,

	// Path to the local GitHub CLI binary file (not a directory).
	// If not provided, the GitHub CLI will be downloaded automatically.
	// +optional
	localGhCliPath *dagger.File,
	// GitHub token for authenticating access to feature repositories
	// specified by a claim's repo field. Required when any claim uses a
	// custom features repository.
	// +optional
	customFeaturesRepoGhToken *dagger.Secret,
) (*UpdateClaimsFeatures, error) {
	var claimsToUpdateList []string = nil
	var featuresToUpdateList []string = nil
	rexp := regexp.MustCompile(`,\s+`)

	if claimsToUpdate != "" {
		claimsToUpdate = rexp.ReplaceAllString(claimsToUpdate, ",")
		claimsToUpdateList = strings.Split(claimsToUpdate, ",")
	}

	if featuresToUpdate != "" {
		featuresToUpdate = rexp.ReplaceAllString(featuresToUpdate, ",")
		featuresToUpdateList = strings.Split(featuresToUpdate, ",")
	}

	return &UpdateClaimsFeatures{
		Repo:                      repo,
		Org:                       strings.Split(repo, "/")[0],
		GhToken:                   ghToken,
		PrefappGhToken:            prefappGhToken,
		GhCliVersion:              ghCliVersion,
		ClaimsDir:                 claimsDir,
		ClaimsDirPath:             claimsDirPath,
		DefaultBranch:             defaultBranch,
		ClaimsToUpdate:            claimsToUpdateList,
		FeaturesToUpdate:          featuresToUpdateList,
		VersionConstraint:         versionConstraint,
		Automerge:                 automerge,
		LocalGhCliPath:            localGhCliPath,
		CustomFeaturesRepoGhToken: customFeaturesRepoGhToken,
	}, nil
}

func (m *UpdateClaimsFeatures) UpdateAllClaimFeatures(
	ctx context.Context,
) (*dagger.File, error) {
	errorMsg := ""
	summary := &UpdateSummary{
		Items: []UpdateSummaryRow{},
	}

	// Get all ComponentClaim claims
	claims, err := m.getAllClaims(ctx)
	if err != nil {
		return nil, err
	}

	compiledSchema, err := m.getComponentValidationSchema(ctx)
	if err != nil {
		return nil, err
	}

	claimsMap := make(map[string]loadedClaim)
	for _, entry := range claims {
		fmt.Printf("Reading claim %s\n", entry)

		claimNode, claim, err := m.getClaimIfKindComponent(ctx, entry, compiledSchema)
		if err != nil {
			summary.addUpdateSummaryRow(entry, err.Error())
			continue
		}

		if claim != nil {
			claimsMap[entry] = loadedClaim{node: claimNode, data: claim}
		}
	}

	if len(claimsMap) == 0 {
		return nil, fmt.Errorf(
			"no ComponentClaim found with name or repository name: %s",
			strings.Join(m.ClaimsToUpdate, ", "),
		)
	}

	if len(m.FeaturesToUpdate) == 0 {
		for _, loaded := range claimsMap {
			claim := loaded.data
			featuresProperty, hasFeatures := claim["providers"].(map[string]any)["github"].(map[string]any)["features"]
			if !hasFeatures {
				continue
			}

			claimFeatures := featuresProperty.([]any)
			for _, feature := range claimFeatures {
				featureName := feature.(map[string]any)["name"].(string)
				if !slices.Contains(m.FeaturesToUpdate, featureName) {
					m.FeaturesToUpdate = append(
						m.FeaturesToUpdate,
						featureName,
					)
				}
			}
		}
	}

	if len(m.FeaturesToUpdate) == 0 {
		return nil, fmt.Errorf(
			"no features found to update: no features specified and none present in claims",
		)
	}

	// Build a map of repo -> features to fetch, based on features declared in
	// claims. If a feature declares a "repo" field (owner/repo) it will be
	// fetched from there using CustomFeaturesRepoGhToken; otherwise it defaults to
	// prefapp/features.
	repoToFeatures := map[string]map[string]struct{}{}
	for _, loaded := range claimsMap {
		claim := loaded.data
		featuresProperty, hasFeatures := claim["providers"].(map[string]any)["github"].(map[string]any)["features"]
		if !hasFeatures {
			continue
		}

		claimFeatures := featuresProperty.([]any)
		for _, feature := range claimFeatures {
			fm := feature.(map[string]any)
			featureName := fm["name"].(string)
			if !slices.Contains(m.FeaturesToUpdate, featureName) {
				continue
			}
			repoStr := featureRepo(fm)

			if repoToFeatures[repoStr] == nil {
				repoToFeatures[repoStr] = map[string]struct{}{}
			}
			repoToFeatures[repoStr][featureName] = struct{}{}
		}
	}

	// For each repo, fetch releases and build maps
	repoLatestMap := map[string]map[string]string{}
	repoAllMap := map[string]map[string][]string{}
	for repoStr, featuresSet := range repoToFeatures {
		// select token
		var token *dagger.Secret
		if repoStr != defaultFeaturesRepo {
			if m.CustomFeaturesRepoGhToken == nil {
				return nil, fmt.Errorf("external repo %q present but CustomFeaturesRepoGhToken is not provided", repoStr)
			}
			token = m.CustomFeaturesRepoGhToken
		} else {
			token = m.PrefappGhToken
		}

		// convert set to slice
		featuresSlice := []string{}
		for f := range featuresSet {
			featuresSlice = append(featuresSlice, f)
		}

		ghReleaseListResult, err := m.getReleasesForRepo(ctx, repoStr, featuresSlice, token)
		if err != nil {
			return nil, err
		}

		latest, all, err := m.getFeaturesMapData(ghReleaseListResult)
		if err != nil {
			return nil, err
		}

		repoLatestMap[repoStr] = latest
		repoAllMap[repoStr] = all
	}

	// Iterate claims and resolve per-feature repo to construct claim-specific
	// latest/all maps that updateClaimFeatures expects (repo|featureName -> latest)
	for entry, loaded := range claimsMap {
		claim := loaded.data
		claimName := claim["name"].(string)
		claimKind := claim["kind"].(string)
		// Build claim-specific latest/all features maps by resolving per-feature repo
		claimLatestMap := map[string]string{}
		claimAllFeatures := map[string][]string{}
		featuresProperty, hasFeatures := claim["providers"].(map[string]any)["github"].(map[string]any)["features"]
		if hasFeatures {
			claimFeatures := featuresProperty.([]any)
			for _, feature := range claimFeatures {
				fm := feature.(map[string]any)
				name := fm["name"].(string)
				if !slices.Contains(m.FeaturesToUpdate, name) {
					continue
				}
				repoStr := featureRepo(fm)

				repoLatest, ok := repoLatestMap[repoStr]
				if !ok {
					return nil, fmt.Errorf("no release data found for repo %q required by feature %s", repoStr, name)
				}
				claimLatestMap[repoFeatureKey(repoStr, name)] = repoLatest[name]

				if repoAllMap[repoStr] != nil {
					claimAllFeatures[repoFeatureKey(repoStr, name)] = repoAllMap[repoStr][name]
				} else {
					claimAllFeatures[repoFeatureKey(repoStr, name)] = []string{}
				}
			}
		}

		updatedFeaturesList, createPR, hydrateClaim, err := m.updateClaimFeatures(
			claim,
			claimLatestMap,
		)
		if err != nil {
			summary.addUpdateSummaryRow(
				claimName, extractErrorMessage(err),
			)
			continue
		}

		if createPR {
			currentFeatureVersionsMap := m.extractCurrentFeatureVersionsFromClaim(
				claim,
			)
			patchClaimFeatureVersions(loaded.node, updatedFeaturesList)
			updatedDir := m.updateDirWithClaim(loaded.node, entry)
			releaseBody, err := m.getPrBodyForFeatureUpdate(
				ctx,
				updatedFeaturesList,
				claimAllFeatures,
				currentFeatureVersionsMap,
			)
			if err != nil {
				summary.addUpdateSummaryRow(
					claimName, extractErrorMessage(err),
				)
				continue
			}

			prLink, err := m.upsertPR(
				ctx,
				fmt.Sprintf("update-%s-%s", claimName, claimKind),
				updatedDir,
				[]string{},
				fmt.Sprintf("Update %s features to latest version", claimName),
				releaseBody,
			)

			if err != nil {
				summary.addUpdateSummaryRow(
					claimName, extractErrorMessage(err),
				)
				errorMsg = fmt.Sprintf("%s\n%s", errorMsg, extractErrorMessage(err))
				continue
			}

			summary.addUpdateSummaryRow(
				claimName,
				fmt.Sprintf("Success: <a href=\"%s\">%s</a>", prLink, prLink),
			)

			if m.Automerge {
				m.MergePullRequest(ctx, prLink)
			}
		} else {
			if hydrateClaim {
				workflowURL, err := m.workflowRun(ctx, claimName)
				if err != nil {
					summary.addUpdateSummaryRow(
						claimName, extractErrorMessage(err),
					)
					continue
				}

				summary.addUpdateSummaryRow(
					claimName,
					fmt.Sprintf(
						"Ref detected. Hydration workflow triggered: <a href=\"%s\">%s</a>",
						workflowURL, workflowURL,
					),
				)
			}
		}
	}

	var returnedError error
	if errorMsg != "" {
		returnedError = fmt.Errorf("%s", errorMsg)
	}

	return m.DeploymentSummaryToFile(ctx, summary), returnedError
}
