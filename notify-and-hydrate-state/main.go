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

	"slices"
)

type NotifyAndHydrateState struct {
	FirestarterImage            string
	FirestarterImageTag         string
	GithubAppID                 string
	GithubInstallationID        string
	GithubPrefappInstallationID string
	GithubPrivateKey            *Secret
	GithubOrganization          string
	GhToken                     *Secret
}

func New(

	// +optional
	// +default="latest-slim"
	firestarterImageTag string,

	// +optional
	// +default="ghcr.io/prefapp/gitops-k8s"
	firestarterImage string,

	// +required
	// Github application ID
	githubAppID string,

	// +required
	// Github installation ID
	githubInstallationID string,

	// +required
	// Github prefapp installation ID
	githubPrefappInstallationID string,

	// +required
	// Github private key
	githubPrivateKey *Secret,

	// +required
	// Github organization
	githubOrganization string,

	// +required
	// Github token
	ghToken *Secret,

) *NotifyAndHydrateState {

	return &NotifyAndHydrateState{

		FirestarterImage: firestarterImage,

		FirestarterImageTag: firestarterImageTag,

		GithubAppID: githubAppID,

		GithubInstallationID: githubInstallationID,

		GithubPrefappInstallationID: githubPrefappInstallationID,

		GithubPrivateKey: githubPrivateKey,

		GithubOrganization: githubOrganization,

		GhToken: ghToken,
	}

}

func (m *NotifyAndHydrateState) Workflow(
	ctx context.Context,
	// Claims repository name
	// +required
	claimsRepo string,
	// Wet repository name
	// +required
	wetRepo string,
	// Claims directory
	// +required
	claimsDir *Directory,
	// Previous CRs directory
	// +required
	crsDir *Directory,
	// Provider to render
	// +required
	provider string,

	claimsPr string,
) {

	newCrsDir := m.CmdHydrate(

		claimsRepo,

		claimsDir,

		crsDir,

		provider,
	)

	comparedResult := m.CompareDirs(ctx, crsDir, newCrsDir)

	m.Verify(

		ctx,

		claimsPr,

		wetRepo,

		slices.Concat(

			comparedResult.AddedFiles,

			comparedResult.ModifiedFiles,

			comparedResult.DeletedFiles,
		),
	)

	// Git checkout -b automated/<cr-name>-<pr-number>

	// Git add  per file

	// Git commit -m "Automated commit for CR <cr-name>"

	// Git push origin automated/<cr-name>-<pr-number>

	// Github create PR automated/<cr-name>-<pr-number> to <wet-repo>

	// Git checkout automated/<cr-name>-<pr-number>

	// Launch cdk8s renderer to add the annotation to the CR

	// Git add per file

	// Git commit -m "Automated commit for CR <cr-name>"

	// Git push origin automated/<cr-name>-<pr-number>

	// Add comment to claims PR with the PR links in list format

}
