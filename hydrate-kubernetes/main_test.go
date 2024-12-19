package main

import (
	"context"
	"fmt"
	"slices"
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

	regularEntries, errGlob := renderedDir.Glob(ctx, "kubernetes/cluster-name/test-tenant/dev/*.yml")

	if errGlob != nil {

		t.Errorf("Error reading rendered files: %v", errGlob)

	}

	mapRegularEntries := map[string]bool{
		"kubernetes/cluster-name/test-tenant/dev/Deployment.sample-app-micro-a.yml": true,
		"kubernetes/cluster-name/test-tenant/dev/Deployment.sample-app-micro-b.yml": true,
		"kubernetes/cluster-name/test-tenant/dev/Deployment.sample-app-micro-c.yml": true,
		"kubernetes/cluster-name/test-tenant/dev/Service.sample-app-micro-a.yml":    true,
		"kubernetes/cluster-name/test-tenant/dev/Service.sample-app-micro-b.yml":    true,
		"kubernetes/cluster-name/test-tenant/dev/Service.sample-app-micro-c.yml":    true,
	}

	if len(regularEntries) != 6 {

		t.Errorf("Expected 6 files to be rendered, got %v", regularEntries)

	}

	for k := range mapRegularEntries {

		if !slices.Contains(regularEntries, k) {

			t.Errorf("Expected %s to be rendered, got %v", k, regularEntries)

		}

	}

	extraArtifacts, errGlob2 := renderedDir.Glob(ctx, "kubernetes/cluster-name/test-tenant/dev/extra_artifacts/*.yml")

	if errGlob2 != nil {

		t.Errorf("Error reading rendered files: %v", errGlob2)

	}

	extraArtifactsMap := map[string]bool{

		"kubernetes/cluster-name/test-tenant/dev/extra_artifacts/ExternalSecret.a.yml": true,

		"kubernetes/cluster-name/test-tenant/dev/extra_artifacts/ExternalSecret.b.yml": true,
	}

	if len(extraArtifacts) != 2 {

		t.Errorf("Expected 2 extra artifacts to be rendered, got %v", extraArtifacts)

	}

	for k := range extraArtifactsMap {

		if !slices.Contains(extraArtifacts, k) {

			t.Errorf("Expected %s to be rendered, got %v", k, extraArtifacts)

		}

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

func TestRenderSysAppsCanRenderWithExtraArtifacts(t *testing.T) {

	ctx := context.Background()

	valuesRepoDir := getDir("./fixtures/values-repo-dir-sys-services")

	helmDir := getDir("./helm-sys-services")

	m := &HydrateKubernetes{
		ValuesDir:    valuesRepoDir.Directory("fixtures/values-repo-dir-sys-services"),
		WetRepoDir:   dag.Directory(),
		Container:    dag.Container().From("ghcr.io/helmfile/helmfile:latest"),
		Helmfile:     helmDir.File("helm-sys-services/helmfile.yaml"),
		ValuesGoTmpl: helmDir.File("helm-sys-services/values.yaml.gotmpl"),
		RenderType:   "sys-services",
	}

	config, errContents := valuesRepoDir.
		File("./fixtures/values-repo-dir-sys-services/.github/hydrate_k8s_config.yaml").
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

	dir := m.Render(ctx, "stakater", "cluster-name", "", "", "")

	extraEntries, errGlob := dir.Glob(ctx, "cluster-name/stakater/extra_artifacts/*.yml")

	if errGlob != nil {

		t.Errorf("Error reading rendered files: %v", errGlob)

	}

	if extraEntries[0] != "cluster-name/stakater/extra_artifacts/ExternalSecret.a.yml" {

		t.Errorf("Expected ExternalSecret.a.yml to be rendered, got %v", extraEntries)

	}

	regularEntries, errGlob2 := dir.Glob(ctx, "cluster-name/stakater/*.yml")

	if errGlob2 != nil {

		t.Errorf("Error reading rendered files: %v", errGlob2)

	}

	mapEntries := map[string]bool{
		"cluster-name/stakater/ClusterRole.stakater-reloader-role.yml":                true,
		"cluster-name/stakater/ClusterRoleBinding.stakater-reloader-role-binding.yml": true,
		"cluster-name/stakater/Deployment.stakater-reloader.yml":                      true,
		"cluster-name/stakater/ServiceAccount.stakater-reloader.yml":                  true,
	}

	if len(regularEntries) != 4 {

		t.Errorf("Expected 4 files to be rendered, got %v", regularEntries)

	}

	for k := range mapEntries {

		if !slices.Contains(regularEntries, k) {

			t.Errorf("Expected %s to be rendered, got %v", k, regularEntries)

		}

	}
}
