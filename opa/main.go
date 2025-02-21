// A generated module for Opa functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
	"dagger/opa/internal/dagger"
)

type Opa struct{}

func (m *Opa) Validate(
	ctx context.Context,
	policy *dagger.File,
	data *dagger.File,
	input *dagger.File,
) (*dagger.Container, error) {
	return dag.Container().
		From("openpolicyagent/conftest").
		WithMountedFile("/validation/policy.rego", policy).
		WithMountedFile("/validation/data.yaml", data).
		WithMountedFile("/validation/input.yaml", input).
		WithWorkdir("/validation").
		WithExec([]string{
			"conftest",
			"--rego-version", "v1",
			"test", "input.yaml",
			"--data", "data.yaml",
			"--policy", "policy.rego",
		}), nil

}
