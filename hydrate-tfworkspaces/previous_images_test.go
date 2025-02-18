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

	crs, error := m.GetPreviousImagesFromCrs(ctx)

	if error != nil {

		t.Errorf("Error getting claim names: %v", error)

	}

	if len(crs) != 1 {

		t.Errorf("Expected 1 claim name, got %v", len(crs))

	}

	if crs[0].Metadata.Annotations.ClaimRef != "TFWorkspaceClaim/example-platform" {

		t.Errorf("Expected claim name TFWorkspaceClaim/example-platform, got %v", crs[0].Metadata.Annotations.ClaimRef)

	}

}
