package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/samber/lo"
)

// Hydrate deployments based on the updated deployments
func (m *HydrateOrchestrator) RunChanges(
	ctx context.Context,
	// Updated deployments in JSON format
	// +required
	updatedDeployments string,
) {

	deployments := m.processUpdatedDeployments(ctx, updatedDeployments)

	for _, kdep := range deployments.KubernetesDeployments {

		// renderedDeployment

		branchName := fmt.Sprintf("kubernetes-%s-%s-%s", kdep.Cluster, kdep.Tenant, kdep.Environment)
		renderedDeployment := dag.HydrateKubernetes(
			m.ValuesStateDir,
			m.WetStateDir,
		).Render(m.App, kdep.Cluster, kdep.Tenant, kdep.Environment)

		m.upsertPR(ctx, branchName, renderedDeployment, []string{})
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
			dep := kubernetesDepFromStr(deployment)
			if !lo.ContainsBy(result.KubernetesDeployments, func(d KubernetesDeployment) bool {
				return d.Equals(*dep)
			}) {
				result.KubernetesDeployments = append(result.KubernetesDeployments, *dep)
			}
		}

	}

	return result

}
