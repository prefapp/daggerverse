package main

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestPatchTfWorkspace(t *testing.T) {

	ctx := context.Background()

	appDir := getDir("./fixtures/render-folder/app-claims/tfworkspaces/example-platform/tenant-test/env-test")

	m := HydrateTfworkspaces{}

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
				Claim:            "example-claim",
				Technology:       "test",
				RepositoryCaller: "test",
			},
		},
	}

	resultDir, err := m.PatchClaimWithNewImageValues(
		ctx,
		imageMatrix,
		appDir.Directory("fixtures/render-folder/app-claims/tfworkspaces/example-platform/tenant-test/env-test"),
	)

	if err != nil {

		t.Errorf("Error patching workspace: %v", err)

	}

	contents, err := resultDir.File("claim.yaml").Contents(ctx)

	if err != nil {

		t.Errorf("Error reading file: %v", err)

	}

	fmt.Printf("contents: %v\n", contents)

	if !strings.Contains(contents, "test-image:latest") {

		t.Errorf("Expected test-image:latest, got %v", contents)

	}

}
