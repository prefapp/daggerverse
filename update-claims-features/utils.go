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
		featureData := strings.Split(feature.Name, " ")

		featureName := strings.Trim(featureData[0], ":")
		featureVersion := strings.Trim(featureData[1], "v")
		featureVersionSemver, err := semver.NewVersion(featureData[1])
		if err != nil {
			fmt.Printf("Version %s of feature %s is not valid SemVer, skipping", featureData[1], feature.Name)
			continue
		}

		versionToCompareTo := "0.0.0"
		currentVersion, hasVersion := featuresMap[featureName]
		if hasVersion {
			versionToCompareTo = currentVersion
		}

		versionConstraint := fmt.Sprintf("> %s", versionToCompareTo)
		if m.VersionConstraint != "" {
			versionConstraint = m.VersionConstraint
		}

		versionIsGreater, err := semver.NewConstraint(versionConstraint)
		if err != nil {
			return nil, err
		}

		if versionIsGreater.Check(featureVersionSemver) {
			featuresMap[featureName] = featureVersion
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
