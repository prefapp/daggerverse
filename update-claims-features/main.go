package main

import (
	"context"
	"dagger/update-claims-features/internal/dagger"
	"fmt"
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
	// +default="v2.66.1"
	ghCliVersion string,

	// Claims repo name
	// +required
	repo string,

	// Name of the default branch of the claims repo
	// +optional
	// +default="main"
	defaultBranch string,

	// Name of the components folder name
	// +optional
	// +default="components"
	componentsFolderName string,
) (*UpdateClaimsFeatures, error) {
	return &UpdateClaimsFeatures{
		Repo:                 repo,
		GhToken:              ghToken,
		PrefappGhToken:       prefappGhToken,
		GhCliVersion:         ghCliVersion,
		ClaimsDir:            claimsDir,
		ClaimsDirPath:        claimsDirPath,
		DefaultBranch:        defaultBranch,
		ComponentsFolderName: componentsFolderName,
	}, nil
}

func (m *UpdateClaimsFeatures) UpdateAllClaimFeatures(
	ctx context.Context,
) (string, error) {
	ghReleaseListResult, err := m.getReleases(ctx)
	if err != nil {
		return "", err
	}

	featuresMap, err := m.getLatestReleasesAsMap(ctx, ghReleaseListResult)
	if err != nil {
		return "", err
	}

	// Get all ComponentClaim claims
	claims, err := m.getAllClaims(ctx)
	if err != nil {
		return "", err
	}

	for _, entry := range claims {
		fmt.Printf("Classifying claims in %s\n", entry)

		claim, err := m.getClaimIfKindComponent(ctx, entry)
		if err != nil {
			return "", err
		}

		if claim != nil {
			updatedFeaturesList, createPR, err := m.updateClaimFeatures(
				ctx,
				claim,
				featuresMap,
			)
			if err != nil {
				return "", err
			}

			if createPR {
				claim.Providers.Github.Features = updatedFeaturesList

				updatedDir := m.updateDirWithClaim(ctx, claim, entry)

				prLink, err := m.upsertPR(
					ctx,
					fmt.Sprintf("update-%s-%s", claim.Name, claim.Kind),
					updatedDir,
					[]string{},
					fmt.Sprintf("Update %s features to latest version", claim.Name),
					fmt.Sprintf("Update %s features to latest version", claim.Name),
					fmt.Sprintf("kubernetes"),
					[]string{},
				)

				if err != nil {
					return "", err
				}

				fmt.Printf("PR LINK: %s", prLink)
			}

		}

	}

	return "ok", nil
}
