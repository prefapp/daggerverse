package main

import (
	"context"
	"time"
)

func (m *HydrateKubernetes) RenderSysService(

	ctx context.Context,

	cluster string,

	app string,

) (string, error) {

	if m.RepositoriesFile == nil {

		reposFile, err := m.BuildHelmRepositoriesFile(
			ctx,
			m.DotFirestartrDir,
			"./kubernetes-sys-services/"+cluster+"/"+app+".yaml",
		)

		if err != nil {

			return "", err

		}

		m.RepositoriesFile = reposFile

	}

	helmfileCtr := m.Container.
		WithDirectory("/values", m.ValuesDir).
		WithWorkdir("/values").
		WithMountedFile("/values/helmfile.yaml.gotmpl", m.Helmfile).
		WithFile("/values/repositories.yaml", m.RepositoriesFile).
		WithMountedFile("/values/values.yaml.gotmpl", m.ValuesGoTmpl).
		WithEnvVariable("BUST", time.Now().String())

	if m.HelmConfigDir != nil {

		helmfileCtr = helmfileCtr.WithDirectory("/helm/.config/helm", m.HelmConfigDir)

	}

	return helmfileCtr.
		WithExec([]string{
			"helmfile", "template",
			"--state-values-set-string", "app=" + app + ",cluster=" + cluster,
			"--state-values-file", "./kubernetes-sys-services/" + cluster + "/" + app + ".yaml",
		}).
		Stdout(ctx)
}
