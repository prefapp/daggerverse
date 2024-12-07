package main

import (
	"context"
	"fmt"
	"testing"
)

func TestRenderAppCanRender(t *testing.T) {

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

	stdout, err := m.RenderApp(
		ctx,
		"dev",
		"sample-app",
		"cluster-name",
		"test-tenant",
		"{\"images\":[]}",
	).Stdout(ctx)

	if err != nil {

		t.Errorf("Error rendering app: %v", err)

	}

	fmt.Println(stdout)
}
