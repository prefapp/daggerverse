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

	resultDir, error := m.PatchClaimsWithNewImageValues(
		ctx,
		"{\"images\": [{\"image_keys\": [\"test\"], \"image\": \"test-image:latest\", \"platform\": \"example-platform\"}]}",
		appDir.Directory("fixtures/render-folder/app-claims/tfworkspaces/example-platform/tenant-test/env-test"),
	)

	if error != nil {

		t.Errorf("Error patching workspace: %v", error)

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
