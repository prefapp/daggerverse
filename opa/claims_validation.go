package main

import (
	"context"
	"dagger/opa/internal/dagger"
	"slices"
)

func (m *Opa) ValidateClaims(
	ctx context.Context,
	claimsDir *dagger.Directory,
	validationsDir *dagger.Directory,
	policiesDir *dagger.Directory,
) (*dagger.Directory, error) {

	return claimsDir, nil

}

func (m *Opa) PolicyIsApplicableToClaim(claimDataFile ClaimsDataFile, claim Claim) bool {

	for _, applicableClaim := range claimDataFile.ApplicableClaims {

		applicable := false

		for _, field := range getFieldsFromInstance(claim) {

			if applicable {

				return true

			}

			value := getValueFromField(&claim, field)

			applicableClaimFields := getFieldsFromInstance(applicableClaim)

			if slices.Contains(applicableClaimFields, field) {

				valueFromApplicableClaim := getValueFromField(applicableClaim, field)

				if value == "*" || value == valueFromApplicableClaim {

					applicable = true

				} else {

					applicable = false

					break

				}

			}

		}

	}

	return false

}
