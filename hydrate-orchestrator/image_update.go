package main

import (
	"context"
	"dagger/hydrate-orchestrator/internal/dagger"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

// JSON Types

type ImageMatrix struct {
	Images []ImageData `json:"images"`
}

type ImageData struct {
	Tenant           string   `json:"tenant"`
	App              string   `json:"app"`
	Env              string   `json:"env"`
	ServiceNameList  []string `json:"service_name_list"`
	Image            string   `json:"image"`
	Reviewers        []string `json:"reviewers"`
	BaseFolder       string   `json:"base_folder"`
	RepositoryCaller string   `json:"repository_caller"`
}

func (m *HydrateOrchestrator) RunDispatch(
	ctx context.Context,
	// +optional
	// +default="{\"images\":[]}"
	newImagesMatrix string,
	// // Pr that triggered the render
	// // +required
	// workflowRun int,
) {

	deployments := m.processImagesMatrix(newImagesMatrix)

	repositoryCaller, repoURL := m.getRepositoryCaller(newImagesMatrix)
	reviewers := m.getReviewers(newImagesMatrix)

	helmAuth := m.GetHelmAuth(ctx)

	for _, kdep := range deployments.KubernetesDeployments {

		branchName := fmt.Sprintf("%s-kubernetes-%s-%s-%s", repositoryCaller, kdep.Cluster, kdep.Tenant, kdep.Environment)

		renderedDeployment := dag.HydrateKubernetes(
			m.ValuesStateDir,
			m.WetStateDir,
			dagger.HydrateKubernetesOpts{
				HelmRegistryLoginNeeded: helmAuth.NeedsAuth,
				HelmRegistry:            helmAuth.Registry,
				HelmRegistryUser:        helmAuth.Username,
				HelmRegistryPassword:    helmAuth.Password,
			},
		).Render(m.App, kdep.Cluster, dagger.HydrateKubernetesRenderOpts{
			Tenant:          kdep.Tenant,
			Env:             kdep.Environment,
			NewImagesMatrix: newImagesMatrix,
		})

		prBody := fmt.Sprintf(`
# New deployment from new image in repository [*%s*](%s)
%s
`, repositoryCaller, repoURL, kdep.String(false))

		m.upsertPR(
			ctx,
			branchName,
			renderedDeployment,
			kdep.Labels(),
			kdep.String(true),
			prBody,
			fmt.Sprintf("kubernetes/%s/%s/%s", kdep.Cluster, kdep.Tenant, kdep.Environment),
			reviewers,
		)
	}

}

func (m *HydrateOrchestrator) processImagesMatrix(
	updatedDeployments string,
) *Deployments {
	result := &Deployments{
		KubernetesDeployments: []KubernetesDeployment{},
	}

	var imagesMatrix ImageMatrix
	err := json.Unmarshal([]byte(updatedDeployments), &imagesMatrix)

	if err != nil {
		panic(err)
	}

	for _, image := range imagesMatrix.Images {

		// At the moment the dispatch does not send the cluster so we extract it from the base folder
		cluster := strings.Split(image.BaseFolder, "/")[1]
		deploymentPath := filepath.Join(
			"kubernetes",
			cluster,
			image.Tenant,
			image.Env,
		)

		kdep := &KubernetesDeployment{
			Deployment: Deployment{
				DeploymentPath: deploymentPath,
			},
			Cluster:     cluster,
			Tenant:      image.Tenant,
			Environment: image.Env,
		}

		result.addDeployment(kdep)

	}

	return result
}

func (m *HydrateOrchestrator) getRepositoryCaller(newImagesMatrix string) (string, string) {
	var imagesMatrix ImageMatrix
	err := json.Unmarshal([]byte(newImagesMatrix), &imagesMatrix)

	if err != nil {
		panic(err)
	}

	org := strings.Split(m.Repo, "/")[0]

	repositoryCaller := imagesMatrix.Images[0].RepositoryCaller

	repoURL := fmt.Sprintf("https://github.com/%s/%s", org, repositoryCaller)

	return repositoryCaller, repoURL
}

func (m *HydrateOrchestrator) getReviewers(newImagesMatrix string) []string {
	var imagesMatrix ImageMatrix
	err := json.Unmarshal([]byte(newImagesMatrix), &imagesMatrix)

	if err != nil {
		panic(err)
	}

	reviewers := []string{}
	for _, image := range imagesMatrix.Images {
		reviewers = append(reviewers, image.Reviewers...)
	}
	return reviewers
}
