package main

import (
	"context"
	"dagger/update-claims-features/internal/dagger"
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

type UpdateClaimsFeatures struct {
	Repo             string
	GhToken          *dagger.Secret
	DeploymentBranch string
	GhCliVersion     string
}

type Claim struct {
	Name      string    `yaml:"name"`
	Kind      string    `yaml:"kind"`
	Providers Providers `yaml:"providers"`
}

type Providers struct {
	Github Github `yaml:"github"`
}

type Github struct {
	Features []Feature `yaml:"features"`
}

type Feature struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

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
		}).
		Stdout(ctx)

	if err != nil {
		return "", err
	}

	var featuresMap map[string]string

	// Get all ComponentClaim claims
	var claims []string

	for _, ext := range []string{".yml", ".yaml"} {

		extClaims, err := claimsDir.Glob(ctx, fmt.Sprintf("*/*/*%s", ext))

		if err != nil {

			return "", err

		}

		claims = append(claims, extClaims...)
	}

	for _, entry := range claims {

		fmt.Printf("Classifying claims in %s\n", entry)

		file := claimsDir.File(entry)

		contents, err := file.Contents(ctx)

		if err != nil {

			return "", err

		}

		claim := &Claim{}

		err = yaml.Unmarshal([]byte(contents), claim)

		if err != nil {

			return "", err

		}

		var updatedFeaturesList []Feature

		if claim.Kind == "ComponentClaim" {

			for _, feature := range claim.Providers.Github.Features {

				feature.Version = featuresMap[feature.Name]

				updatedFeaturesList = append(updatedFeaturesList, feature)

			}

			marshalledClaim, err := yaml.Marshal(claim)

			if err != nil {

				return "", err

			}

			updatedDir := claimsDir.WithNewFile(entry, string(marshalledClaim))

			// create PR
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

			fmt.Printf("PR LINK: %s", prLink)

		}

	}

	fmt.Printf("Features list: %s", featuresList)

	return featuresList, nil
}
