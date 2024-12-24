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

	newImagesFile := m.BuildNewImages(ctx, newImagesMatrix)

	imagesFile := m.GetImagesFile(ctx, cluster, tenant, env)

	previousImagesFileContent := m.BuildPreviousImagesApp(ctx, cluster, tenant, env)

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

		helmfileCtr = prepareHelmLogin(
			ctx,
			helmfileCtr,
			m.HelmRegistry,
			m.HelmRegistryUser,
			m.HelmRegistryPassword,
		)
	}

	return helmfileCtr.
		WithExec([]string{
			"helmfile", "-e", env, "template",
			"--state-values-set-string", "tenant=" + tenant + ",app=" + app + ",cluster=" + cluster,
			"--state-values-file", "./kubernetes/" + cluster + "/" + tenant + "/" + env + ".yaml",
		}).
		Stdout(ctx)
}
