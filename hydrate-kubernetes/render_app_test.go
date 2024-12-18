package main

import (
	"context"
	"fmt"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestRenderAppCanRender(t *testing.T) {

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

	stdout, err := m.RenderApp(
		ctx,
		"dev",
		"sample-app",
		"cluster-name",
		"test-tenant",
		"{\"images\":[]}",
	)

	if err != nil {

		t.Errorf("Error rendering app: %v", err)

	}

	fmt.Println(stdout)
}
