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

	imageMatrix := ImageMatrix{
		Images: []ImageData{
			{
				Tenant:           "test",
				App:              "test",
				Env:              "test",
				ServiceNameList:  []string{"test"},
				ImageKeys:        []string{"test"},
				Image:            "test-image:latest",
				Reviewers:        []string{"test"},
				Platform:         "example-platform",
				Technology:       "test",
				RepositoryCaller: "test",
			},
		},
	}

	crs, error := m.GetPreviousImagesFromCrs(ctx, imageMatrix)

	if error != nil {

		t.Errorf("Error getting claim names: %v", error)

	}

	if len(crs) != 0 {

		t.Errorf("Expected 0 claim names, got %v", len(crs))

	}

}
