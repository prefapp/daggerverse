package main

import (
	"context"
	"fmt"
	"slices"

	"github.com/Masterminds/semver/v3"
	"gopkg.in/yaml.v3"
)

func (m *UpdateClaimsFeatures) getAllClaims(
	ctx context.Context,
) ([]string, error) {
	var claims []string

	for _, ext := range []string{".yml", ".yaml"} {
		extClaims, err := m.ClaimsDir.Glob(ctx, fmt.Sprintf("**/*%s", ext))

		if err != nil {
			return []string{}, err
		}

		claims = append(claims, extClaims...)
	}

	return claims, nil
}

func (m *UpdateClaimsFeatures) getClaimIfKindComponent(
	ctx context.Context,
	claimPath string,
) (map[string]any, error) {
	file := m.ClaimsDir.File(claimPath)
	contents, err := file.Contents(ctx)
	if err != nil {
		return nil, err
	}

	var claim map[string]any
	err = yaml.Unmarshal([]byte(contents), &claim)
	if err != nil {
		return nil, err
	}

	if claim["kind"].(string) == "ComponentClaim" &&
		(m.ClaimsToUpdate == nil || slices.Contains(m.ClaimsToUpdate, claim["name"].(string))) {

		return claim, nil

	}

	return nil, nil
}

func (m *UpdateClaimsFeatures) updateClaimFeatures(
	claim map[string]any,
	featuresMap map[string]string,
) ([]map[string]string, bool, error) {
	updatedFeaturesList := []map[string]string{}
	createPR := false
	featuresList := claim["providers"].(map[string]any)["github"].(map[string]any)["features"].([]any)

	for idx, feature := range featuresList {
		featureName := feature.(map[string]any)["name"].(string)
		featureVersion := feature.(map[string]any)["version"].(string)
		if m.FeaturesToUpdate == nil || slices.Contains(m.FeaturesToUpdate, featureName) {
			featureVersionSemver, err := semver.NewVersion(
				featuresMap[featureName],
			)
			if err != nil {
				return []map[string]string{}, false, err
			}

			if featureVersion != "" {
				versionIsDifferent, err := semver.NewConstraint(
					fmt.Sprintf("!=%s", featureVersion),
				)
				if err != nil {
					return []map[string]string{}, false, err
				}

				// if instead of createPR = versionIsGreater.Check()
				// because a later updated feature could override this value
				if versionIsDifferent.Check(featureVersionSemver) {
					createPR = true
					claim["providers"].(map[string]any)["github"].(map[string]any)["features"].([]any)[idx].(map[string]any)["version"] = featuresMap[featureName]
					updatedFeaturesList = append(updatedFeaturesList, map[string]string{
						"name":    featureName,
						"version": featuresMap[featureName],
					})
				}
			}
		}
	}

	return updatedFeaturesList, createPR, nil
}
