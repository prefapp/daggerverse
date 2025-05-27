package main

import (
	"context"
	"testing"
)

func TestPreviousImages(t *testing.T) {

	ctx := context.Background()

	dir := "./fixtures/crs"

	wetDir := getDir(dir)

	m := HydrateTfworkspaces{
		WetRepoDir: wetDir.Directory(dir),
	}

	previousCr, err := m.GetPreviousCr(ctx, "example-platform")

	if err != nil {

		t.Errorf("Error getting claim names: %v", err)

	}

	if previousCr == nil {

		t.Errorf("Expected previousCr to be non-nil")

	}

}
