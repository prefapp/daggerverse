package main

import (
	"bytes"
	"context"
	"dagger/update-claims-features/internal/dagger"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"
	"gopkg.in/yaml.v3"
)

func (m *UpdateClaimsFeatures) getLatestReleasesAsMap(
	ctx context.Context,
	ghReleaseListResult string,
) (map[string]string, error) {
	var featuresMap = make(map[string]string)
	var releasesList []ReleasesList
	err := json.Unmarshal([]byte(ghReleaseListResult), &releasesList)

	if err != nil {
		return nil, err
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
		currentVersion, hasVersion := featuresMap[featureTag]
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
			return nil, err
		}

		if versionIsGreater.Check(featureVersionSemver) {
			featuresMap[featureTag] = featureVersion
		}
	}

	return featuresMap, nil
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
) (string, error) {
	releaseBody := ""
	var parsedJson ReleaseBody

	for _, feature := range featureList {
		fullFeatureTag := fmt.Sprintf("%s-v%s", feature.Name, feature.Version)
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

		releaseBody = fmt.Sprintf(
			"##%s:\n%s\n\n\n%s",
			feature.Name,
			parsedJson.Body,
			releaseBody,
		)
	}

	return releaseBody, nil
}
