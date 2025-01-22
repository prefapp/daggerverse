package main

import (
	"context"
	"time"
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
		WithNewFile("/values/kubernetes/"+cluster+"/"+tenant+"/"+env+"/previous_images.yaml", previousImagesFileContent).
		WithFile("/values/kubernetes/"+cluster+"/"+tenant+"/"+env+"/images.yaml", imagesFile).
		WithFile("/values/kubernetes/"+cluster+"/"+tenant+"/"+env+"/new_images.yaml", newImagesFile)

	if m.HelmConfigDir != nil {

		helmfileCtr = helmfileCtr.WithDirectory("/root/.config/helm", m.HelmConfigDir)
	}

	return helmfileCtr.
		WithExec([]string{
			"helmfile", "-e", env, "template",
			"--registry-config", "/root/.config/helm/registry/config.json",
			"--state-values-set-string", "tenant=" + tenant + ",app=" + app + ",cluster=" + cluster,
			"--state-values-file", "./kubernetes/" + cluster + "/" + tenant + "/" + env + ".yaml",
		}).
		Stdout(ctx)
}

type EnvYaml struct {
	RemoteArtifacts []struct {
		Filename string `yaml:"filename"`

		URL string `yaml:"url"`
	} `yaml:"remoteArtifacts,omitempty"`
}
