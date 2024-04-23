package main

import (
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
    githubAppID string,
    // GitHub installation ID
    // +required
    githubInstallationID string,
    // Github Prefapp App installation ID
    // +required
    githubPrefappInstallationID string,
    // GitHub private key
    // +required
    githubPrivateKey *Secret,
    // GitHub Organization
    // +required
    githubOrganization string,
) *Directory {

    claimsTargetDir := "/claims"
    crsTargetDir := "/crs"
    outputDir := "/output"

	cmd := m.CmdContainer().
        WithMountedDirectory(claimsTargetDir, claimsDir).
        WithMountedDirectory(crsTargetDir, crsDir).
        WithEnvVariable("GITHUB_APP_ID", githubAppID).
        WithEnvVariable("GITHUB_INSTALLATION_ID", githubInstallationID).
        WithEnvVariable("GITHUB_APP_INSTALLATION_ID_PREFAPP", githubPrefappInstallationID).
        WithSecretVariable("GITHUB_APP_PEM_FILE", githubPrivateKey).
        WithEnvVariable("ORG", githubOrganization).
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
