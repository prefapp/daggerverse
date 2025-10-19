package main

import (
	"context"
	"testing"
)

func TestDataRulesLoader(t *testing.T) {

	m := Opa{
		App: "sample-app",
	}

	claimsDir := getDir("fixtures/.firestartr/validations")

	t.Run("Can load data rules", func(t *testing.T) {

		dataRules, err := m.LoadDataRules(
			context.Background(),
			claimsDir.Directory("fixtures/.firestartr/validations"),
			"sample-app",
		)

		if err != nil {

			t.Errorf("Error: %v", err)

		}

		if len(dataRules) != 2 {

			t.Errorf("Expected 2, got %d", len(dataRules))

		}

	})
}
