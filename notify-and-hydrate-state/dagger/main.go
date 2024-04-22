// A generated module for NotifyAndHydrateState functions
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
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type NotifyAndHydrateState struct{}

func (m *NotifyAndHydrateState) Run(
	branchRef string,
	prNumber int,
	claimsRepo string,
	claimsDir *Directory,
	crsDir *Directory,
	// Firestarter image tag
	// +optional
	// +default="latest"
	firestarterImageTag string,
) *Container {
	operator := dag.Container().From(fmt.Sprintf("ghcr.io/prefapp/gitops-k8s:%s", firestarterImageTag)).WithMountedDirectory("/claims", claimsDir).WithMountedDirectory("/crs", crsDir)
	return operator.WithExec([]string{"echo", "works"})
}

type Metadata struct {
	Name string `yaml:"name"`
}

type CR struct {
	Metadata Metadata
}

func (m *NotifyAndHydrateState) Foo(ctx context.Context) *Container {

	return dag.Container().
		From("alpine:latest")

}

func (m *NotifyAndHydrateState) CrHasOpenedPr(
	ctx context.Context,
	ghToken Secret,
	// Claims Pr (e.g. "prefapp/claims#123")
	// +optional
	// +default="prefapp/claims#123"
	claimsPr string,
	cr *File,
) (bool, error) {

	content, err := cr.Contents(ctx)

	if err != nil {

		return false, err

	}

	crObj := CR{}

	yaml.Unmarshal([]byte(content), &crObj)

	return true, nil

}

type PrBranchName struct {
	HeadRefName string `json:"headRefName"`
}

func (m *NotifyAndHydrateState) GetRepoPrsByBranchName(

	ctx context.Context,

	ghToken *Secret,

	// Repository name (e.g. "firestartr-test/state-github")
	ghRepo string,

) ([]PrBranchName, error) {

	command := strings.Join(

		[]string{

			"pr",

			"list",

			"--json",

			"headRefName",

			"-R",

			ghRepo,
		},

		" ",
	)

	content, err := dag.
		Gh().
		Run(
			ctx,
			ghToken,
			command,
		)

	if err != nil {

		return nil, err
	}

	prs := []PrBranchName{}

	json.Unmarshal(
		[]byte(content),
		&prs,
	)

	return prs, nil
}
