package main

import (
	"context"
	"fmt"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestRenderAppsCanRenderNewImages(t *testing.T) {

	ctx := context.Background()

	valuesRepoDir := getDir("./fixtures/values-repo-dir")

	wetRepoDir := getDir("./fixtures/wet-repo-dir")

	helmDir := getDir("./helm-apps")

	m := &HydrateKubernetes{
		ValuesDir:    valuesRepoDir.Directory("fixtures/values-repo-dir"),
		WetRepoDir:   wetRepoDir.Directory("fixtures/wet-repo-dir"),
		Container:    dag.Container().From("ghcr.io/helmfile/helmfile:latest"),
		Helmfile:     helmDir.File("helm-apps/helmfile.yaml"),
		ValuesGoTmpl: helmDir.File("helm-apps/values.yaml.gotmpl"),
		RenderType:   "apps",
	}

	config, errContents := valuesRepoDir.
		File("./fixtures/values-repo-dir/.github/hydrate_k8s_config.yaml").
		Contents(ctx)

	if errContents != nil {

		t.Errorf("Error reading deps file: %v", errContents)

	}

	configStruct := Config{}

	errUnmsh := yaml.Unmarshal([]byte(config), &configStruct)

	if errUnmsh != nil {

		t.Errorf("Error unmarshalling deps file: %v", errUnmsh)

	}

	m.Container = m.Container.From(configStruct.Image)

	m.Container = containerWithCmds(m.Container, configStruct.Commands)

	renderedDir := m.Render(
		ctx,
		"sample-app",
		"cluster-name",
		"test-tenant",
		"dev",
		"{\"images\":[{\"service_name_list\":[\"micro-a\"],\"image\":\"new-image:1.0.0\",\"env\":\"dev\",\"app\":\"sample-app\",\"tenant\":\"test-tenant\",\"base_folder\":\"kubernetes/cluster-name\"}]}",
	)

	newDpRendered := renderedDir.File("kubernetes/cluster-name/test-tenant/dev/Deployment.sample-app-micro-a.yml")

	if newDpRendered == nil {
		t.Errorf("Expected new Deployment.sample-app-micro-a.yml to be rendered")
	}

	artifact := Artifact{}

	content, err := newDpRendered.Contents(ctx)

	if err != nil {

		t.Errorf("Error reading new Deployment.sample-app-micro-a.yml: %v", err)

	}

	errUnms := yaml.Unmarshal([]byte(content), &artifact)

	if errUnms != nil {

		t.Errorf("Error unmarshalling new Deployment.sample-app-micro-a.yml: %v", errUnms)

	}

	// check if the jsonPatch works
	if artifact.Metadata.Labels["test-label"] != "test-value" {

		t.Errorf("Expected new Deployment.sample-app-micro-a.yml to have label test-label=test-value, got %s", artifact.Metadata.Labels)

	}

	// check if the new image is applied
	if artifact.Metadata.Annotations.Image != "new-image:1.0.0" {

		t.Errorf("Expected new Deployment.sample-app-micro-a.yml to have image new-image:1.0.0, got %s", artifact.Metadata.Annotations)

	}
}

func TestRenderAppsCanRenderNewImagesWithoutExecs(t *testing.T) {

	ctx := context.Background()

	valuesRepoDir := getDir("./fixtures/values-repo-dir")

	wetRepoDir := getDir("./fixtures/wet-repo-dir")

	helmDir := getDir("./helm-apps")

	m := &HydrateKubernetes{
		ValuesDir:    valuesRepoDir.Directory("fixtures/values-repo-dir"),
		WetRepoDir:   wetRepoDir.Directory("fixtures/wet-repo-dir"),
		Container:    dag.Container().From("ghcr.io/helmfile/helmfile:latest"),
		Helmfile:     helmDir.File("helm-apps/helmfile.yaml"),
		ValuesGoTmpl: helmDir.File("helm-apps/values.yaml.gotmpl"),
		RenderType:   "apps",
	}

	config, errContents := valuesRepoDir.
		File("./fixtures/values-repo-dir/.github/hydrate_k8s_config.yaml").
		Contents(ctx)

	if errContents != nil {

		t.Errorf("Error reading deps file: %v", errContents)

	}

	configStruct := Config{}

	errUnmsh := yaml.Unmarshal([]byte(config), &configStruct)

	if errUnmsh != nil {

		t.Errorf("Error unmarshalling deps file: %v", errUnmsh)

	}

	m.Container = m.Container.From(configStruct.Image)

	m.Container = containerWithCmds(m.Container, configStruct.Commands)

	renderedDir := m.Render(
		ctx,
		"sample-app",
		"cluster-name",
		"test-tenant",
		"without_execs",
		"{\"images\":[]}",
	)

	fmt.Printf("Rendered dir: %v", renderedDir)

	newDpRendered := renderedDir.File("kubernetes/cluster-name/test-tenant/dev/Deployment.sample-app-micro-a.yml")

	if newDpRendered == nil {
		t.Errorf("Expected new Deployment.sample-app-micro-a.yml to be rendered")
	}

	artifact := Artifact{}

	content, err := newDpRendered.Contents(ctx)

	if err != nil {

		t.Errorf("Error reading new Deployment.sample-app-micro-a.yml: %v", err)

	}

	errUnms := yaml.Unmarshal([]byte(content), &artifact)

	if errUnms != nil {

		t.Errorf("Error unmarshalling new Deployment.sample-app-micro-a.yml: %v", errUnms)

	}

	if artifact.Metadata.Annotations.Image != "image-a:1.16.0" {

		t.Errorf("Expected new Deployment.sample-app-micro-a.yml to have image image-a:1.16.0, got %s", artifact.Metadata.Annotations)

	}
}
