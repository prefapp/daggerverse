package main

import (
	"context"
	"dagger/update-claims-features/internal/dagger"
	"fmt"
	"regexp"
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

	// Path to the local GitHub CLI binary file (not a directory)
	// +optional
	localGhCliPath *dagger.File,
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
		Repo:              repo,
		Org:               strings.Split(repo, "/")[0],
		GhToken:           ghToken,
		PrefappGhToken:    prefappGhToken,
		GhCliVersion:      ghCliVersion,
		ClaimsDir:         claimsDir,
		ClaimsDirPath:     claimsDirPath,
		DefaultBranch:     defaultBranch,
		ClaimsToUpdate:    claimsToUpdateList,
		FeaturesToUpdate:  featuresToUpdateList,
		VersionConstraint: versionConstraint,
		Automerge:         automerge,
		LocalGhCliPath:    localGhCliPath,
	}, nil
}

func (m *UpdateClaimsFeatures) UpdateAllClaimFeatures(
	ctx context.Context,
) (string, error) {
	ghReleaseListResult, err := m.getReleases(ctx)
	if err != nil {
		return "", err
	}

	latestFeaturesMap, allFeaturesMap, err := m.getFeaturesMapData(
		ctx, ghReleaseListResult,
	)
	if err != nil {
		return "", err
	}

	// Get all ComponentClaim claims
	claims, err := m.getAllClaims(ctx)
	if err != nil {
		return "", err
	}

	for _, entry := range claims {
		fmt.Printf("Reading claim %s\n", entry)

		claim, err := m.getClaimIfKindComponent(ctx, entry)
		if err != nil {
			return "", err
		}

		if claim != nil {
			updatedFeaturesList, createPR, err := m.updateClaimFeatures(
				ctx,
				claim,
				latestFeaturesMap,
			)
			if err != nil {
				return "", err
			}

			if createPR {
				currentFeatureVersionsMap := m.extractCurrentFeatureVersionsFromClaim(
					ctx, claim,
				)
				claim.Providers.Github.Features = updatedFeaturesList
				updatedDir := m.updateDirWithClaim(ctx, claim, entry)
				releaseBody, err := m.getPrBodyForFeatureUpdate(
					ctx,
					updatedFeaturesList,
					allFeaturesMap,
					currentFeatureVersionsMap,
				)
				if err != nil {
					return "", err
				}

				prLink, err := m.upsertPR(
					ctx,
					fmt.Sprintf("update-%s-%s", claim.Name, claim.Kind),
					updatedDir,
					[]string{},
					fmt.Sprintf("Update %s features to latest version", claim.Name),
					releaseBody,
				)

				if err != nil {
					return "", err
				}

				fmt.Printf("PR LINK: %s\n", prLink)

				if m.Automerge {
					m.MergePullRequest(ctx, prLink)
				}
			}
		}
	}

	return "ok", nil
}
