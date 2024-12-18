package main

import (
	"context"
	"dagger/hydrate-kubernetes/internal/dagger"
	"time"
)

func (m *HydrateKubernetes) RenderSysApp(

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

	if m.HelmRegistryLoginNeeded == true {

		pass, err := m.HelmRegistryPassword.Plaintext(ctx)

		if err != nil {
			panic("ERROR_PASSWORD " + err.Error())

		}

		helmfileCtr = helmfileCtr.
			WithExec([]string{
				"helm", "registry", "login", m.HelmRegistry,
				"--username", m.HelmRegistryUser,
				"--password-stdin",
			},
				dagger.ContainerWithExecOpts{
					Stdin: pass,
				},
			)
	}

	return helmfileCtr.
		WithExec([]string{
			"helmfile", "template",
			"--state-values-set-string", "app=" + app + ",cluster=" + cluster,
			"--state-values-file", "./" + cluster + "/" + app + ".yaml",
		}).Stdout(ctx)
}
