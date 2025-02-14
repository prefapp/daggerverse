// A generated module for HydrateTfworkspaces functions
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
	"dagger/hydrate-tfworkspaces/internal/dagger"
	"fmt"
	"path"
)

type HydrateTfworkspaces struct {
	FirestarterImage            string
	FirestarterImageTag         string
	GithubAppID                 string
	GithubInstallationID        string
	GithubPrefappInstallationID string
	GithubPrivateKey            *dagger.Secret
	GithubOrganization          string
	GhToken                     *dagger.Secret
	ClaimsDefaultBranch         string // +default="main"
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
	githubPrivateKey *dagger.Secret,
	// +required
	// Github organization
	githubOrganization string,
	// +required
	// Github token
	ghToken *dagger.Secret,
	// +default="main"
	claimsDefaultBranch string,

) *HydrateTfworkspaces {
	return &HydrateTfworkspaces{
		FirestarterImage: firestarterImage,

		FirestarterImageTag: firestarterImageTag,

		GithubAppID: githubAppID,

		GithubInstallationID: githubInstallationID,

		GithubPrefappInstallationID: githubPrefappInstallationID,

		GithubPrivateKey: githubPrivateKey,

		GithubOrganization: githubOrganization,

		GhToken: ghToken,

		ClaimsDefaultBranch: claimsDefaultBranch,
	}
}

func (m *HydrateTfworkspaces) CmdHydrate(
	// Claims repository name
	// +required
	claimsRepo string,
	// Claims directory
	// +required
	claimsDir *dagger.Directory,
	// Previous CRs directory
	// +required
	crsDir *dagger.Directory,
	// Provider to render
	// +required
	provider string,
	// GitHub application ID
	// +required
) *dagger.Directory {

	fmt.Printf(fmt.Sprintf("Hydrating CRs for %s", provider))

	claimsTargetDir := "/claims"
	crsTargetDir := "/crs"
	outputDir := "/output"

	cmd := m.CmdContainer().
		WithMountedDirectory(claimsTargetDir, claimsDir).
		WithMountedDirectory(crsTargetDir, crsDir).
		WithEnvVariable("GITHUB_APP_ID", m.GithubAppID).
		WithEnvVariable("GITHUB_INSTALLATION_ID", m.GithubInstallationID).
		WithEnvVariable("GITHUB_APP_INSTALLATION_ID_PREFAPP", m.GithubPrefappInstallationID).
		WithSecretVariable("GITHUB_APP_PEM_FILE", m.GithubPrivateKey).
		WithEnvVariable("ORG", m.GithubOrganization).
		WithEnvVariable("DEBUG", "firestartr-test:*").
		WithExec(
			[]string{
				"./run.sh",
				"cdk8s",
				"--render",
				"--disableRenames",
				"--globals", path.Join(crsTargetDir, ".config"),
				"--initializers", path.Join(crsTargetDir, ".config"),
				"--claims", path.Join(claimsTargetDir, "claims"),
				"--previousCRs", crsTargetDir,
				"--excludePath", path.Join(crsTargetDir, ".github"),
				"--claimsDefaults", path.Join(claimsTargetDir, ".config"),
				"--outputCrDir", outputDir,
				"--validateReferentialIntegrity", "disabled",
				"--provider", provider,
			},
		)

	return cmd.Directory(outputDir)

}

func (m *HydrateTfworkspaces) CmdContainer() *dagger.Container {

	return dag.Container().
		From(fmt.Sprintf("%s:%s", m.FirestarterImage, m.FirestarterImageTag))

}
