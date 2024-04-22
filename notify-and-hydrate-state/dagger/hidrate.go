package main

import (
    "fmt"
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
) *Container {

	cmd := m.CmdContainer().
		WithExec(
			[]string{},
		)

	return cmd

}

func (m *NotifyAndHydrateSt) CmdAffectedWetRepos(

    claimsFromMain *Dir,
    claimsFromPr *Dir,
    claimsDefaults *Dir,
    wetReposConfig *File

) *File {

    return m.CmdContainer().
      WithMountedDirectory("/w/main/claims", claimsFromMain).
      WithMountedDirectory("/w/claims", claimsFromPr).
      WithMountedDirectory("/w/pr/.config", claimsDefaults).
      WithMountedDirectory("/w/pr/.config/wet-repositories-config.yaml", wetReposConfig).
      WithExec([]string{

          "./run.sh"
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
      })


}
