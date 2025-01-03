package main

import (
	"context"
	"dagger/hydrate-orchestrator/internal/dagger"
	"encoding/json"
	"fmt"

	"github.com/samber/lo"
)

// Hydrate deployments based on the updated deployments
func (m *HydrateOrchestrator) GenerateDeployment(
	ctx context.Context,
	// Identifier that triggered the render, this could be a PR number or a workflow run id
	// +optional
	// +default=0
	id int,
	// Author of the PR
	// +optional
	// +default="author"
	author string,
	// Type of deployment
	// +required
	deploymentType string,
	// Cluster name
	// +required
	cluster string,
	// Tenant name
	// +required
	tenant string,
	// Environment name
	// +required
	environment string,
) *dagger.File {

	branchInfo := m.getBranchInfo(ctx)

	deployments := m.processDeploymentGlob(ctx, m.ValuesStateDir, deploymentType, cluster, tenant, environment)

	helmAuth := m.GetHelmAuth(ctx)

	summary := &DeploymentSummary{
		Items: []DeploymentSummaryRow{},
	}

	for _, kdep := range deployments.KubernetesDeployments {

		branchName := fmt.Sprintf("kubernetes-%s-%s-%s", kdep.Cluster, kdep.Tenant, kdep.Environment)

		renderedDeployment, err := dag.HydrateKubernetes(
			m.ValuesStateDir,
			m.WetStateDir,
			dagger.HydrateKubernetesOpts{
				HelmRegistryLoginNeeded: helmAuth.NeedsAuth,
				HelmRegistry:            helmAuth.Registry,
				HelmRegistryUser:        helmAuth.Username,
				HelmRegistryPassword:    helmAuth.Password,
			},
		).Render(ctx, m.App, kdep.Cluster, dagger.HydrateKubernetesRenderOpts{
			Tenant: kdep.Tenant,
			Env:    kdep.Environment,
		})

		if err != nil {
			summary.addDeploymentSummaryRow(
				fmt.Sprintf("kubernetes/%s/%s/%s", kdep.Cluster, kdep.Tenant, kdep.Environment),
				fmt.Sprintf("Failed: %s", err.Error()),
			)
			continue
		}
		prBody := fmt.Sprintf(`
# New deployment manually triggered
Created by @%s from %s within commit [%s](%s)
%s
`,
			author,
			branchInfo.Name,
			branchInfo.SHA,
			fmt.Sprintf("https://github.com/%s/commit/%s", m.Repo, branchInfo.SHA),
			kdep.String(false),
		)

		err = m.upsertPR(
			ctx,
			id,
			branchName,
			&renderedDeployment[0],
			kdep.Labels(),
			kdep.String(true),
			prBody,
			fmt.Sprintf("kubernetes/%s/%s/%s", kdep.Cluster, kdep.Tenant, kdep.Environment),
			lo.Ternary(author == "author", []string{}, []string{author}),
		)

		if err != nil {
			summary.addDeploymentSummaryRow(
				fmt.Sprintf("kubernetes/%s/%s/%s", kdep.Cluster, kdep.Tenant, kdep.Environment),
				"Failed",
			)
		} else {
			summary.addDeploymentSummaryRow(
				fmt.Sprintf("kubernetes/%s/%s/%s", kdep.Cluster, kdep.Tenant, kdep.Environment),
				"Success",
			)
		}
	}

	return m.DeploymentSummaryToFile(ctx, summary)

}

// Hydrate deployments based on the updated deployments
func (m *HydrateOrchestrator) ValidateChanges(
	ctx context.Context,
	// Updated deployments in JSON format
	// +required
	updatedDeployments string,
) {

	deployments := m.processUpdatedDeployments(updatedDeployments)

	helmAuth := m.GetHelmAuth(ctx)

	for _, kdep := range deployments.KubernetesDeployments {

		renderedDeployment, err := dag.HydrateKubernetes(
			m.ValuesStateDir,
			m.WetStateDir,
			dagger.HydrateKubernetesOpts{
				HelmRegistryLoginNeeded: helmAuth.NeedsAuth,
				HelmRegistry:            helmAuth.Registry,
				HelmRegistryUser:        helmAuth.Username,
				HelmRegistryPassword:    helmAuth.Password,
			},
		).Render(ctx, m.App, kdep.Cluster, dagger.HydrateKubernetesRenderOpts{
			Tenant: kdep.Tenant,
			Env:    kdep.Environment,
		})

		if err != nil {
			panic(err)
		}

		_, err = renderedDeployment[0].Sync(ctx)

		if err != nil {
			panic(err)
		}

	}
}

// Function that returns a deployment object from a type, cluster, tenant and environment considering glob patterns
func (m *HydrateOrchestrator) processDeploymentGlob(
	ctx context.Context,
	// Values state directory
	// +required
	valuesStateDir *dagger.Directory,
	// Type of deployment
	// +required
	deploymentType string,
	// Cluster name
	// +required
	cluster string,
	// Tenant name
	// +required
	tenant string,
	// Environment name
	// +required
	environment string,
) *Deployments {

	affected_files, err := valuesStateDir.Glob(ctx, fmt.Sprintf("%s/%s/%s/%s", deploymentType, cluster, tenant, environment))

	if err != nil {
		panic(err)
	}

	jsonString, err := json.Marshal(affected_files)
	if err != nil {
		panic(err)
	}
	return m.processUpdatedDeployments(string(jsonString))
}

// Process updated deployments and return all unique deployments after validating and processing them
func (m *HydrateOrchestrator) processUpdatedDeployments(
	// List of updated deployments in JSON format
	// +required
	updatedDeployments string,
) *Deployments {
	// Load the updated deployments from JSON string using gojq
	var deployments []string
	err := json.Unmarshal([]byte(updatedDeployments), &deployments)

	if err != nil {
		panic(err)
	}

	result := &Deployments{
		KubernetesDeployments: []KubernetesDeployment{},
	}

	for _, deployment := range deployments {

		dirs := splitPath(deployment)

		if len(dirs) == 0 {
			panic(fmt.Sprintf("Invalid deployment path provided: %s", deployment))
		}

		deploymentType := dirs[0]

		switch deploymentType {
		case "kubernetes":
			// Process kubernetes deployment
			if lo.Contains([]string{"repositories.yaml", "environments.yaml"}, dirs[1]) {
				continue
			}
			kdep := kubernetesDepFromStr(deployment)
			result.addDeployment(kdep)

		}

	}

	return result

}
