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
) (string, error) {

	fileName, err := file.Name(ctx)

	if err != nil {

		return "", err

	}

	policyFileName, err := policy.Name(ctx)

	if err != nil {

		return "", err

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
		}, dagger.ContainerWithExecOpts{
			RedirectStderr: "/tmp/stderr",
			RedirectStdout: "/tmp/stdout",
			Expect:         "ANY",
		}).Sync(ctx)

	eC, err := ctr.ExitCode(ctx)

	if err != nil {

		return "", err

	}

	if eC != 0 {

		stderr, _ := ctr.File("/tmp/stderr").Contents(ctx)
		stdout, _ := ctr.File("/tmp/stdout").Contents(ctx)

		return "", fmt.Errorf("%s\n%s", stdout, stderr)
	}

	return "", nil
}
