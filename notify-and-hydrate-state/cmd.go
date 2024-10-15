package main

import (
	"context"
	"fmt"
	"path"
)

func (m *NotifyAndHydrateState) CmdContainer() *Container {

	return dag.Container().
		From(fmt.Sprintf("%s:%s", m.FirestarterImage, m.FirestarterImageTag))

}

// Render claims into CRs
func (m *NotifyAndHydrateState) CmdHydrate(
	// Claims repository name
	// +required
	claimsRepo string,
	// Claims directory
	// +required
	claimsDir *Directory,
	// Previous CRs directory
	// +required
	crsDir *Directory,
	// Provider to render
	// +required
	provider string,
	// GitHub application ID
	// +required
) *Directory {

	fsLog(fmt.Sprintf("Hydrating CRs for %s", provider))

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
				"--provider", provider,
			},
		)

	return cmd.Directory(outputDir)

}

// Render claims into CRs
func (m *NotifyAndHydrateState) CmdAnnotateCrPr(
	ctx context.Context,
	// Last claim PR link  (https://...//pulls/123)
	// +required
	lastClaimPrLink string,
	// Last state PR link (https://...//pulls/123)
	// +required
	lastStatePrLink string,
	// Previous CRs directory
	// +required
	wetRepo *Directory,
	// Path to the cr
	// +required
	crFileName string,
) *Directory {

	wetRepoPath := "/repo"

	cmd := m.CmdContainer().
		WithMountedDirectory(wetRepoPath, wetRepo).
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
				"--crLocation", wetRepoPath + "/" + crFileName,
				"--lastStatePrLink", lastStatePrLink,
				"--lastClaimPrLink", lastClaimPrLink,
			},
		)

	return cmd.Directory(wetRepoPath)

}
func (m *NotifyAndHydrateState) CmdAffectedWetRepos(

	claimsFromMain *Directory,
	claimsFromPr *Directory,
	claimsDefaults *Directory,
	wetReposConfig *File,

) *File {

	return m.CmdContainer().
		WithMountedDirectory("/w/main/claims", claimsFromMain).
		WithMountedDirectory("/w/claims", claimsFromPr).
		WithMountedDirectory("/w/pr/.config", claimsDefaults).
		WithMountedFile("/w/pr/.config/wet-repositories-config.yaml", wetReposConfig).
		WithExec([]string{

			"./run.sh",
			"cdk8s",
			"--compare",
			"--disableRenames",
			"--claimsFromMain",
			"/w/main/claims",
			"--claimsFromPr",
			"/w/claims",
			"--claimsDefaults",
			"/w/pr/.config",
			"--wetReposConfig",
			"/w/pr/.config/wet-repositories-config.yaml",
			"--outputComparer",
			"/w/AFFECTED_WET_REPOSITORIES.json",
		}).File("/w/AFFECTED_WET_REPOSITORIES.json")

}

func (m *NotifyAndHydrateState) CmdAnnotateCRs(
	// Claims repository name
	// +required
	claimsRepo string,
	// Wet repository name
	// +required
	wetRepo string,
	// Wet PR number
	// +required
	wetPrNumber string,
	// CRs directory
	// +required
	crsDir *Directory,
) *Directory {

	targetCrsDir := "/output"

	return m.CmdContainer().
		WithMountedDirectory(targetCrsDir, crsDir).
		WithExec([]string{
			"./run.sh", "cdk8s",
			"--crLocation", targetCrsDir,
			"--lastStatePrLink", fmt.Sprintf("%s#%s", wetRepo, wetPrNumber),
			"--lastClaimPrLink", fmt.Sprintf("%s#%s", claimsRepo, wetPrNumber),
		}).Directory(targetCrsDir)
}
