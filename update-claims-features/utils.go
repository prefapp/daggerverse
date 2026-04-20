package main

import (
	"bytes"
	"context"
	"dagger/update-claims-features/internal/dagger"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

var fullSemverRegex = regexp.MustCompile(
	`^v?(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$`,
)

func extractErrorMessage(err error) string {
	switch e := err.(type) {
	case *dagger.ExecError:
		errorMsg := ""

		if e.Stderr != "" {
			errorMsg += fmt.Sprintf("::error::%s\n", e.Stderr)
		}
		if e.Stdout != "" {
			errorMsg += fmt.Sprintf("::info::%s", e.Stdout)
		}

		return errorMsg
	default:
		return fmt.Sprintf("::error::%s", strings.ReplaceAll(err.Error(), "::error::", ""))
	}
}

func loadSchemaList(
	schemasList []interface{},
	schemaLoader gojsonschema.SchemaLoader,
	currentCall int,
) error {
	currentCall++
	if currentCall > 10 {
		return fmt.Errorf("too many recursive calls to loadSchemaList, possible circular reference in schemas")
	}

	for _, schema := range schemasList {
		_, isArray := schema.([]interface{})

		if isArray {
			err := loadSchemaList(schema.([]interface{}), schemaLoader, currentCall)
			if err != nil {
				return err
			}
		} else {
			goLoader := gojsonschema.NewGoLoader(schema)
			if err := schemaLoader.AddSchemas(goLoader); err != nil {
				return err
			}
		}
	}

	return nil
}

func validateClaimMap(
	claim map[string]interface{},
	schemaLoader *gojsonschema.SchemaLoader,
) error {
	targetID := "firestartr.dev://common/ComponentClaim"

	compiledSchema, err := schemaLoader.Compile(
		gojsonschema.NewReferenceLoader(targetID),
	)
	if err != nil {
		panic(err)
	}

	documentLoader := gojsonschema.NewGoLoader(claim)
	result, err := compiledSchema.Validate(documentLoader)

	if result.Valid() {
		return nil
	} else {
		return err
	}
}

func (m *UpdateClaimsFeatures) getFeaturesMapData(
	ghReleaseListResult string,
) (map[string]string, map[string][]string, error) {
	var latestFeaturesMap = make(map[string]string)
	var allFeaturesMap = make(map[string][]string)
	var sortedFeaturesMap = make(map[string][]*semver.Version)
	releasesList := strings.Split(ghReleaseListResult, "\n")

	for _, featureTag := range releasesList {
		if featureTag == "" {
			continue
		}

		featureData := strings.Split(featureTag, "-")

		if len(featureData) < 2 {
			fmt.Printf(
				"Feature tag %s is not valid, skipping\n",
				featureTag,
			)
			continue
		}

		featureName := strings.Join(featureData[:len(featureData)-1], "-")
		featureVersion := strings.Trim(featureData[len(featureData)-1], "v")
		featureVersionSemver, err := semver.NewVersion(
			featureData[len(featureData)-1],
		)
		if err != nil {
			fmt.Printf(
				"Version %s of feature %s is not a valid SemVer, skipping\n",
				featureData[len(featureData)-1],
				featureName,
			)
			continue
		}

		if !fullSemverRegex.MatchString(featureVersion) {
			fmt.Printf(
				"Version %s of feature %s is not a full SemVer (X.Y.Z), skipping as it's probably a rolling release tag\n",
				featureVersion,
				featureName,
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
			versionToCompareTo, hasVersion := latestFeaturesMap[featureName]
			if !hasVersion {
				versionToCompareTo = "0.0.0"
			}

			versionIsValid, err = semver.NewConstraint(
				fmt.Sprintf("> %s", versionToCompareTo),
			)
			if err != nil {
				return nil, nil, err
			}
		}

		if versionIsValid.Check(featureVersionSemver) {
			latestFeaturesMap[featureName] = featureVersion
		}

		if sortedFeaturesMap[featureName] == nil {
			sortedFeaturesMap[featureName] = []*semver.Version{}
		}

		sortedFeaturesMap[featureName] = append(
			sortedFeaturesMap[featureName], featureVersionSemver,
		)
	}

	// Sort map
	for key := range sortedFeaturesMap {
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
	claim map[string]interface{},
	claimPath string,
) *dagger.Directory {
	var buffer bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&buffer)
	yamlEncoder.SetIndent(2)
	yamlEncoder.Encode(&claim)

	updatedDir := m.ClaimsDir.WithNewFile(claimPath, buffer.String())

	return updatedDir
}

func (m *UpdateClaimsFeatures) getPrBodyForFeatureUpdate(
	ctx context.Context,
	updatedFeaturesList []map[string]string,
	allFeaturesMap map[string][]string,
	originalVersionMap map[string]string,
) (string, error) {
	prBody := ""
	var parsedJson ReleaseBody

	for _, updatedFeature := range updatedFeaturesList {
		updatedFeatureName := updatedFeature["name"]
		updatedFeatureVersion := updatedFeature["version"]
		if updatedFeatureVersion != "" {
			updatedFeatureVersionSemver, err := semver.NewVersion(updatedFeatureVersion)

			if err != nil {
				return "", err
			}

			if originalVersionMap[updatedFeatureName] != "" && updatedFeatureVersion != "" {
				versionIsDifferentThanOriginal, err := semver.NewConstraint(
					fmt.Sprintf("!=%s", originalVersionMap[updatedFeatureName]),
				)
				if err != nil {
					return "", err
				}

				// Updated features are still added to the updatedFeaturesList with
				// the same version as they originally had, so we filter them here
				// (they are added so they don't get deleted when updating the feature list)
				if versionIsDifferentThanOriginal.Check(updatedFeatureVersionSemver) {
					addChangeLog, err := semver.NewConstraint(
						fmt.Sprintf(
							"> %s, <= %s || =%s",
							originalVersionMap[updatedFeatureName],
							updatedFeatureVersion,
							updatedFeatureVersion,
						),
					)
					if err != nil {
						return "", err
					}

					for _, featureVersion := range allFeaturesMap[updatedFeatureName] {
						featureVersionSemver, err := semver.NewVersion(featureVersion)
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
								"%s-v%s", updatedFeatureName, featureVersion,
							)
							changelog, err := m.getReleaseChangelog(ctx, fullFeatureTag)

							if err != nil {
								fmt.Printf(
									"☢️ No changelog for tag %s exists, skipping\n",
									fullFeatureTag,
								)
								continue
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
		}
	}

	return prBody, nil
}

func (m *UpdateClaimsFeatures) extractCurrentFeatureVersionsFromClaim(
	claim map[string]interface{},
) map[string]string {
	var currentFeaturesVersion = make(map[string]string)
	featuresList := claim["providers"].(map[string]any)["github"].(map[string]any)["features"].([]any)

	for _, featureData := range featuresList {
		featureName := featureData.(map[string]any)["name"].(string)
		featureVersion := featureData.(map[string]any)["version"].(string)
		currentFeaturesVersion[featureName] = featureVersion
	}

	return currentFeaturesVersion
}
