package main

import (
	"context"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestNewImagesYAMLPreservesDistinctServiceImages(t *testing.T) {
	contents, err := buildNewImagesYAML(`{"images":[
		{"service_name_list":["micro-a"],"image":"image-A"},
		{"service_name_list":["micro-b"],"image":"image-B"}
	]}`)
	if err != nil {
		t.Fatal(err)
	}
	var values map[string]map[string]string
	if err := yaml.Unmarshal([]byte(contents), &values); err != nil {
		t.Fatal(err)
	}
	for service, expected := range map[string]string{"micro-a": "image-A", "micro-b": "image-B"} {
		if got := values[service]["image"]; got != expected {
			t.Errorf("expected %s image %q, got %q", service, expected, got)
		}
	}
}

func TestNewImagesYAMLPreservesImageKeysSharingParent(t *testing.T) {
	contents, err := buildNewImagesYAML(`{"images":[
		{"image_keys":["/workers/main/image"],"image":"image-A"},
		{"image_keys":["/workers/sidecar/image"],"image":"image-B"}
	]}`)
	if err != nil {
		t.Fatal(err)
	}
	var values map[string]map[string]map[string]string
	if err := yaml.Unmarshal([]byte(contents), &values); err != nil {
		t.Fatal(err)
	}
	for worker, expected := range map[string]string{"main": "image-A", "sidecar": "image-B"} {
		if got := values["workers"][worker]["image"]; got != expected {
			t.Errorf("expected workers/%s image %q, got %q", worker, expected, got)
		}
	}
}

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
