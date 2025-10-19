package main

import (
	"context"
	"testing"
)

func TestClassifyClaims(t *testing.T) {

	m := Opa{
		App: "sample-app",
	}

	claimsDir := getDir("fixtures/tfworkspaces-valid")

	classifications, err := m.ClassifyClaims(
		context.Background(),
		claimsDir.Directory("fixtures/tfworkspaces-valid"),
	)

	if err != nil {

		t.Errorf("Error: %v", err)

	}

	if len(classifications) != 2 {

		t.Errorf("Expected 2, got %d", len(classifications))

	}

}
