package main

import (
	"fmt"
)

func (m *NotifyAndHydrateState) CmdContainer() *Container {

	return dag.Container().
		From(fmt.Sprintf("%s:%s", m.FirestarterImage, m.FirestarterImageTag)).
		WithWorkdir("/library")

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
