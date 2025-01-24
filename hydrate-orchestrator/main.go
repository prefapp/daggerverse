package main

import (
	"context"
	"dagger/hydrate-orchestrator/internal/dagger"
)

type EventType string

const (
	// Undetermined risk; analyze further.
	PullRequest EventType = "pr"

	// Minimal risk; routine fix.
	Manual EventType = "manual"

	// Moderate risk; timely fix.
	Dispatch EventType = "dispatch"
)

type HydrateOrchestrator struct {
	Repo             string
	GhToken          *dagger.Secret
	App              string
	ValuesStateDir   *dagger.Directory
	WetStateDir      *dagger.Directory
	AuthDir          *dagger.Directory
	DeploymentBranch string
	Event            EventType
	DotFirestartr    *dagger.Directory
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
	// +optional
	// +default=""
	app string,
	// State values directory (e.g. state-app-<app>#main)
	// +required
	valuesStateDir *dagger.Directory,
	// Wet state directory (e.g. wet-state-app-<app>#<deployment-branch>)
	// +required
	wetStateDir *dagger.Directory,
	// Auth directory
	// +required
	authDir *dagger.Directory,
	// Deployment branch to hydrate
	// +optional
	// +default="deployment"
	deploymentBranch string,
	// Event that triggered the workflow
	// +optional
	// +default="pr"
	event EventType,

	dotFirestartr *dagger.Directory,

) *HydrateOrchestrator {
	return &HydrateOrchestrator{
		Repo:             repo,
		GhToken:          ghToken,
		App:              app,
		ValuesStateDir:   valuesStateDir,
		WetStateDir:      wetStateDir,
		DeploymentBranch: deploymentBranch,
		AuthDir:          authDir,
		Event:            event,
		DotFirestartr:    dotFirestartr,
	}
}
