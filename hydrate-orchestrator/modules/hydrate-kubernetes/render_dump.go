package main

import (
	"bytes"
	"context"
	"dagger/hydrate-kubernetes/internal/dagger"
	"errors"
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

func (m *HydrateKubernetes) SplitRenderInFiles(

	ctx context.Context,

	renderFile *dagger.File,

) (*dagger.Directory, error) {

	content, err := renderFile.Contents(ctx)

	if err != nil {

		return nil, err

	}

	reader := strings.NewReader(content)
	dec := yaml.NewDecoder(reader)

	dir := dag.Directory()

	for {
		var node yaml.Node
		err := dec.Decode(&node)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}

		buf := bytes.Buffer{}
		enc := yaml.NewEncoder(&buf)
		enc.SetIndent(2)
		enc.Encode(&node)
		enc.Close()
		content := buf.Bytes()

		k8sresource := KubernetesResource{}

		err = yaml.Unmarshal(content, &k8sresource)

		if err != nil {

			return nil, err

		}

		if k8sresource.Kind == "" || k8sresource.Metadata.Name == "" {

			continue
		}

		// create a new file for each k8s manifest
		// with <kind>.<metadata.name>.yml as the file name
		fileName := fmt.Sprintf("%s.%s.yml", k8sresource.Kind, k8sresource.Metadata.Name)

		//add the --- at the beginning of the file
		data := "---\n" + string(content)

		dir = dir.WithNewFile(fileName, data)

	}

	return dir, nil
}

func (m *HydrateKubernetes) DumpSysAppRenderToWetDir(

	ctx context.Context,

	app string,

	cluster string,

) (*dagger.Directory, error) {

	renderedChartFile, err := m.RenderSysService(ctx, cluster, app)

	if err != nil {

		return nil, err

	}

	tmpDir, err := m.SplitRenderInFiles(ctx,
		dag.Directory().
			WithNewFile("rendered.yaml", renderedChartFile).
			File("rendered.yaml"),
	)

	if err != nil {
		return nil, err
	}

	m.WetRepoDir = m.WetRepoDir.
		WithoutDirectory("kubernetes-sys-services/"+cluster+"/"+app).
		WithDirectory("kubernetes-sys-services/"+cluster+"/"+app, tmpDir)

	envYaml, errEnvYaml := m.ValuesDir.File("kubernetes-sys-services/" + cluster + "/" + app + ".yaml").Contents(ctx)

	if errEnvYaml != nil {

		return nil, errEnvYaml

	}

	envYamlStruct := EnvYaml{}

	errUnmshEnv := yaml.Unmarshal([]byte(envYaml), &envYamlStruct)

	if errUnmshEnv != nil {

		return nil, errUnmshEnv

	}

	if envYamlStruct.RemoteArtifacts != nil {

		for _, remoteArtifact := range envYamlStruct.RemoteArtifacts {

			withRemotesArtifacts, err := m.Container.
				WithExec([]string{
					"curl",
					"-o",
					"/tmp/" + remoteArtifact.Filename, remoteArtifact.URL}).
				Sync(ctx)

			if err != nil {

				return nil, err

			}

			m.ValuesDir = m.ValuesDir.WithFile(
				"kubernetes-sys-services/"+cluster+"/"+app+"/extra_artifacts/"+remoteArtifact.Filename,
				withRemotesArtifacts.File("/tmp/"+remoteArtifact.Filename),
			)

		}

	}

	for _, regex := range []string{"**/*.yml", "**/*.yaml"} {

		entries, err := m.ValuesDir.
			Glob(ctx, "kubernetes-sys-services/"+cluster+"/"+app+"/extra_artifacts/"+regex)

		if err != nil {

			return nil, err

		}

		for _, entry := range entries {

			fmt.Printf("entry: %s\n", entry)

			extraFile := m.ValuesDir.File(entry)

			entry = strings.Replace(entry, "/extra_artifacts", "", 1)

			fmt.Printf("entry replaced: %s\n", entry)

			m.WetRepoDir = m.WetRepoDir.WithFile(entry, extraFile)

		}

	}

	renderedFiles, err := m.WetRepoDir.Glob(ctx, "kubernetes-sys-services/"+cluster+"/"+app+"/*.y*ml")
	if err != nil {
		return nil, err
	}

	err = m.ValidateYamlFiles(ctx, renderedFiles)

	if err != nil {
		return nil, err
	}

	return m.WetRepoDir, nil
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

) (*dagger.Directory, error) {

	renderedChartFile, err := m.RenderApp(
		ctx,
		env,
		app,
		cluster,
		tenant,
		newImagesMatrix,
	)

	if err != nil {

		return nil, err

	}

	tmpDir, err := m.SplitRenderInFiles(ctx,
		dag.Directory().
			WithNewFile("rendered.yaml", renderedChartFile).
			File("rendered.yaml"),
	)

	if err != nil {
		return nil, err
	}

	m.WetRepoDir = m.WetRepoDir.
		WithoutDirectory("kubernetes/"+cluster+"/"+tenant+"/"+env).
		WithDirectory(
			"kubernetes/"+cluster+"/"+tenant+"/"+env,
			tmpDir,
		)

	envYaml, errEnvYaml := m.ValuesDir.File("kubernetes/" + cluster + "/" + tenant + "/" + env + ".yaml").Contents(ctx)

	if errEnvYaml != nil {

		return nil, errEnvYaml

	}

	envYamlStruct := EnvYaml{}

	errUnmshEnv := yaml.Unmarshal([]byte(envYaml), &envYamlStruct)

	if errUnmshEnv != nil {

		return nil, errUnmshEnv

	}

	if envYamlStruct.RemoteArtifacts != nil {

		for _, remoteArtifact := range envYamlStruct.RemoteArtifacts {

			withRemotesArtifacts, err := m.Container.
				WithExec([]string{
					"curl",
					"-o",
					"/tmp/" + remoteArtifact.Filename, remoteArtifact.URL}).
				Sync(ctx)

			if err != nil {

				return nil, err

			}

			m.ValuesDir = m.ValuesDir.WithFile(
				"kubernetes/"+cluster+"/"+tenant+"/"+env+"/extra_artifacts/"+remoteArtifact.Filename,
				withRemotesArtifacts.File("/tmp/"+remoteArtifact.Filename),
			)

		}

	}

	for _, regex := range []string{"*.yml", "*.yaml"} {

		entries, err := m.ValuesDir.
			Glob(ctx, "kubernetes/"+cluster+"/"+tenant+"/"+env+"/extra_artifacts/"+regex)

		if err != nil {

			return nil, err

		}

		for _, entry := range entries {

			extraFile := m.ValuesDir.File(entry)

			entry = strings.Replace(entry, "/extra_artifacts", "", 1)

			m.WetRepoDir = m.WetRepoDir.WithFile(entry, extraFile)

		}

	}

	renderedFiles, err := m.WetRepoDir.Glob(ctx, "kubernetes/"+cluster+"/"+tenant+"/"+env+"/*.y*ml")
	if err != nil {
		return nil, err
	}

	err = m.ValidateYamlFiles(ctx, renderedFiles)

	if err != nil {
		return nil, err
	}

	return m.WetRepoDir, nil
}
