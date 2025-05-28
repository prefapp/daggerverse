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

	repositoryFileDir := getDir("./fixtures/repository_file")

	helmDir := getDir("./helm-apps")

	m := &HydrateKubernetes{
		ValuesDir:        valuesRepoDir.Directory("fixtures/values-repo-dir"),
		WetRepoDir:       wetRepoDir.Directory("fixtures/wet-repo-dir"),
		Container:        dag.Container().From("ghcr.io/helmfile/helmfile:latest"),
		Helmfile:         helmDir.File("helm-apps/helmfile.yaml.gotmpl"),
		ValuesGoTmpl:     helmDir.File("helm-apps/values.yaml.gotmpl"),
		RenderType:       "apps",
		RepositoriesFile: repositoryFileDir.File("fixtures/repository_file/repositories.yaml"),
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

	renderedDir, _ := m.Render(
		ctx,
		"sample-app",
		"cluster-name",
		"test-tenant",
		"dev",
		"{\"images\":[{\"service_name_list\":[\"micro-a\"],\"image\":\"new-image:1.0.0\",\"env\":\"dev\",\"app\":\"sample-app\",\"tenant\":\"test-tenant\",\"technology\":\"kubernetes\",\"platform\":\"cluster-name\"}]}",
	)

	newDpRendered := renderedDir[0].File("kubernetes/cluster-name/test-tenant/dev/Deployment.sample-app-micro-a.yml")

	if newDpRendered == nil {
		t.Errorf("Expected new Deployment.sample-app-micro-a.yml to be rendered")
	}

	dpMicroA := Artifact{}

	content, err := newDpRendered.Contents(ctx)

	if err != nil {

		t.Errorf("Error reading new Deployment.sample-app-micro-a.yml: %v", err)

	}

	errUnms := yaml.Unmarshal([]byte(content), &dpMicroA)

	if errUnms != nil {

		t.Errorf("Error unmarshalling new Deployment.sample-app-micro-a.yml: %v", errUnms)

	}

	// check if the jsonPatch works
	if dpMicroA.Metadata.Labels["test-label"] != "test-value" {

		t.Errorf("Expected new Deployment.sample-app-micro-a.yml to have label test-label=test-value, got %s", dpMicroA.Metadata.Labels)

	}

	// check if the new image is applied
	if dpMicroA.Metadata.Annotations.Image != "new-image:1.0.0" {

		t.Errorf("Expected new Deployment.sample-app-micro-a.yml to have image new-image:1.0.0, got %s", dpMicroA.Metadata.Annotations)

	}

	regularEntries, errGlob := renderedDir[0].Glob(ctx, "kubernetes/cluster-name/test-tenant/dev/*.yml")

	if errGlob != nil {

		t.Errorf("Error reading rendered files: %v", errGlob)

	}

	if len(regularEntries) != 8 {

		t.Errorf("Expected 8 files to be rendered, got %v", regularEntries)

	}

	mapEntries := map[string]bool{
		"kubernetes/cluster-name/test-tenant/dev/Deployment.sample-app-micro-a.yml": true,
		"kubernetes/cluster-name/test-tenant/dev/Deployment.sample-app-micro-b.yml": true,
		"kubernetes/cluster-name/test-tenant/dev/Deployment.sample-app-micro-c.yml": true,
		"kubernetes/cluster-name/test-tenant/dev/Service.sample-app-micro-a.yml":    true,
		"kubernetes/cluster-name/test-tenant/dev/Service.sample-app-micro-b.yml":    true,
		"kubernetes/cluster-name/test-tenant/dev/Service.sample-app-micro-c.yml":    true,
		"kubernetes/cluster-name/test-tenant/dev/ExternalSecret.a.yml":              true,
		"kubernetes/cluster-name/test-tenant/dev/ExternalSecret.b.yml":              true,
	}

	for k := range mapEntries {

		if !slices.Contains(regularEntries, k) {

			t.Errorf("Expected %s to be rendered, got %v", k, regularEntries)

		}

	}

}

