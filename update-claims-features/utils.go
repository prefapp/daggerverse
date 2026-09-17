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

// defaultFeaturesRepo is the feature repository used for features that do not
// declare a "repo" field.
const defaultFeaturesRepo = "prefapp/features"

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
	schemaLoader *gojsonschema.SchemaLoader,
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
	compiledSchema *gojsonschema.Schema,
) (string, error) {
	documentLoader := gojsonschema.NewGoLoader(claim)
	result, err := compiledSchema.Validate(documentLoader)

	if err != nil {
		return "", err
	}

	if result.Valid() {
		return "", nil
	} else {
		errorMsg := "The document is not valid. errors :\n"
		for _, desc := range result.Errors() {
			errorMsg = fmt.Sprintf("%s- %s\n", errorMsg, desc)
		}

		return errorMsg, nil
	}
}

func cloneMap(originalMap map[string]interface{}) map[string]interface{} {
	clonedMap := make(map[string]interface{})
	for key, value := range originalMap {
		clonedMap[key] = value
	}
	return clonedMap
}

// nodeFindByPath walks a yaml.Node tree following mapping keys (e.g.
// "providers", "github", "features", "version") and returns the matching
// node, or nil when the path cannot be resolved.
func nodeFindByPath(node *yaml.Node, path ...string) *yaml.Node {
	if node == nil {
		return nil
	}

	current := node
	if current.Kind == yaml.DocumentNode {
		if len(current.Content) == 0 {
			return nil
		}
		current = current.Content[0]
	}

	for _, key := range path {
		if current.Kind != yaml.MappingNode {
			return nil
		}

		matched := false
		for i := 0; i < len(current.Content); i += 2 {
			keyNode := current.Content[i]
			if keyNode.Value == key {
				current = current.Content[i+1]
				matched = true
				break
			}
		}

		if !matched {
			return nil
		}
	}

	return current
}

// nodeSetValue replaces the value of a scalar yaml.Node in place, keeping the
// surrounding document structure (and therefore the original field order) and
// the original scalar style (quoting/flow) when possible.
func nodeSetValue(node *yaml.Node, value string) {
	node.Value = value
	node.Tag = "!!str"
}

// nodeToMap converts a yaml.Node document into a map[string]any so the claim
// can be inspected and modified with the existing processing logic. The node
// tree itself is left untouched so it can be re-serialized preserving order.
func nodeToMap(node *yaml.Node) (map[string]any, error) {
	var claim map[string]any
	encoded, err := yaml.Marshal(node)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(encoded, &claim)
	if err != nil {
		return nil, err
	}

	return claim, nil
}

// patchClaimFeatureVersions updates the "version" value of each feature in
// the yaml.Node tree to match its entry in updatedFeaturesList. Features are
// matched positionally because updateClaimFeatures preserves the order of the
// original features list.
func patchClaimFeatureVersions(
	claimNode *yaml.Node,
	updatedFeaturesList []map[string]any,
) {
	featuresNode := nodeFindByPath(claimNode, "providers", "github", "features")
	if featuresNode == nil || featuresNode.Kind != yaml.SequenceNode {
		return
	}

	for i, feature := range updatedFeaturesList {
		if i >= len(featuresNode.Content) {
			break
		}

		versionProperty, hasVersion := feature["version"]
		if !hasVersion {
			continue
		}

		versionNode := nodeFindByPath(featuresNode.Content[i], "version")
		if versionNode == nil {
			continue
		}

		versionValue := versionProperty.(string)
		if versionNode.Value == versionValue {
			continue
		}

		nodeSetValue(versionNode, versionValue)
	}
}

// repoFeatureKey returns the composite key identifying a feature by its
// repository and name, so features with the same name from different repos are
// not collapsed into a single entry.
func repoFeatureKey(repo, name string) string {
	return fmt.Sprintf("%s|%s", repo, name)
}

// featureRepo returns the repository a feature is pinned to, defaulting to
// defaultFeaturesRepo when it does not declare a "repo" field. The returned
// value is trimmed to avoid passing whitespace-padded identifiers to the
// GitHub CLI (via GH_REPO).
func featureRepo(feature map[string]any) string {
	if r, ok := feature["repo"]; ok {
		if rs, ok2 := r.(string); ok2 {
			trimmed := strings.TrimSpace(rs)
			if trimmed != "" {
				return trimmed
			}
		}
	}
	return defaultFeaturesRepo
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
	claimNode *yaml.Node,
	claimPath string,
) *dagger.Directory {
	var buffer bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&buffer)
	yamlEncoder.SetIndent(2)
	yamlEncoder.Encode(claimNode)

	updatedDir := m.ClaimsDir.WithNewFile(claimPath, buffer.String())

	return updatedDir
}

func (m *UpdateClaimsFeatures) getPrBodyForFeatureUpdate(
	ctx context.Context,
	updatedFeaturesList []map[string]any,
	allFeaturesMap map[string][]string,
	originalVersionMap map[string]string,
) (string, error) {
	prBody := ""
	var parsedJson ReleaseBody

	for _, updatedFeature := range updatedFeaturesList {
		updatedFeatureName := updatedFeature["name"].(string)
		versionProperty, hasVersion := updatedFeature["version"]
		if !hasVersion {
			continue
		}

		updatedFeatureVersion := versionProperty.(string)
		if updatedFeatureVersion != "" {
			updatedFeatureVersionSemver, err := semver.NewVersion(updatedFeatureVersion)

			if err != nil {
				return "", err
			}

			repoStr := featureRepo(updatedFeature)
			featureKey := repoFeatureKey(repoStr, updatedFeatureName)

			if originalVersionMap[featureKey] != "" && updatedFeatureVersion != "" {
				versionIsDifferentThanOriginal, err := semver.NewConstraint(
					fmt.Sprintf("!=%s", originalVersionMap[featureKey]),
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
							originalVersionMap[featureKey],
							updatedFeatureVersion,
							updatedFeatureVersion,
						),
					)
					if err != nil {
						return "", err
					}

					for _, featureVersion := range allFeaturesMap[featureKey] {
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

							// Select token for the feature's repo
							var token *dagger.Secret
							if repoStr != defaultFeaturesRepo {
								if m.CustomFeaturesRepoGhToken == nil {
									return "", fmt.Errorf("external repo %q present but CustomFeaturesRepoGhToken is not provided", repoStr)
								}
								token = m.CustomFeaturesRepoGhToken
							} else {
								token = m.PrefappGhToken
							}

							changelog, err := m.getReleaseChangelog(ctx, fullFeatureTag, repoStr, token)

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
		feature := featureData.(map[string]any)
		featureName := feature["name"].(string)
		versionProperty, hasVersion := feature["version"]
		if !hasVersion {
			continue
		}

		featureVersion := versionProperty.(string)
		currentFeaturesVersion[repoFeatureKey(featureRepo(feature), featureName)] = featureVersion
	}

	return currentFeaturesVersion
}
