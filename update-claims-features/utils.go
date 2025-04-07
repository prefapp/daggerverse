package main

import (
	"bytes"
	"context"
	"dagger/update-claims-features/internal/dagger"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"gopkg.in/yaml.v3"
)

func (m *UpdateClaimsFeatures) getLatestReleasesAsMap(
	ctx context.Context,
	ghReleaseListResult string,
) (map[string]string, map[string][]string, error) {
	var latestFeaturesMap = make(map[string]string)
	var allFeaturesMap = make(map[string][]string)
	var sortedFeaturesMap = make(map[string][]*semver.Version)
	var releasesList []ReleasesList
	err := json.Unmarshal([]byte(ghReleaseListResult), &releasesList)

	if err != nil {
		return nil, nil, err
	}

	for _, feature := range releasesList {
		featureData := strings.Split(feature.TagName, "-")

		featureTag := featureData[0]
		featureVersion := strings.Trim(featureData[1], "v")
		featureVersionSemver, err := semver.NewVersion(featureData[1])
		if err != nil {
			fmt.Printf(
				"Version %s of feature %s is not valid SemVer, skipping",
				featureData[1],
				featureTag,
			)
			continue
		}

		versionToCompareTo := "0.0.0"
		currentVersion, hasVersion := latestFeaturesMap[featureTag]
		if hasVersion {
			versionToCompareTo = currentVersion
		}

		versionConstraint := fmt.Sprintf("> %s", versionToCompareTo)
		if m.VersionConstraint != "" {
			versionConstraint = fmt.Sprintf(
				"%s, %s", versionToCompareTo, m.VersionConstraint,
			)
		}

		versionIsGreater, err := semver.NewConstraint(versionConstraint)
		if err != nil {
			return nil, nil, err
		}

		if versionIsGreater.Check(featureVersionSemver) {
			latestFeaturesMap[featureTag] = featureVersion
		}

		if sortedFeaturesMap[featureTag] == nil {
			sortedFeaturesMap[featureTag] = []*semver.Version{}
		}

		sortedFeaturesMap[featureTag] = append(
			sortedFeaturesMap[featureTag], featureVersionSemver,
		)
	}

	// Sort map
	for key, _ := range sortedFeaturesMap {
		sort.Sort(semver.Collection(sortedFeaturesMap[key]))
	}

	for featureName, featuresList := range sortedFeaturesMap {
		allFeaturesMap[featureName] = []string{}
		for _, feature := range featuresList {
			allFeaturesMap[featureName] = append(
				allFeaturesMap[featureName], fmt.Sprintf("%s", feature),
			)
		}
	}

	return latestFeaturesMap, allFeaturesMap, nil
}

func (m *UpdateClaimsFeatures) updateDirWithClaim(
	ctx context.Context,
	claim *Claim,
	claimPath string,
) *dagger.Directory {
	var buffer bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&buffer)
	yamlEncoder.SetIndent(2)
	yamlEncoder.Encode(&claim)

	updatedDir := m.ClaimsDir.WithNewFile(claimPath, string(buffer.Bytes()))

	return updatedDir
}

func (m *UpdateClaimsFeatures) getReleaseBodyForFeatureList(
	ctx context.Context,
	featureList []Feature,
	allFeaturesMap map[string][]string,
	originalVersionMap map[string]string,
) (string, error) {
	releaseBody := ""
	var parsedJson ReleaseBody

	for _, feature := range featureList {
		releaseBody = fmt.Sprintf("%s## %s:\n", releaseBody, feature.Name)
		versionConstraint := fmt.Sprintf("> %s", originalVersionMap[feature.Name])
		currentFeatureVersionSemver, err := semver.NewVersion(feature.Version)

		versionIsGreater, err := semver.NewConstraint(versionConstraint)
		if err != nil {
			return "", err
		}

		if versionIsGreater.Check(currentFeatureVersionSemver) {
			for _, featureVersion := range allFeaturesMap[feature.Name] {
				featureVersionSemver, err := semver.NewVersion(featureVersion)
				if err != nil {
					return "", err
				}

				versionInfo := ""
				if versionIsGreater.Check(featureVersionSemver) {
					fullFeatureTag := fmt.Sprintf("%s-v%s", feature.Name, featureVersion)
					changelog, err := m.getReleaseChangelog(
						ctx,
						fullFeatureTag,
					)

					if err != nil {
						return "", err
					}

					err = json.Unmarshal([]byte(changelog), &parsedJson)
					if err != nil {
						return "", err
					}

					versionInfo = fmt.Sprintf(
						"%s\n\n\n%s",
						versionInfo,
						parsedJson.Body,
					)
				}

				releaseBody = fmt.Sprintf(
					"%s\n\n\n%s",
					releaseBody,
					versionInfo,
				)
			}
		}
	}

	return releaseBody, nil
}

func (m *UpdateClaimsFeatures) extractCurrentFeatureVersionsFromClaim(
	ctx context.Context,
	claim *Claim,
) map[string]string {
	var currentFeaturesVersion = make(map[string]string)

	for _, featureData := range claim.Providers.Github.Features {
		currentFeaturesVersion[featureData.Name] = featureData.Version
	}

	return currentFeaturesVersion
}
