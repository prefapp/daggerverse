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

			ctr, err := m.Validate(
				ctx,
				policiesDir.File(dataRule.RegoFile),
				dataRule.File,
				claim.File,
			)

			if err != nil {

				return err

			}

			exitCode, err := ctr.ExitCode(ctx)

			if err != nil {

				return err

			}

			fmt.Printf("Exit code: %d\n", exitCode)
		}
	}

	return nil

}

func (m *Opa) FindApplicableClaims(claims []ClaimClassification, data ClaimsDataRules) []ClaimClassification {

	var applicableClaims []ClaimClassification

	for _, claim := range claims {

		for _, applicableRule := range data.ApplyTo {

			applicable := true

			matchProperties := []string{
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

				if aCpropVal != "" && aCpropVal != claimPropVal {

					applicable = false

				}

			}

			if applicable {

				applicableClaims = append(applicableClaims, claim)

			}
		}

	}

	return applicableClaims

}
