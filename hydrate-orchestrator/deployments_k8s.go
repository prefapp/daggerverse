package main

import (
	"context"
	"dagger/hydrate-orchestrator/internal/dagger"
	"encoding/json"
	"fmt"
	"path/filepath"
)

func (m *HydrateOrchestrator) GenerateKubernetesDeployments(
	ctx context.Context,
	newImagesMatrix string,
	repositoryCaller string,
	repoURL string,
	reviewers []string,
) (*dagger.File, error) {
	deployments := m.processImagesMatrix(newImagesMatrix)

	summary := &DeploymentSummary{
		Items: []DeploymentSummaryRow{},
	}

	for _, kdep := range deployments.KubernetesDeployments {

		// Extract main service name and list of updated services
		var mainServiceName string
		var updatedServices string
		var newImage string
		if len(kdep.ImagesMatrix) > 0 {
			var imgMatrix ImageMatrix
			_ = json.Unmarshal([]byte(kdep.ImagesMatrix), &imgMatrix)
			if len(imgMatrix.Images) > 0 {
				mainServiceName = imgMatrix.Images[0].App
				if len(imgMatrix.Images[0].ServiceNameList) > 0 {
					updatedServices = ""
					for _, svc := range imgMatrix.Images[0].ServiceNameList {
						updatedServices += fmt.Sprintf("  - **%s**\n", svc)
					}
				} else {
					updatedServices = fmt.Sprintf("  - **%s**\n", mainServiceName)
				}
				newImage = imgMatrix.Images[0].Image
			}
		}

		branchName := fmt.Sprintf("%s-kubernetes-%s-%s-%s", repositoryCaller, kdep.Cluster, kdep.Tenant, kdep.Environment)

		renderedDeployment, err := dag.HydrateKubernetes(
			m.ValuesStateDir,
			m.WetStateDir,
			m.DotFirestartr,
			dagger.HydrateKubernetesOpts{
				HelmConfigDir: m.AuthDir,
			},
		).Render(ctx, m.App, kdep.Cluster, dagger.HydrateKubernetesRenderOpts{
			Tenant:          kdep.Tenant,
			Env:             kdep.Environment,
			NewImagesMatrix: kdep.ImagesMatrix,
		})

		if err != nil {
			summary.addDeploymentSummaryRow(
				kdep.DeploymentPath,
				fmt.Sprintf("Failed: %s", err.Error()),
			)

			continue
		}

		title := fmt.Sprintf("Deployment of %s in cluster: %s, tenant: %s, env: %s", mainServiceName, kdep.Cluster, kdep.Tenant, kdep.Environment)
		prBody := fmt.Sprintf(`
### New automatic deployment triggered by new image built:
- Repository [*%s*](%s)
- Services updated:
%s- New image: %s

#### Deployment coordinates
* Cluster: %s
* Tenant: %s
* Environment: %s
`, repositoryCaller, repoURL, updatedServices, newImage, kdep.Cluster, kdep.Tenant, kdep.Environment)

		globPattern := fmt.Sprintf("%s/%s/%s/%s", "kubernetes", kdep.Cluster, kdep.Tenant, kdep.Environment)

		prLink, err := m.upsertPR(
			ctx,
			branchName,
			&renderedDeployment[0],
			kdep.Labels(),
			title,
			prBody,
			fmt.Sprintf("kubernetes/%s/%s/%s", kdep.Cluster, kdep.Tenant, kdep.Environment),
			reviewers,
		)

		if err != nil {

			summary.addDeploymentSummaryRow(
				kdep.DeploymentPath,
				fmt.Sprintf("Failed: %s", err.Error()),
			)

			continue
		}

		if m.AutomergeFileExists(ctx, globPattern) {

			fmt.Printf("Automerge file found, merging PR %s\n", prLink)

			if prLink == "" {

				summary.addDeploymentSummaryRow(
					kdep.DeploymentPath,
					"Failed: PR link is empty, cannot merge PR",
				)

				continue

			}

			err := m.MergePullRequest(ctx, prLink)

			if err != nil {

				summary.addDeploymentSummaryRow(
					kdep.DeploymentPath,
					fmt.Sprintf("Failed: %s", err.Error()),
				)

				continue

			}

			summary.addDeploymentSummaryRow(
				kdep.DeploymentPath,
				fmt.Sprintf(
					"Success, pr merged: <a href=\"%s\">%s</a>",
					prLink,
					prLink,
				),
			)

		} else {

			fmt.Println("Automerge file does not exist, skipping automerge")

			summary.addDeploymentSummaryRow(
				kdep.DeploymentPath,
				fmt.Sprintf(
					"Success, pr created: <a href=\"%s\">%s</a>",
					prLink,
					prLink,
				),
			)
		}

	}

	return m.DeploymentSummaryToFile(ctx, summary), nil
}

func (m *HydrateOrchestrator) processImagesMatrix(
	updatedDeployments string,
) *Deployments {
	result := &Deployments{
		KubernetesDeployments: []KubernetesAppDeployment{},
	}

	var imagesMatrix ImageMatrix
	err := json.Unmarshal([]byte(updatedDeployments), &imagesMatrix)

	if err != nil {
		panic(err)
	}

	for _, image := range imagesMatrix.Images {

		// At the moment the dispatch does not send the cluster so we extract it from the base folder

		deploymentPath := filepath.Join(
			"kubernetes",
			image.Platform,
			image.Tenant,
			image.Env,
		)

		uniqueImage := []ImageData{image}

		uniqueImageMatrix := ImageMatrix{
			Images: uniqueImage,
		}

		jsonUniqueImage, err := json.Marshal(uniqueImageMatrix)

		if err != nil {

			panic(err)

		}

		kdep := &KubernetesAppDeployment{
			Deployment: Deployment{
				DeploymentPath: deploymentPath,
			},
			Cluster:      image.Platform,
			Tenant:       image.Tenant,
			Environment:  image.Env,
			ImagesMatrix: string(jsonUniqueImage),
		}

		result.addDeployment(kdep)

	}

	return result
}
