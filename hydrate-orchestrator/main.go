package main

import (
	"context"
	"dagger/hydrate-orchestrator/internal/dagger"
	"encoding/json"
	"path/filepath"
)

type HydrateOrchestrator struct {
	WetRepoDir *dagger.Directory
}

func New(
	// The path to the wet repo directory, where the wet manifests are stored
	// +optional
	wetRepoDir *dagger.Directory,
) *HydrateOrchestrator {
	return &HydrateOrchestrator{}
}

func (m *HydrateOrchestrator) Run(
	ctx context.Context,
	app string,
	// The path to the values directory, where the helm values are stored
	valuesDir *dagger.Directory,
	repositoryDir *dagger.Directory,
	// extra packages to install
	// +optional
	depsFile *dagger.File,
	// +optional
	// +default="[]"
	updatedDeployments string,
	// +optional
	// +default="{\"images\":[]}"
	newImagesMatrix string,
) {

	// Load the updated deployments from JSON string using gojq
	var deployments []string
	json.Unmarshal([]byte(updatedDeployments), &deployments)

	// Iterate over the deployments
	kubernetesDeployments := make([]string, 0)

	for _, deployment := range deployments {
		// Get the first directory from the deployment string
		deploymentType := filepath.SplitList(deployment)[1]
		switch deploymentType {
		case "kubernetes":
			affectedJson, err := json.Marshal([]string{deployment})
			if err != nil {
				// skip the deployment if it can't be marshalled
				continue
			}
			dag.HydrateKubernetes(
				valuesDir,
				dagger.HydrateKubernetesOpts{
					DepsFile: depsFile,
				},
			).RenderApps(app, dagger.HydrateKubernetesRenderAppsOpts{
				AffectedPaths:   string(affectedJson),
				NewImagesMatrix: newImagesMatrix,
			})
			kubernetesDeployments = append(kubernetesDeployments, deployment)

		}
	}

}
