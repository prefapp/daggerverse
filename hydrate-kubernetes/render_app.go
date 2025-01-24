package main

import (
	"context"
	"dagger/hydrate-kubernetes/internal/dagger"
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

func (m *HydrateKubernetes) RenderApp(

	ctx context.Context,

	env string,

	app string,

	cluster string,

	tenant string,

	// +optional
	// +default="{\"images\":[]}"
	newImagesMatrix string,

) (string, error) {

	if m.RepositoriesFile == nil {

		reposFile, err := m.BuildHelmRepositoriesFile(
			ctx,
			m.DotFirestartrDir,
			"./kubernetes/"+cluster+"/"+tenant+"/"+env+".yaml",
		)

		if err != nil {

			return "", err

		}

		m.RepositoriesFile = reposFile

	}

	newImagesFile, err := m.BuildNewImages(ctx, newImagesMatrix)

	if err != nil {
		return "", err
	}

	imagesFile, err := m.GetImagesFile(ctx, cluster, tenant, env)

	if err != nil {
		return "", err
	}

	previousImagesFileContent, err := m.BuildPreviousImagesApp(ctx, cluster, tenant, env)

	if err != nil {
		return "", err
	}

	helmfileCtr := m.Container.
		WithDirectory("/values", m.ValuesDir).
		WithWorkdir("/values").
		WithMountedFile("/values/helmfile.yaml", m.Helmfile).
		WithMountedFile("/values/values.yaml.gotmpl", m.ValuesGoTmpl).
		WithEnvVariable("BUST", time.Now().String()).
		WithFile("/values/kubernetes/repositories.yaml", m.RepositoriesFile).
		WithFile("/values/kubernetes/environments.yaml", m.CreateEnvironmentsFile(env)).
		WithNewFile("/values/kubernetes/"+cluster+"/"+tenant+"/"+env+"/previous_images.yaml", previousImagesFileContent).
		WithFile("/values/kubernetes/"+cluster+"/"+tenant+"/"+env+"/images.yaml", imagesFile).
		WithFile("/values/kubernetes/"+cluster+"/"+tenant+"/"+env+"/new_images.yaml", newImagesFile)

	if m.HelmConfigDir != nil {

		helmfileCtr = helmfileCtr.WithDirectory("/helm/.config/helm", m.HelmConfigDir)
	}

	return helmfileCtr.
		WithExec([]string{
			"helmfile", "-e", env, "template",
			"--state-values-set-string", "tenant=" + tenant + ",app=" + app + ",cluster=" + cluster,
			"--state-values-file", "./kubernetes/" + cluster + "/" + tenant + "/" + env + ".yaml",
		}).
		Stdout(ctx)
}

func (m *HydrateKubernetes) CreateEnvironmentsFile(env string) *dagger.File {

	envs := map[string]map[string]interface{}{
		"environments": {
			fmt.Sprintf("%v", env): make(map[string]interface{}),
		},
	}

	contentFile, err := yaml.Marshal(envs)

	if err != nil {

		panic(err)

	}

	return dag.Directory().
		WithNewFile("environments.yaml", string(contentFile)).
		File("environments.yaml")

}
