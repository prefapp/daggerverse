package main

import (
	"context"
	"testing"
)

func TestClaimsValidation(t *testing.T) {

	m := Opa{
		App: "sample-app",
	}

	validationsDir := getDir("fixtures/.firestartr/validations")
	policiesDir := getDir("fixtures/.firestartr/validations/policies")

	t.Run("ValidateClaims must fail on invalid claims", func(t *testing.T) {

		claimsDir := getDir("fixtures/tfworkspaces-invalid")

		err := m.ValidateClaims(
			context.Background(),
			claimsDir.Directory("fixtures/tfworkspaces-invalid"),
			validationsDir.Directory("fixtures/.firestartr/validations"),
			policiesDir.Directory("fixtures/.firestartr/validations/policies"),
		)

		if err == nil {

			t.Errorf("Validation must fail, but it passed")

		}

	})

	t.Run("ValidateClaims must pass on valid claims", func(t *testing.T) {

		claimsDir := getDir("fixtures/tfworkspaces-valid")

		err := m.ValidateClaims(
			context.Background(),
			claimsDir.Directory("fixtures/tfworkspaces-valid"),
			validationsDir.Directory("fixtures/.firestartr/validations"),
			policiesDir.Directory("fixtures/.firestartr/validations/policies"),
		)

		if err != nil {

			t.Errorf("Validation must pass, but it failed")

		}

	})

}
