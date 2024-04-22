package main

import "fmt"

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

	cmd := m.CmdContainer()
		// .WithExec(
		// 	[]string{
        //         "./run.sh",
        //         "cdk8s",
        //         "--disableRenames",
        //         "--globals",
        //     },
		// )

	return cmd

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
