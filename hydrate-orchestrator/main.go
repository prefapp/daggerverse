package main

import (
	"context"
	"dagger/hydrate-orchestrator/internal/dagger"
)

type HydrateOrchestrator struct {
	Repo             string
	GhToken          *dagger.Secret
	App              string
	ValuesStateDir   *dagger.Directory
	WetStateDir      *dagger.Directory
	DeploymentBranch string
}

func New(
	ctx context.Context,
	// Github repository name <owner>/<repo>
	// +required
	repo string,
	// GitHub token
	// +required
	ghToken *dagger.Secret,
	// Application name
	// +required
	app string,
	// State values directory (e.g. state-app-<app>#main)
	// +required
	valuesStateDir *dagger.Directory,
	// Wet state directory (e.g. wet-state-app-<app>#<deployment-branch>)
	// +required
	wetStateDir *dagger.Directory,
	// Deployment branch to hydrate
	// +optional
	// +default="deployment"
	deploymentBranch string,

) *HydrateOrchestrator {
	return &HydrateOrchestrator{
		Repo:             repo,
		GhToken:          ghToken,
		App:              app,
		ValuesStateDir:   valuesStateDir,
		WetStateDir:      wetStateDir,
		DeploymentBranch: deploymentBranch,
	}
}
