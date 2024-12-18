package main

import (
	"context"
	"dagger/hydrate-kubernetes/internal/dagger"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

func (m *HydrateKubernetes) SplitRenderInFiles(

	ctx context.Context,

	renderFile *dagger.File,

) *dagger.Directory {

	content, err := renderFile.Contents(ctx)

	if err != nil {

		panic(err)

	}

	k8sManifests := strings.Split(string(content), "\n---\n")

	dir := dag.Directory()

	for _, manifest := range k8sManifests {

		k8sresource := KubernetesResource{}

		err := yaml.Unmarshal([]byte(manifest), &k8sresource)

		if err != nil {

			panic(err)

		}

		// create a new file for each k8s manifest
		// with <kind>.<metadata.name>.yml as the file name

		fileName := fmt.Sprintf("%s.%s.yml", k8sresource.Kind, k8sresource.Metadata.Name)

		dir = dir.WithNewFile(fileName, manifest)

	}

	return dir
}

func (m *HydrateKubernetes) DumpSysAppRenderToWetDir(

	ctx context.Context,

	app string,

	cluster string,

) *dagger.Directory {

	renderedChartFile, renderErr := m.RenderSysApp(ctx, cluster, app)

	if renderErr != nil {

		panic(renderErr)

	}

	tmpDir := m.SplitRenderInFiles(ctx,
		dag.Directory().
			WithNewFile("rendered.yaml", renderedChartFile).
			File("rendered.yaml"),
	)

	m.WetRepoDir = m.WetRepoDir.
		WithoutDirectory(cluster+"/"+app).
		WithDirectory(cluster+"/"+app, tmpDir)

	for _, regex := range []string{"*.yml", "*.yaml"} {

		entries, err := m.WetRepoDir.Glob(ctx, cluster+"/"+app+"/extra_artifacts/"+regex)

		if err != nil {

			panic(err)

		}

		for _, entry := range entries {

			extraFile := m.WetRepoDir.File(entry)

			m.WetRepoDir = m.WetRepoDir.WithFile(entry, extraFile)
		}

	}

	return m.WetRepoDir
}

func (m *HydrateKubernetes) DumpAppRenderToWetDir(

	ctx context.Context,

	app string,

	cluster string,

	tenant string,

	env string,

	// +optional
	// +default="{\"images\":[]}"
	newImagesMatrix string,

) *dagger.Directory {

	renderedChartFile, renderErr := m.RenderApp(
		ctx,
		env,
		app,
		cluster,
		tenant,
		newImagesMatrix,
	)

	if renderErr != nil {

		panic(renderErr)

	}

	tmpDir := m.SplitRenderInFiles(ctx,
		dag.Directory().
			WithNewFile("rendered.yaml", renderedChartFile).
			File("rendered.yaml"),
	)

	m.WetRepoDir = m.WetRepoDir.
		WithoutDirectory("kubernetes/"+cluster+"/"+tenant+"/"+env).
		WithDirectory(
			"kubernetes/"+cluster+"/"+tenant+"/"+env,
			tmpDir,
		)

	return m.WetRepoDir
}
