package main

import (
	"context"
	"dagger/hydrate-orchestrator/internal/dagger"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

type HydrateOrchestrator struct {
	WetRepoDir *dagger.Directory
	ghToken    dagger.Secret
}

func New(

	// Github token
	// +required
	ghToken dagger.Secret,
	// The path to the wet repo directory, where the wet manifests are stored
	// +optional
	wetRepoDir *dagger.Directory,
) *HydrateOrchestrator {
	return &HydrateOrchestrator{
		WetRepoDir: wetRepoDir,
		ghToken:    ghToken,
	}
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

			// Create a branch for the updated deployment
			wetRepoPath := "/wet_repo"
			// split the path into a slice
			deploymentPath := strings.Split(deployment, "/")
			depType := deploymentPath[0]
			cluster := deploymentPath[1]
			tenant := deploymentPath[2]
			env := deploymentPath[3]

			branchName := fmt.Sprintf("%s-%s-%s-%s", depType, cluster, tenant, env)

			dag.Gh().Container(dagger.GhContainerOpts{
				Token: &m.ghToken,
			}).
				WithDirectory(wetRepoPath, m.WetRepoDir).
				WithWorkdir(wetRepoPath).
				WithExec([]string{
					"git",
					"checkout",
					"-b",
					branchName,
				}, dagger.ContainerWithExecOpts{},
				).WithExec([]string{
				"git",
				"push",
				"origin",
				branchName,
			})

			// Create each label
			dag.Gh(dagger.GhOpts{Token: &m.ghToken}).Run(fmt.Sprintf("label create --force type/%s", depType))
			dag.Gh(dagger.GhOpts{Token: &m.ghToken}).Run(fmt.Sprintf("label create --force app/%s", app))
			dag.Gh(dagger.GhOpts{Token: &m.ghToken}).Run(fmt.Sprintf("label create --force cluster/%s", cluster))
			dag.Gh(dagger.GhOpts{Token: &m.ghToken}).Run(fmt.Sprintf("label create --force tenant/%s", tenant))
			dag.Gh(dagger.GhOpts{Token: &m.ghToken}).Run(fmt.Sprintf("label create --force env/%s", env))

			// Create a PR for the updated deployment
			dag.Gh(dagger.GhOpts{Token: &m.ghToken}).Run(fmt.Sprintf("pr create --title 'Update deployment' --body 'Update deployment' --base main --head %s", branchName))

		}
	}

}
