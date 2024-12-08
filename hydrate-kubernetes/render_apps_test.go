package main

import (
	"context"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestRenderAppsCanRenderNewImages(t *testing.T) {

	ctx := context.Background()

	valuesRepoDir := getDir("./fixtures/values-repo-dir")

	wetRepoDir := getDir("./fixtures/wet-repo-dir")

	helmDir := getDir("./helm")

	m := &HydrateKubernetes{
		ValuesDir:    valuesRepoDir.Directory("fixtures/values-repo-dir"),
		WetRepoDir:   wetRepoDir.Directory("fixtures/wet-repo-dir"),
		Container:    dag.Container().From("ghcr.io/helmfile/helmfile:latest"),
		Helmfile:     helmDir.File("helm/helmfile.yaml"),
		ValuesGoTmpl: helmDir.File("helm/values.yaml.gotmpl"),
	}

	depsContent, errContents := valuesRepoDir.
		File("./fixtures/values-repo-dir/.github/hydrate_deps.yaml").
		Contents(ctx)

	if errContents != nil {

		t.Errorf("Error reading deps file: %v", errContents)

	}

	m.Container = installDeps(depsContent, m.Container)

	renderedDir := m.RenderApps(
		ctx,
		"[\"kubernetes/cluster-name/test-tenant/dev/values.yaml\",\"kubernetes/cluster-name/test-tenant/pre/values.yaml\"]",
		"{\"images\":[{\"service_name_list\":[\"micro-a\"],\"image\":\"new-image:1.0.0\",\"env\":\"dev\",\"app\":\"sample-app\",\"tenant\":\"test-tenant\",\"base_folder\":\"kubernetes/cluster-name\"}]}",
		"sample-app",
	)

	// type ImageData struct {
	// 	Tenant           string   `json:"tenant"`
	// 	App              string   `json:"app"`
	// 	Env              string   `json:"env"`
	// 	ServiceNameList  []string `json:"service_name_list"`
	// 	Image            string   `json:"image"`
	// 	Reviewers        []string `json:"reviewers"`
	// 	BaseFolder       string   `json:"base_folder"`
	// 	RepositoryCaller string   `json:"repository_caller"`
	// }

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

	if artifact.Metadata.Annotations.Image != "new-image:1.0.0" {

		t.Errorf("Expected new Deployment.sample-app-micro-a.yml to have image new-image:1.0.0, got %s", artifact.Metadata.Annotations)

	}
}
