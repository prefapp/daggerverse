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

		ctr, err := prepareHelmLogin(
			ctx,
			helmfileCtr,
			m.HelmRegistry,
			m.HelmRegistryUser,
			m.HelmRegistryPassword,
		)

		helmfileCtr = ctr

		if err != nil {
			return "", err
		}
	}

	return helmfileCtr.
		WithExec([]string{
			"helmfile", "template",
			"--state-values-set-string", "app=" + app + ",cluster=" + cluster,
			"--state-values-file", "./" + cluster + "/" + app + ".yaml",
		}).
		Stdout(ctx)
}
