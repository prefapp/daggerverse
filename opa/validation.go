package main

import (
	"context"
	"dagger/opa/internal/dagger"
	"fmt"
	"time"
)

func (m *Opa) Validate(
	ctx context.Context,
	policy *dagger.File,
	data *dagger.File,
	file *dagger.File,
) (*dagger.Container, error) {

	fileName, err := file.Name(ctx)

	if err != nil {

		return nil, err

	}

	policyFileName, err := policy.Name(ctx)

	if err != nil {

		return nil, err

	}

	fmt.Printf("Validating file %s against policy %s\n", fileName, policyFileName)

	ctr, err := dag.Container().
		From("openpolicyagent/conftest").
		WithMountedFile("/validation/policy.rego", policy).
		WithMountedFile("/validation/data.yaml", data).
		WithMountedFile("/validation/input.yaml", file).
		WithWorkdir("/validation").
		WithEnvVariable("BUST_CACHE", time.Now().String()).
		WithExec([]string{
			"conftest",
			"--rego-version", "v1",
			"test", "input.yaml",
			"--data", "data.yaml",
			"--policy", "policy.rego",
		}).
		Sync(ctx)

	if err != nil {

		stderr, err := ctr.Stderr(ctx)

		if err != nil {

			return nil, err

		}

		return nil, fmt.Errorf("failed to run conftest: \n%s", stderr)

	}

	return ctr, nil
}
