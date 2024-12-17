package main

import (
	"context"
	"dagger/hydrate-orchestrator/internal/dagger"
	"encoding/json"
	"fmt"
)

// Hydrate deployments based on the updated deployments
func (m *HydrateOrchestrator) RunChanges(
	ctx context.Context,
	// Updated deployments in JSON format
	// +required
	updatedDeployments string,
	// Pr that triggered the render
	// +required
	valuesPrNumber int,
) {

	deployments := m.processUpdatedDeployments(ctx, updatedDeployments)

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
		).Render(m.App, kdep.Cluster, kdep.Tenant, kdep.Environment)

		prBody := fmt.Sprintf(`
		New deployment created by @author, from PR #%d
		%s
		`, valuesPrNumber, kdep.String(true))

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

// Process updated deployments and return all unique deployments after validating and processing them
func (m *HydrateOrchestrator) processUpdatedDeployments(
	ctx context.Context,
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
			kdep := kubernetesDepFromStr(deployment)
			result.addDeployment(kdep)

		}

	}

	return result

}
