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

) (*dagger.Container, error) {

	pass, err := helmRegistryPassword.Plaintext(ctx)

	if err != nil {
		return nil, err
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
		), nil
}
