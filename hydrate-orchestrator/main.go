package main

import (
	"context"
	"dagger/hydrate-orchestrator/internal/dagger"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type HydrateOrchestrator struct {
}

func (m *HydrateOrchestrator) Run(
	ctx context.Context,
	// Github repository name <owner>/<repo>
	// +required
	repo string,
	// GitHub token
	// +required
	ghToken *dagger.Secret,
	app string,
	// The path to the values directory, where the helm values are stored
	valuesDir *dagger.Directory,
	wetRepoDir *dagger.Directory,
	// extra packages to install
	// +optional
	depsFile *dagger.File,
	// +optional
	// +default="[]"
	updatedDeployments string,
	// +optional
	// +default="{\"images\":[]}"
	newImagesMatrix string,
	// +optional
	// +default="deployment"
	depBranch string,
) {

	// Load the updated deployments from JSON string using gojq
	var deployments []string
	json.Unmarshal([]byte(updatedDeployments), &deployments)

	for _, deployment := range deployments {
		// Get the first directory from the deployment string
		deploymentType := strings.Split(deployment, "/")[0]
		switch deploymentType {
		case "kubernetes":

			affectedJson, err := json.Marshal([]string{deployment})
			if err != nil {
				// skip the deployment if it can't be marshalled
				continue
			}
			renderedDep := dag.HydrateKubernetes(
				valuesDir,
				dagger.HydrateKubernetesOpts{
					DepsFile:   depsFile,
					WetRepoDir: wetRepoDir,
				},
			).RenderApps(app, dagger.HydrateKubernetesRenderAppsOpts{
				AffectedPaths:   string(affectedJson),
				NewImagesMatrix: newImagesMatrix,
			})

			// split the path into a slice
			deploymentPath := strings.Split(deployment, "/")
			depType := deploymentPath[0]
			cluster := deploymentPath[1]
			tenant := deploymentPath[2]
			env := deploymentPath[3]

			branchName := fmt.Sprintf("%s-%s-%s-%s", depType, cluster, tenant, env)

			m.CreateRemoteBranch(ctx, wetRepoDir, branchName, ghToken)

			// Create each label
			labels := []string{
				fmt.Sprintf("type/%s", depType),
				fmt.Sprintf("app/%s", app),
				fmt.Sprintf("cluster/%s", cluster),
				fmt.Sprintf("tenant/%s", tenant),
				fmt.Sprintf("env/%s", env),
			}

			for _, label := range labels {
				dag.Gh(dagger.GhOpts{Token: ghToken}).Run(fmt.Sprintf("label create -R %s --force %s", repo, label), dagger.GhRunOpts{DisableCache: true}).Sync(ctx)
			}

			m.UpsertPR(ctx, repo, ghToken, branchName, depBranch, renderedDep)
		}
	}

}

func (m *HydrateOrchestrator) UpsertPR(
	ctx context.Context,
	// Github repository name <owner>/<repo>
	// +required
	repo string,
	// GitHub token
	// +required
	ghToken *dagger.Secret,
	// +required
	newBranchName string,
	// +required
	baseBranchName string,
	// +required
	contents *dagger.Directory,
) {
	contentsDirPath := "/contents"
	dag.Gh().Container(dagger.GhContainerOpts{Token: ghToken, Plugins: []string{"prefapp/gh-commit"}}).
		WithDirectory(contentsDirPath, contents, dagger.ContainerWithDirectoryOpts{
			Exclude: []string{".git"},
		}).
		WithWorkdir(contentsDirPath).
		WithEnvVariable("CACHE_BUSTER", time.Now().String()).
		WithExec([]string{
			"gh",
			"commit",
			"-R", repo,
			"-b", newBranchName,
		}).Sync(ctx)
	// Create a PR for the updated deployment
	dag.Gh().Run(
		fmt.Sprintf("pr create -R %s --base %s --title 'Update deployment' --body 'Update deployment' --head %s", repo, baseBranchName, newBranchName),
		dagger.GhRunOpts{DisableCache: true},
	).Sync(ctx)
}

func (m *HydrateOrchestrator) CreateRemoteBranch(

	ctx context.Context,
	// Base branch name
	// +required
	gitDir *dagger.Directory,
	// New branch name
	// +required
	newBranch string,
	// GitHub token
	// +required
	ghToken *dagger.Secret,
) {
	gitDirPath := "/git_dir"
	dag.Gh().Container(dagger.GhContainerOpts{
		Token: ghToken,
	}).
		WithDirectory(gitDirPath, gitDir).
		WithWorkdir(gitDirPath).
		WithEnvVariable("CACHE_BUSTER", time.Now().String()).
		WithExec([]string{
			"git",
			"checkout",
			"-b",
			newBranch,
		}, dagger.ContainerWithExecOpts{},
		).WithExec([]string{
		"git",
		"push",
		"origin",
		newBranch,
	}).Sync(ctx)
}
