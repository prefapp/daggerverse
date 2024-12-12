package main

import (
	"context"
	"slices"
	"strings"
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

	entries, err := m.WetRepoDir.Glob(ctx, "kubernetes/*/*/*")

	if err != nil {
		panic(err)
	}

	targetDir := strings.Join([]string{"kubernetes", cluster, tenant, env}, "/")

	previousImagesFileContent := "{}"

	if slices.Contains(entries, targetDir) {

		previousImagesFileContent = m.BuildPreviousImagesApp(
			ctx,
			m.WetRepoDir.Directory(targetDir),
		)

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
			"/values/kubernetes/"+cluster+"/"+tenant+"/"+env+"/new_images.yaml",
			newImagesFile,
		)

	if m.HelmRegistryLoginNeeded == true {

		pass, err := m.HelmRegistryPassword.Plaintext(ctx)

		if err != nil {

			panic(err)

		}

		helmfileCtr = helmfileCtr.
			WithExec([]string{
				"helm", "registry", "login", m.HelmRegistry,
				"--username", m.HelmRegistryUser,
				"--password", pass,
			})
	}

	return helmfileCtr.
		WithExec([]string{
			"helmfile", "-e", env, "template",
			"--state-values-set-string", "tenant=" + tenant + ",app=" + app + ",cluster=" + cluster,
			"--state-values-file", "./kubernetes/" + cluster + "/" + tenant + "/" + env + ".yaml",
		}).
		Stdout(ctx)
}
