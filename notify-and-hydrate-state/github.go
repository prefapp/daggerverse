package main

import (
	"context"
	"strings"
)

func (m *NotifyAndHydrateState) AddPRReferences(
	ctx context.Context,
	// Claims repository name
	// +required
	claimsRepo string,
	// Claim PR number
	// +required
	claimPrNumber string,
	// List of PR links
	// +required
	prLinks []string,
	// Github Token
	// +required
	githubToken *Secret,

) (string, error) {

	return dag.Gh().Run(ctx, githubToken, strings.Join([]string{
		"issue",
		"-R", claimsRepo,
		"edit", claimPrNumber,
		"-b", strings.Join(prLinks, " "),
	}, " "))

}
