package main

import (
	"context"
	"fmt"
	"slices"

	"github.com/Masterminds/semver/v3"
	"github.com/xeipuuv/gojsonschema"
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
	schemaLoader *gojsonschema.SchemaLoader,
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

	schemaErrs, err := validateClaimMap(claim, schemaLoader)
	if err != nil {
		return nil, err
	}

	if schemaErrs == "" {
		claimName := claim["name"].(string)
		claimKind := claim["kind"].(string)
		claimProviderName := claim["providers"].(map[string]any)["github"].(map[string]any)["name"].(string)
		if claimKind == "ComponentClaim" &&
			(m.ClaimsToUpdate == nil ||
				slices.Contains(m.ClaimsToUpdate, claimName) ||
				slices.Contains(m.ClaimsToUpdate, claimProviderName)) {

			return claim, nil

		}
	} else {
		return nil, fmt.Errorf("Claim %s did not pass validation:\n%s\nSkipping\n", claimPath, schemaErrs)
	}

	return nil, nil
}

func (m *UpdateClaimsFeatures) updateClaimFeatures(
	claim map[string]any,
	featuresMap map[string]string,
) ([]map[string]any, bool, bool, error) {
	updatedFeaturesList := []map[string]any{}
	createPR := false
	hydrateClaim := false
	featuresListProperty, hasFeatures := claim["providers"].(map[string]any)["github"].(map[string]any)["features"]

	if hasFeatures {
		featuresList := featuresListProperty.([]any)

		for idx, feature := range featuresList {
			featureName := feature.(map[string]any)["name"].(string)
			featureVersionProperty, hasVersion := feature.(map[string]any)["version"]
			_, hasRef := feature.(map[string]any)["ref"]
			if m.FeaturesToUpdate == nil || slices.Contains(m.FeaturesToUpdate, featureName) {
				featureVersion := featureVersionProperty.(string)

				if hasVersion && featureVersion != "" {
					featureVersionSemver, err := semver.NewVersion(
						featuresMap[featureName],
					)
					if err != nil {
						return []map[string]any{}, false, false, err
					}

					versionIsDifferent, err := semver.NewConstraint(
						fmt.Sprintf("!=%s", featureVersion),
					)
					if err != nil {
						return []map[string]any{}, false, false, err
					}

					featureDataCopy := cloneMap(claim["providers"].(map[string]any)["github"].(map[string]any)["features"].([]any)[idx].(map[string]any))

					// if instead of createPR = versionIsGreater.Check()
					// because a later updated feature could override this value
					if versionIsDifferent.Check(featureVersionSemver) {
						createPR = true

						featureDataCopy["version"] = featuresMap[featureName]
					}

					updatedFeaturesList = append(
						updatedFeaturesList,
						featureDataCopy,
					)
				} else {
					if hasRef {
						hydrateClaim = true
					}
				}
			}
		}
	}

	return updatedFeaturesList, createPR, hydrateClaim, nil
}
