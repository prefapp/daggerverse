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
	GhCliVersion     string
	ArtifactRef      string
	LocalGhCliPath   *dagger.File
}

func New(
	ctx context.Context,
	// Github repository name <owner>/<repo>
	// +required
	repo string,
	// GitHub token
	// +required
	ghToken *dagger.Secret,
	// State values directory (e.g. state-app-<app>#main)
	// +required
	valuesStateDir *dagger.Directory,
	// Wet state directory (e.g. wet-state-app-<app>#<deployment-branch>)
	// +required
	wetStateDir *dagger.Directory,
	// Auth directory
	// +optional
	// +defaultPath="/tmp"
	authDir *dagger.Directory,
	// Deployment branch to hydrate
	// +optional
	// +default="deployment"
	deploymentBranch string,
	// Event that triggered the workflow
	// +optional
	// +default="pr"
	event EventType,
	// .firestartr directory. It contains de org global configurations.
	// +required
	dotFirestartr *dagger.Directory,

	//Gh CLI Version
	// +optional
	// +default="v2.66.1"
	ghCliVersion string,

	// runner's gh dir path
	// +optional
	localGhCliPath *dagger.File,

) *HydrateOrchestrator {
	appName := ""
	appData, err := dag.FirestartrConfig(dotFirestartr).Apps(ctx)
	if err != nil {
		panic(err)
	}

	for _, app := range appData {
		appStateRepo, asrErr := app.StateRepo(ctx)

		if asrErr != nil {
			panic(asrErr)
		}

		if appStateRepo == repo {
			appName, _ = app.Name(ctx)
		}
	}

	return &HydrateOrchestrator{
		Repo:             repo,
		GhToken:          ghToken,
		App:              appName,
		ValuesStateDir:   valuesStateDir,
		WetStateDir:      wetStateDir,
		DeploymentBranch: deploymentBranch,
		AuthDir:          authDir,
		Event:            event,
		DotFirestartr:    dotFirestartr,
		GhCliVersion:     ghCliVersion,
		LocalGhCliPath:   localGhCliPath,
	}
}
