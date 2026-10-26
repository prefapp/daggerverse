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
				extractErrorMessage(err),
			)
			continue
		}

		prBody := kdep.String(false, repoURL)

		globPattern := fmt.
			Sprintf("%s/%s/%s/%s", "kubernetes", kdep.Cluster, kdep.Tenant, kdep.Environment)

		labels := kubernetesAppDeploymentLabels(kdep.Cluster, kdep.Tenant, kdep.Environment)

		output, err := m.upsertPR(
			ctx,
			branchName,
			&renderedDeployment[0],
			labels,
			kdep.String(true, repoURL),
			prBody,
			fmt.Sprintf("kubernetes/%s/%s/%s", kdep.Cluster, kdep.Tenant, kdep.Environment),
			reviewers,
			"deployment",
		)

		if err != nil {
			if output != "" {
				summary.addDeploymentSummaryRow(
					kdep.DeploymentPath,
					output,
				)

				continue
			}

			summary.addDeploymentSummaryRow(
				kdep.DeploymentPath,
				extractErrorMessage(err),
			)

			continue
		}

		if m.AutomergeFileExists(ctx, globPattern) {

			fmt.Printf("AUTO_MERGE file found, merging PR %s\n", output)

			if output == "" {

				summary.addDeploymentSummaryRow(
					kdep.DeploymentPath,
					"Failed: PR link is empty, cannot merge PR",
				)
				continue
			}

			err := m.MergePullRequest(ctx, output)

			if err != nil {

				summary.addDeploymentSummaryRow(
					kdep.DeploymentPath,
					extractErrorMessage(err),
				)
				continue
			}

			summary.addDeploymentSummaryRow(
				kdep.DeploymentPath,
				fmt.Sprintf(
					"Success, pr merged: <a href=\"%s\">%s</a>",
					output,
					output,
				),
			)

		} else {

			fmt.Println("AUTO_MERGE file does not exist, skipping automerge")

			summary.addDeploymentSummaryRow(
				kdep.DeploymentPath,
				fmt.Sprintf(
					"Success, pr created: <a href=\"%s\">%s</a>",
					output,
					output,
				),
			)
		}

	}

	return m.DeploymentSummaryToFile(ctx, summary), nil
}

func kubernetesAppDeploymentLabels(cluster_name string, tenant_name string, env_name string) []LabelInfo {
	return []LabelInfo{
		{
			Name:        "type/kubernetes",
			Color:       getDefaultColorForDeploymentLabel("type/kubernetes"),
			Description: getDefaultDescriptionForDeploymentLabel("type/kubernetes"),
		},
		{
			Name:        fmt.Sprintf("cluster/%s", cluster_name),
			Color:       getDefaultColorForDeploymentLabel(fmt.Sprintf("cluster/%s", cluster_name)),
			Description: getDefaultDescriptionForDeploymentLabel(fmt.Sprintf("cluster/%s", cluster_name)),
		},
		{
			Name:        fmt.Sprintf("tenant/%s", tenant_name),
			Color:       getDefaultColorForDeploymentLabel(fmt.Sprintf("tenant/%s", tenant_name)),
			Description: getDefaultDescriptionForDeploymentLabel(fmt.Sprintf("tenant/%s", tenant_name)),
		},
		{
			Name:        fmt.Sprintf("env/%s", env_name),
			Color:       getDefaultColorForDeploymentLabel(fmt.Sprintf("env/%s", env_name)),
			Description: getDefaultDescriptionForDeploymentLabel(fmt.Sprintf("env/%s", env_name)),
		},
	}
}

func kubernetesSysServiceDeploymentLabels(cluster_name string, sys_service_name string) []LabelInfo {
	return []LabelInfo{
		{
			Name:        "type/kubernetes",
			Color:       getDefaultColorForDeploymentLabel("type/kubernetes"),
			Description: getDefaultDescriptionForDeploymentLabel("type/kubernetes"),
		},
		{
			Name:        fmt.Sprintf("cluster/%s", cluster_name),
			Color:       getDefaultColorForDeploymentLabel(fmt.Sprintf("cluster/%s", cluster_name)),
			Description: getDefaultDescriptionForDeploymentLabel(fmt.Sprintf("cluster/%s", cluster_name)),
		},
		{
			Name:        fmt.Sprintf("sys-service/%s", sys_service_name),
			Color:       getDefaultColorForDeploymentLabel(fmt.Sprintf("sys-service/%s", sys_service_name)),
			Description: getDefaultDescriptionForDeploymentLabel(fmt.Sprintf("sys-service/%s", sys_service_name)),
		},
	}
}

func (m *HydrateOrchestrator) processImagesMatrix(
	updatedDeployments string,
) *Deployments {
	var imagesMatrix ImageMatrix
	err := json.Unmarshal([]byte(updatedDeployments), &imagesMatrix)

	if err != nil {
		panic(err)
	}

	// Group images by deployment path (kubernetes/<platform>/<tenant>/<env>)
	groups := make(map[string]struct {
		images   []ImageData
		services map[string]bool
		caller   string
	})

	for _, image := range imagesMatrix.Images {
		deploymentPath := filepath.Join("kubernetes", image.Platform, image.Tenant, image.Env)
		g, ok := groups[deploymentPath]
		if !ok {
			g = struct {
				images   []ImageData
				services map[string]bool
				caller   string
			}{images: []ImageData{}, services: make(map[string]bool), caller: image.RepositoryCaller}
			groups[deploymentPath] = g
		}
		g.images = append(g.images, image)
		for _, s := range image.ServiceNameList {
			g.services[s] = true
		}
		if g.caller == "" {
			g.caller = image.RepositoryCaller
		}
	}

	result := &Deployments{
		KubernetesDeployments: []KubernetesAppDeployment{},
	}

	for path, g := range groups {
		// Build combined ServiceNames (dedup)
		combinedServices := make([]string, 0, len(g.services))
		for s := range g.services {
			combinedServices = append(combinedServices, s)
		}

		// Build combined ImagesMatrix JSON
		combinedMatrix := ImageMatrix{Images: g.images}
		jsonBytes, err := json.Marshal(combinedMatrix)
		if err != nil {
			panic(err)
		}

		kdep := &KubernetesAppDeployment{
			Deployment: Deployment{
				DeploymentPath: path,
			},
			Cluster:          g.images[0].Platform,
			Tenant:           g.images[0].Tenant,
			Environment:      g.images[0].Env,
			ImagesMatrix:     string(jsonBytes),
			ServiceNames:     combinedServices,
			RepositoryCaller: g.caller,
			Image:            g.images[0].Image,
		}

		result.addDeployment(kdep)
	}

	return result
}
