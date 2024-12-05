package main

import (
	"context"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestBuildPreviousImageCanRecoverPreviousImages(t *testing.T) {

	ctx := context.Background()

	wetRepoDir := getDir("./test/wet-repo-dir")

	m := &HydrateKubernetes{}

	buildImagesContent := m.BuildPreviousImagesApp(
		ctx,
		wetRepoDir.Directory(
			"./test/wet-repo-dir/kubernetes/cluster-name/test-tenant/dev/",
		),
	)

	yamlDecoded := map[string]map[string]string{}

	err := yaml.Unmarshal([]byte(buildImagesContent), &yamlDecoded)

	if err != nil {

		t.Errorf("Error decoding yaml: %v", err)

	}

	if yamlDecoded["micro-a"]["image"] != "image-a:1.16.0" {

		t.Errorf("Expected image-a:1.16.0, got %v", yamlDecoded["micro-a"]["image"])

	}

	if yamlDecoded["micro-b"]["image"] != "image-b:1.16.0" {

		t.Errorf("Expected image-b:1.16.0, got %v", yamlDecoded["micro-b"]["image"])

	}

	if yamlDecoded["micro-c"]["image"] != "image-c:other-image" {

		t.Errorf("Expected image-c:other-image, got %v", yamlDecoded["micro-c"]["image"])

	}

}