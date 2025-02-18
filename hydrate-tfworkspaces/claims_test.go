package main

import (
	"context"
	"slices"
	"testing"
)

func TestClaims(t *testing.T) {

	ctx := context.Background()

	dir := "./fixtures/render-folder"

	appDir := getDir(dir)

	m := HydrateTfworkspaces{
		ValuesDir: appDir.Directory(dir),
	}

	claimNames, error := m.GetAppClaimNames(ctx)

	if error != nil {

		t.Errorf("Error getting claim names: %v", error)

	}

	if len(claimNames) != 1 {

		t.Errorf("Expected 1 claim name, got %v", len(claimNames))

	}

	for _, claimName := range []string{"example-platform"} {

		if !slices.Contains(claimNames, claimName) {

			t.Errorf("Expected claim name %v, got %v", claimName, claimNames)

		}

	}

}
