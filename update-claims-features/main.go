package main

import (
	"context"
	"dagger/update-claims-features/internal/dagger"
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

type UpdateClaimsFeatures struct {
	Repo          string
	GhToken       *dagger.Secret
	GhCliVersion  string
	ClaimsDirPath string
	ClaimsDir     *dagger.Directory
	DefaultBranch string
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

func (m *UpdateClaimsFeatures) New(
	ctx context.Context,
	claimsDir *dagger.Directory,

	// Claims dir path
	// +required
	claimsDirPath string,

	// GitHub token
	// +required
	ghToken *dagger.Secret,

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
) (*UpdateClaimsFeatures, error) {
	return &UpdateClaimsFeatures{
		Repo:          repo,
		GhToken:       ghToken,
		GhCliVersion:  ghCliVersion,
		ClaimsDir:     claimsDir,
		ClaimsDirPath: claimsDirPath,
		DefaultBranch: defaultBranch,
	}, nil
}

func (m *UpdateClaimsFeatures) UpdateAllClaimFeatures(
	ctx context.Context,
) (string, error) {
	// Get latest feature version
	featuresList, err := dag.Gh(dagger.GhOpts{
		Version: m.GhCliVersion,
	}).Container(dagger.GhContainerOpts{
		Token: m.GhToken,
	}).WithDirectory(m.ClaimsDirPath, m.ClaimsDir, dagger.ContainerWithDirectoryOpts{}).
		WithWorkdir(m.ClaimsDirPath).
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

		extClaims, err := m.ClaimsDir.Glob(ctx, fmt.Sprintf("claims/*/*%s", ext))

		if err != nil {

			return "", err

		}

		claims = append(claims, extClaims...)
	}

	for _, entry := range claims {

		fmt.Printf("Classifying claims in %s\n", entry)

		file := m.ClaimsDir.File(entry)

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

			updatedDir := m.ClaimsDir.WithNewFile(entry, string(marshalledClaim))

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
