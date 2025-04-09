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

func (m *UpdateClaimsFeatures) getFeaturesMapData(
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

		var versionIsValid *semver.Constraints
		if m.VersionConstraint != "" {
			versionIsValid, err = semver.NewConstraint(m.VersionConstraint)
			if err != nil {
				return nil, nil, err
			}
		} else {
			versionToCompareTo := "0.0.0"
			currentVersion, hasVersion := latestFeaturesMap[featureTag]
			if hasVersion {
				versionToCompareTo = currentVersion
			}

			versionIsValid, err = semver.NewConstraint(
				fmt.Sprintf("> %s", versionToCompareTo),
			)
			if err != nil {
				return nil, nil, err
			}
		}

		if versionIsValid.Check(featureVersionSemver) {
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

func (m *UpdateClaimsFeatures) getPrBodyForFeatureUpdate(
	ctx context.Context,
	updatedFeaturesList []Feature,
	allFeaturesMap map[string][]string,
	originalVersionMap map[string]string,
) (string, error) {
	prBody := ""
	var parsedJson ReleaseBody

	for _, updatedFeature := range updatedFeaturesList {
		updatedFeatureVersionSemver, err := semver.NewVersion(updatedFeature.Version)

		versionIsDifferentThanOriginal, err := semver.NewConstraint(
			fmt.Sprintf("!= %s", originalVersionMap[updatedFeature.Name]),
		)
		if err != nil {
			return "", err
		}

		// Unupdated features are still added to the updatedFeaturesList with
		// the same version as they originally had, so we filter them here
		// (they are added so they don't get deleted when updating the feature list)
		if versionIsDifferentThanOriginal.Check(updatedFeatureVersionSemver) {
			prBody = fmt.Sprintf("%s## %s:", prBody, updatedFeature.Name)
			for _, featureVersion := range allFeaturesMap[updatedFeature.Name] {
				featureVersionSemver, err := semver.NewVersion(featureVersion)
				if err != nil {
					return "", err
				}

				addChangeLog, err := semver.NewConstraint(
					fmt.Sprintf(
						"> %s, <= %s || =%s",
						originalVersionMap[updatedFeature.Name],
						updatedFeature.Version,
						updatedFeature.Version,
					),
				)
				if err != nil {
					return "", err
				}

				// allFeaturesMap contains every release for every feature, so
				// they are filtered here so only the changelogs for versions
				// that are greater than the originally installed one but
				// lesser or equal to the version that is being currently
				// installed (which won't necessarily be latest)
				versionInfo := ""
				if addChangeLog.Check(featureVersionSemver) {
					fullFeatureTag := fmt.Sprintf(
						"%s-v%s", updatedFeature.Name, featureVersion,
					)
					changelog, err := m.getReleaseChangelog(ctx, fullFeatureTag)

					if err != nil {
						return "", err
					}

					err = json.Unmarshal([]byte(changelog), &parsedJson)
					if err != nil {
						return "", err
					}

					versionInfo = fmt.Sprintf(
						"%s\n%s",
						versionInfo,
						parsedJson.Body,
					)
				}

				prBody = fmt.Sprintf(
					"%s\n%s\n\n\n",
					prBody,
					versionInfo,
				)
			}
		}
	}

	return prBody, nil
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
