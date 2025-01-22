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

	helmfileCtr := m.Container.
		WithDirectory("/values", m.ValuesDir).
		WithWorkdir("/values").
		WithMountedFile("/values/helmfile.yaml", m.Helmfile).
		WithMountedFile("/values/values.yaml.gotmpl", m.ValuesGoTmpl).
		WithEnvVariable("BUST", time.Now().String())

	if m.HelmRegistryLoginNeeded {

		helmfileCtr = prepareHelmLogin(
			helmfileCtr,
			m.HelmConfigDir,
		)
	}

	return helmfileCtr.
		WithExec([]string{
			"helmfile", "template",
			"--state-values-set-string", "app=" + app + ",cluster=" + cluster,
			"--state-values-file", "./kubernetes/" + cluster + "/" + app + ".yaml",
		}).
		Stdout(ctx)
}
