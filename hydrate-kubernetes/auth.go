package main

import (
	"context"
	"dagger/hydrate-kubernetes/internal/dagger"
)

func prepareHelmLogin(

	ctx context.Context,

	ctr *dagger.Container,

	helmRegistry string,

	helmRegistryUser string,

	helmRegistryPassword *dagger.Secret,

) *dagger.Container {

	pass, err := helmRegistryPassword.Plaintext(ctx)

	if err != nil {
		panic("ERROR_PASSWORD " + err.Error())

	}

	return ctr.
		WithExec([]string{
			"helm", "registry", "login", helmRegistry,
			"--username", helmRegistryUser,
			"--password-stdin",
		},
			dagger.ContainerWithExecOpts{
				Stdin: pass,
			},
		)
}
