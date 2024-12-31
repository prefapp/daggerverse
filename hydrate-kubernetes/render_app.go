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
		WithNewFile(
			"/values/kubernetes/"+cluster+"/"+tenant+"/"+env+"/previous_images.yaml",
			previousImagesFileContent,
		).
		WithFile(
			"/values/kubernetes/"+cluster+"/"+tenant+"/"+env+"/images.yaml",
			imagesFile,
		).
		WithFile(
			"/values/kubernetes/"+cluster+"/"+tenant+"/"+env+"/new_images.yaml",
			newImagesFile,
		)

	if m.HelmRegistryLoginNeeded {

		helmfileCtr, err = prepareHelmLogin(
			ctx,
			helmfileCtr,
			m.HelmRegistry,
			m.HelmRegistryUser,
			m.HelmRegistryPassword,
		)

		if err != nil {
			return "", err
		}
	}

	return helmfileCtr.
		WithExec([]string{
			"helmfile", "-e", env, "template",
			"--state-values-set-string", "tenant=" + tenant + ",app=" + app + ",cluster=" + cluster,
			"--state-values-file", "./kubernetes/" + cluster + "/" + tenant + "/" + env + ".yaml",
		}).
		Stdout(ctx)
}
