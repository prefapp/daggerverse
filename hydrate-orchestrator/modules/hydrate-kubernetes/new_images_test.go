package main

import (
	"context"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestNewImagesCanGenerateNewFile(t *testing.T) {

	ctx := context.Background()

	m := &HydrateKubernetes{}

	daggerFile, _ := m.BuildNewImages(
		ctx,
		"{ \"images\": [{ \"service_name_list\": [\"test\"], \"image\": \"test-image\" }] }",
	)

	contents, err := daggerFile.Contents(ctx)

	if err != nil {

		t.Errorf("Error reading file: %v", err)

	}

	unmarshaled := map[string]map[string]string{}

	err = yaml.Unmarshal([]byte(contents), &unmarshaled)

	if err != nil {

		t.Errorf("Error decoding yaml: %v", err)

	}

	if unmarshaled["test"]["image"] != "test-image" {

		t.Errorf("Expected test-image, got %v", unmarshaled["test"]["image"])

	}

}

func TestNewImagesCanGenerateNewFileFromImageKeys(t *testing.T) {

	ctx := context.Background()

	m := &HydrateKubernetes{}

	daggerFile, _ := m.BuildNewImages(
		ctx,
		"{ \"images\": [{ \"image_keys\": [\"/test/image\"], \"image\": \"test-image\" }] }",
	)

	contents, err := daggerFile.Contents(ctx)

	if err != nil {

		t.Errorf("Error reading file: %v", err)

	}

	unmarshaled := map[string]map[string]string{}

	err = yaml.Unmarshal([]byte(contents), &unmarshaled)

	if err != nil {

		t.Errorf("Error decoding yaml: %v", err)

	}

	if unmarshaled["test"]["image"] != "test-image" {

		t.Errorf("Expected test-image, got %v", unmarshaled["test"]["image"])

	}

}
