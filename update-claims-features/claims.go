package main

import (
	"context"
	"fmt"

	"github.com/Masterminds/semver/v3"
	"gopkg.in/yaml.v3"
)

func (m *UpdateClaimsFeatures) getAllClaims(
	ctx context.Context,
) ([]string, error) {
	var claims []string

	for _, ext := range []string{".yml", ".yaml"} {
		extClaims, err := m.ClaimsDir.Glob(
			ctx,
			fmt.Sprintf("claims/%s/*%s", m.ComponentsFolderName, ext),
		)

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
) (*Claim, error) {
	file := m.ClaimsDir.File(claimPath)
	contents, err := file.Contents(ctx)
	if err != nil {
		return nil, err
	}

	claim := &Claim{}

	err = yaml.Unmarshal([]byte(contents), claim)

	if err != nil {
		return nil, err
	}

	if claim.Kind == "ComponentClaim" &&
		(m.ClaimToUpdate == "" || m.ClaimToUpdate == claim.Name) {

		return claim, nil

	}

	return nil, nil
}

func (m *UpdateClaimsFeatures) updateClaimFeatures(
	ctx context.Context,
	claim *Claim,
	featuresMap map[string]string,
) ([]Feature, bool, error) {
	var updatedFeaturesList []Feature
	createPR := false

	for _, feature := range claim.Providers.Github.Features {
		if m.FeatureToUpdate == "" || m.FeatureToUpdate == feature.Name {
			featureVersionSemver, err := semver.NewVersion(
				featuresMap[feature.Name],
			)
			if err != nil {
				return []Feature{}, false, err
			}

			versionIsGreater, err := semver.NewConstraint(
				fmt.Sprintf("> %s", feature.Version),
			)
			if err != nil {
				return []Feature{}, false, err
			}

			// if instead of createPR = versionIsGreater.Check()
			// because a latter unupdated feature could override this value
			if versionIsGreater.Check(featureVersionSemver) {
				createPR = true
				feature.Version = featuresMap[feature.Name]
			}
		}

		// Add feature whether its version is greater or not,
		// so unupdated features are not deleted
		updatedFeaturesList = append(updatedFeaturesList, feature)
	}

	return updatedFeaturesList, createPR, nil
}
