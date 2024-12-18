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

	repositoryCaller := m.getRepositoryCaller(newImagesMatrix)

	helmAuth := m.GetHelmAuth(ctx)

	for _, kdep := range deployments.KubernetesDeployments {

		branchName := fmt.Sprintf("kubernetes-%s-%s-%s", kdep.Cluster, kdep.Tenant, kdep.Environment)

		renderedDeployment := dag.HydrateKubernetes(
			m.ValuesStateDir,
			m.WetStateDir,
			dagger.HydrateKubernetesOpts{
				HelmRegistryLoginNeeded: helmAuth.NeedsAuth,
				HelmRegistry:            helmAuth.Registry,
				HelmRegistryUser:        helmAuth.Username,
				HelmRegistryPassword:    helmAuth.Password,
			},
		).Render(m.App, kdep.Cluster, kdep.Tenant, kdep.Environment, dagger.HydrateKubernetesRenderOpts{
			NewImagesMatrix: newImagesMatrix,
		})

		prBody := fmt.Sprintf(`
New deployment created by @author, from repository %s
%s
`, repositoryCaller, m.Event, kdep.String(true))

		m.upsertPR(
			ctx,
			branchName,
			renderedDeployment,
			kdep.Labels(),
			kdep.String(true),
			prBody,
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

func (m *HydrateOrchestrator) getRepositoryCaller(newImagesMatrix string) string {
	var imagesMatrix ImageMatrix
	err := json.Unmarshal([]byte(newImagesMatrix), &imagesMatrix)

	if err != nil {
		panic(err)
	}

	return imagesMatrix.Images[0].RepositoryCaller
}
