package main

import (
	"context"
	"dagger/opa/internal/dagger"
	"fmt"
	"reflect"
)

func (m *Opa) ValidateClaims(
	ctx context.Context,
	claimsDir *dagger.Directory,
	validationsDir *dagger.Directory,
	policiesDir *dagger.Directory,
) error {

	claims, err := m.ClassifyClaims(ctx, claimsDir)

	if err != nil {

		return err

	}

	dataRules, err := m.LoadDataRules(ctx, validationsDir, m.App)

	if err != nil {
		return err
	}

	for _, dataRule := range dataRules {

		applicableClaims := m.FindApplicableClaims(claims, dataRule)

		for _, claim := range applicableClaims {

			_, err := m.Validate(
				ctx,
				policiesDir.File(dataRule.RegoFile),
				dataRule.File,
				claim.File,
			)

			if err != nil {

				return err

			}
		}
	}

	return nil

}

func (m *Opa) FindApplicableClaims(claims []ClaimClassification, data ClaimsDataRules) []ClaimClassification {

	fmt.Printf("Finding applicable claims for data rule %v\n", data)

	var applicableClaims []ClaimClassification

	for _, claim := range claims {

		for _, applicableRule := range data.ApplyTo {

			applicable := true

			matchProperties := []string{
				"App",
				"Name",
				"Kind",
				"ResourceType",
				"Environment",
				"Tenant",
				"Platform",
			}

			for _, property := range matchProperties {

				aCpropVal := reflect.ValueOf(applicableRule).FieldByName(property).String()

				claimPropVal := reflect.ValueOf(claim).FieldByName(property).String()

				fmt.Printf("PROPERTY: %s\n", property)
				fmt.Printf("ACPROPVAL: %s\n", aCpropVal)
				fmt.Printf("CLAIMPROPVAL: %s\n", claimPropVal)

				if aCpropVal != "" && aCpropVal != claimPropVal {

					fmt.Printf("❌ property %s does not match\n", property)

					applicable = false

					break

				} else {
					fmt.Printf("✅ property is skipped or matches, property: '%s', value from rule: '%s', value from claim: '%s'\n", property, aCpropVal, claimPropVal)
				}

			}

			if applicable {

				applicableClaims = append(applicableClaims, claim)

			}
		}

	}

	fmt.Printf("Applicable claims: %v\n", applicableClaims)

	return applicableClaims

}
