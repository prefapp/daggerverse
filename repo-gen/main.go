// A generated module for RepoGen functions
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
	"strings"
)

type RepoGen struct {
	GhToken *Secret
}

func New(
	// +required
	// Github token
	ghToken *Secret,
) *RepoGen {
	return &RepoGen{
		GhToken: ghToken,
	}
}

func (m *RepoGen) GetRepositories(ctx context.Context, owner string) string {

	command := strings.Join([]string{

		"repo",

		"list",

		owner,
	},
		" ",
	)

	if m.GhToken == nil {

		panic("GhToken is required")

	}

	resp, err := dag.Gh().Run(ctx, m.GhToken, command, GhRunOpts{DisableCache: true})

	if err != nil {

		panic(err)

	}

	return resp

}