func TestRenderAppsCanRenderNewImagesWithoutExecs(t *testing.T) {

	ctx := context.Background()

	valuesRepoDir := getDir("./fixtures/values-repo-dir")

	wetRepoDir := getDir("./fixtures/wet-repo-dir")

	repositoryFileDir := getDir("./fixtures/repository_file")

	helmDir := getDir("./helm-apps")

	m := &HydrateKubernetes{
		ValuesDir:        valuesRepoDir.Directory("fixtures/values-repo-dir"),
		WetRepoDir:       wetRepoDir.Directory("fixtures/wet-repo-dir"),
		Container:        dag.Container().From("ghcr.io/helmfile/helmfile:latest"),
		Helmfile:         helmDir.File("helm-apps/helmfile.yaml.gotmpl"),
		ValuesGoTmpl:     helmDir.File("helm-apps/values.yaml.gotmpl"),
		RepositoriesFile: repositoryFileDir.File("fixtures/repository_file/repositories.yaml"),
		RenderType:       "apps",
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

	renderedDir, _ := m.Render(
		ctx,
		"sample-app",
		"cluster-name",
		"test-tenant",
		"without_execs",
		"{\"images\":[]}",
	)

	fmt.Printf("Rendered dir: %v", renderedDir)

	newDpRendered := renderedDir[0].File("kubernetes/cluster-name/test-tenant/dev/Deployment.sample-app-micro-a.yml")

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

	repositoryFileDir := getDir("./fixtures/repository_file")

	helmDir := getDir("./helm-sys-services")

	m := &HydrateKubernetes{
		ValuesDir:        valuesRepoDir.Directory("fixtures/values-repo-dir-sys-services"),
		WetRepoDir:       dag.Directory(),
		Container:        dag.Container().From("ghcr.io/helmfile/helmfile:latest"),
		Helmfile:         helmDir.File("helm-sys-services/helmfile.yaml.gotmpl"),
		ValuesGoTmpl:     helmDir.File("helm-sys-services/values.yaml.gotmpl"),
		RepositoriesFile: repositoryFileDir.File("fixtures/repository_file/repositories.yaml"),
		RenderType:       "sys-services",
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

	dir, _ := m.Render(ctx, "stakater", "cluster-name", "", "", "")

	entries, errGlob := dir[0].Glob(ctx, "kubernetes-sys-services/cluster-name/stakater/*.yml")

	if errGlob != nil {

		t.Errorf("Error reading rendered files: %v", errGlob)

	}

	mapEntries := map[string]bool{
		"kubernetes-sys-services/cluster-name/stakater/ClusterRole.stakater-reloader-role.yml":                true,
		"kubernetes-sys-services/cluster-name/stakater/ClusterRoleBinding.stakater-reloader-role-binding.yml": true,
		"kubernetes-sys-services/cluster-name/stakater/Deployment.stakater-reloader.yml":                      true,
		"kubernetes-sys-services/cluster-name/stakater/ServiceAccount.stakater-reloader.yml":                  true,
		"kubernetes-sys-services/cluster-name/stakater/ExternalSecret.a.yml":                                  true,
	}

	if len(entries) != 5 {

		t.Errorf("Expected 5 files to be rendered, got %v", entries)

	}

	for k := range mapEntries {

		if !slices.Contains(entries, k) {

			t.Errorf("Expected %s to be rendered, got %v", k, entries)

		}

	}
}

func TestRenderAppsCanRenderImages(t *testing.T) {

	ctx := context.Background()

	valuesRepoDir := getDir("./fixtures/values-repo-dir")

	repositoryFileDir := getDir("./fixtures/repository_file")

	wetRepoDir := getDir("./fixtures/wet-repo-dir")

	helmDir := getDir("./helm-apps")

	m := &HydrateKubernetes{
		ValuesDir:        valuesRepoDir.Directory("fixtures/values-repo-dir"),
		WetRepoDir:       wetRepoDir.Directory("fixtures/wet-repo-dir"),
		Container:        dag.Container().From("ghcr.io/helmfile/helmfile:latest"),
		Helmfile:         helmDir.File("helm-apps/helmfile.yaml.gotmpl"),
		ValuesGoTmpl:     helmDir.File("helm-apps/values.yaml.gotmpl"),
		RepositoriesFile: repositoryFileDir.File("fixtures/repository_file/repositories.yaml"),
		RenderType:       "apps",
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

	renderedDir, _ := m.Render(
		ctx,
		"sample-app",
		"cluster-name",
		"test-tenant",
		"with_images_file",
		"{\"images\":[]}",
	)

	newDpRendered := renderedDir[0].File("kubernetes/cluster-name/test-tenant/with_images_file/Deployment.sample-app-micro-b.yml")

	dpMicroB := Artifact{}

	content, err := newDpRendered.Contents(ctx)

	if err != nil {

		t.Errorf("Error reading new Deployment.sample-app-micro-b.yml: %v", err)
	}

	errUnms := yaml.Unmarshal([]byte(content), &dpMicroB)

	if errUnms != nil {

		t.Errorf("Error unmarshalling new Deployment.sample-app-micro-b.yml: %v", errUnms)

	}

	// check if the new image is applied
	if dpMicroB.Metadata.Annotations.Image != "custom_image:0.1.0" {

		fmt.Printf("Annotations: %v", dpMicroB.Metadata.Annotations)

		t.Errorf("Expected new Deployment.sample-app-micro-b.yml to have image custom_image:0.1.0, got %s", dpMicroB.Metadata.Annotations)

	}

}
