package main

import (
	"bytes"
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
	PrefappGhToken       *dagger.Secret
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
	Kind      string    `yaml:"kind,omitempty"`
	Version   string    `yaml:"version,omitempty"`
	Type      string    `yaml:"type,omitempty"`
	Lifecycle string    `yaml:"lifecycle,omitempty"`
	System    string    `yaml:"system,omitempty"`
	Name      string    `yaml:"name,omitempty"`
	Providers Providers `yaml:"providers,omitempty"`
	Owner     string    `yaml:"owner,omitempty"`
}

type Providers struct {
	Github Github `yaml:"github,omitempty"`
}

type Github struct {
	Description        string         `yaml:"description,omitempty"`
	Name               string         `yaml:"name,omitempty"`
	Org                string         `yaml:"org,omitempty"`
	Visibility         string         `yaml:"visibility,omitempty"`
	BranchStrategy     BranchStrategy `yaml:"branchStrategy,omitempty"`
	Actions            Actions        `yaml:"actions,omitempty"`
	Features           []Feature      `yaml:"features,omitempty"`
	AdditionalBranches []Branch       `yaml:"additionalBranches,omitempty"`
}

type BranchStrategy struct {
	Name          string `yaml:"name,omitempty"`
	DefaultBranch string `yaml:"defaultBranch,omitempty"`
}

type Actions struct {
	Oidc OIDC `yaml:"oidc,omitempty"`
}

type OIDC struct {
	UseDefault       bool     `yaml:"useDefault,omitempty"`
	IncludeClaimKeys []string `yaml:"includeClaimKeys,omitempty"`
}

type Feature struct {
	Name    string            `yaml:"name,omitempty"`
	Version string            `yaml:"version,omitempty"`
	Args    map[string]string `yaml:"args,omitempty"`
}

type Branch struct {
	Name   string `yaml:"name,omitempty"`
	Orphan bool   `yaml:"orphan,omitempty"`
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
	// Get latest feature version
	ghReleaseListResult, err := dag.Gh(dagger.GhOpts{
		Version: m.GhCliVersion,
	}).Container(dagger.GhContainerOpts{
		Token: m.PrefappGhToken,
		Repo:  "prefapp/features",
	}).WithDirectory(m.ClaimsDirPath, m.ClaimsDir, dagger.ContainerWithDirectoryOpts{}).
		WithWorkdir(m.ClaimsDirPath).
		WithEnvVariable("CACHE_BUSTER", time.Now().String()).
		WithExec([]string{
			"gh",
			"release",
			"list",
			"--limit",
			"999",
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
			fmt.Printf("Version %s of feature %s is not valid SemVer, skipping", featureData[1], feature.Name)
			continue
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

	var buffer bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&buffer)
	yamlEncoder.SetIndent(2)

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
		var createPR bool

		if claim.Kind == "ComponentClaim" {
			createPR = false

			for _, feature := range claim.Providers.Github.Features {
				featureVersionSemver, err := semver.NewVersion(
					featuresMap[feature.Name],
				)
				if err != nil {
					return "", err
				}

				versionIsGreater, err := semver.NewConstraint(
					fmt.Sprintf("> %s", feature.Version),
				)
				if err != nil {
					return "", err
				}

				// if instead of createPR = versionIsGreater.Check()
				// because a latter unupdated feature could override this value
				if versionIsGreater.Check(featureVersionSemver) {
					createPR = true
				}

				// Add feature whether its version is greater or not,
				// so unupdated features are not deleted
				feature.Version = featuresMap[feature.Name]
				updatedFeaturesList = append(updatedFeaturesList, feature)
			}

			if createPR {
				claim.Providers.Github.Features = updatedFeaturesList
				// marshalledClaim, err := yaml.Marshal(claim)

				// 	if err != nil {
				// 		return "", err
				// 	}

				yamlEncoder.Encode(&claim)

				updatedDir := m.ClaimsDir.WithNewFile(entry, string(buffer.Bytes()))

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

				if err != nil {
					return "", err
				}

				fmt.Printf("PR LINK: %s", prLink)
			}

		}

	}

	return "ok", nil
}
