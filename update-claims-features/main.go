package main

import (
	"context"
	"dagger/update-claims-features/internal/dagger"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"gopkg.in/yaml.v3"
)

type UpdateClaimsFeatures struct {
	Repo                 string
	GhToken              *dagger.Secret
	GhCliVersion         string
	ClaimsDirPath        string
	ClaimsDir            *dagger.Directory
	DefaultBranch        string
	ComponentsFolderName string
}

type ReleasesList struct {
	Name string `json:"name"`
}

type Claim struct {
	Kind      string    `yaml:"kind"`
	Version   string    `yaml:"version"`
	Type      string    `yaml:"type"`
	Lifecycle string    `yaml:"lifecycle"`
	System    string    `yaml:"system"`
	Name      string    `yaml:"name"`
	Providers Providers `yaml:"providers"`
}

type Providers struct {
	Github Github `yaml:"github"`
}

type Github struct {
	Description    string         `yaml:"description"`
	Name           string         `yaml:"name"`
	Org            string         `yaml:"org"`
	Visibility     string         `yaml:"visibility"`
	BranchStrategy BranchStrategy `yaml:"branchStrategy"`
	Actions        Actions        `yaml:"actions"`
	Features       []Feature      `yaml:"features"`
}

type BranchStrategy struct {
	Name          string `yaml:"name"`
	DefaultBranch string `yaml:"defaultBranch"`
}

type Actions struct {
	Oidc OIDC `yaml:"oidc"`
}

type OIDC struct {
	UseDefault       bool     `yaml:"useDefault"`
	IncludeClaimKeys []string `yaml:"includeClaimKeys"`
}

type Feature struct {
	Name    string            `yaml:"name"`
	Version string            `yaml:"version"`
	Args    map[string]string `yaml:"args"`
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

	// Name of the components folder name
	// +optional
	// +default="components"
	componentsFolderName string,
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
	ghReleaseListResult, err := dag.Gh(dagger.GhOpts{
		Version: m.GhCliVersion,
	}).Container(dagger.GhContainerOpts{
		Token: m.GhToken,
		Repo:  "prefapp/features",
	}).WithDirectory(m.ClaimsDirPath, m.ClaimsDir, dagger.ContainerWithDirectoryOpts{}).
		WithWorkdir(m.ClaimsDirPath).
		WithEnvVariable("CACHE_BUSTER", time.Now().String()).
		WithExec([]string{
			"gh",
			"release",
			"list",
			"--json",
			"name",
		}).
		Stdout(ctx)

	if err != nil {
		return "", err
	}

	var featuresMap = make(map[string]string)
	var releasesList []ReleasesList
	err = json.Unmarshal([]byte(ghReleaseListResult), &releasesList)

	if err != nil {
		return "", err
	}

	for _, feature := range releasesList {
		featureData := strings.Split(feature.Name, " ")

		featureName := strings.Trim(featureData[0], ":")
		featureVersion := strings.Trim(featureData[1], "v")
		featureVersionSemver, err := semver.NewVersion(featureData[1])

		if err != nil {
			return "", err
		}

		currentVersion, hasVersion := featuresMap[featureName]

		if hasVersion {
			versionIsGreater, err := semver.NewConstraint(fmt.Sprintf("> %s", currentVersion))

			if err != nil {
				return "", err
			}

			if versionIsGreater.Check(featureVersionSemver) {
				featuresMap[featureName] = featureVersion
			}
		} else {
			featuresMap[featureName] = featureVersion
		}
	}

	fmt.Printf("FEATURE LIST-----------------------------------------")
	for k, v := range featuresMap {
		fmt.Printf("FEATURE INFO>>>>>>>>>>>>>>%s, %s", k, v)
	}

	// Get all ComponentClaim claims
	var claims []string

	for _, ext := range []string{".yml", ".yaml"} {
		extClaims, err := m.ClaimsDir.Glob(
			ctx,
			fmt.Sprintf("claims/%s/*%s", m.ComponentsFolderName, ext),
		)

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

			claim.Providers.Github.Features = updatedFeaturesList

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

	return "ok", nil
}
