package main

import (
	"context"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestGetImagesFileCanGenerateImagesFile(t *testing.T) {

	ctx := context.Background()

	valuesRepoDir := getDir("./fixtures/values-repo-dir")

	m := &HydrateKubernetes{
		ValuesDir: valuesRepoDir.Directory("./fixtures/values-repo-dir"),
	}

	imagesFile := m.GetImagesFile(ctx, "cluster-name", "test-tenant", "with_images_file")

	yamlContent, errCnts := imagesFile.Contents(ctx)

	if errCnts != nil {

		t.Errorf("Error reading images file: %v", errCnts)

	}

	unmarshaled := map[string]map[string]string{}

	errUnmsh := yaml.Unmarshal([]byte(yamlContent), &unmarshaled)

	if errUnmsh != nil {

		t.Errorf("Error decoding yaml: %v", errUnmsh)

	}

	if unmarshaled["micro-b"]["image"] != "custom_image:0.1.0" {

		t.Errorf("Expected custom_image:0.1.0, got %v", unmarshaled["micro-b"]["image"])

	}
}
