package main

import (
	"context"
	"dagger/update-claims-features/internal/dagger"
	"fmt"
	"time"
)

type UpdateClaimsFeatures struct{}

func (m *UpdateClaimsFeatures) UpdateAllClaimFeatures(
	ctx context.Context,
	claimsDir *dagger.Directory,

	// Claims dir path
	// +required
	claimsDirPath string,

	// GitHub token
	// +required
	ghToken *dagger.Secret,

	//Gh CLI Version
	// +optional
	// +default="v2.66.1"
	ghCliVersion string,
) (string, error) {
	// Get latest feature version
	featuresList, err := dag.Gh(dagger.GhOpts{
		Version: ghCliVersion,
	}).Container(dagger.GhContainerOpts{
		Token: ghToken,
	}).WithDirectory(claimsDirPath, claimsDir, dagger.ContainerWithDirectoryOpts{}).
		WithWorkdir(claimsDirPath).
		WithEnvVariable("CACHE_BUSTER", time.Now().String()).
		WithExec([]string{
			"gh",
			"release",
			"list",
			// "--exclude-drafts",
			// "--exclude-pre-releases",
		}).
		Stdout(ctx)

	if err != nil {
		return "", err
	}

	// Get all ComponentClaim claims

	// Update individually, create PR

	fmt.Printf("Features list: %s", featuresList)

	return featuresList, nil
}
