package main

import (
	"context"
	"dagger/hydrate-orchestrator/internal/dagger"
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
	// Identifier that triggered the render, this could be a PR number or a workflow run id
	// +optional
	// +default=0
	id int,
	// Author of the PR
	// +optional
	// +default="author"
	author string,
) {

	deployments := m.processUpdatedDeployments(updatedDeployments)

	helmAuth := m.GetHelmAuth(ctx)

	for _, kdep := range deployments.KubernetesDeployments {

		branchName := fmt.Sprintf("%d-kubernetes-%s-%s-%s", id, kdep.Cluster, kdep.Tenant, kdep.Environment)

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
			Tenant: kdep.Tenant,
			Env:    kdep.Environment,
		})

		var prBody string
		if m.Event == PullRequest {
			prBody = fmt.Sprintf(`
# New deployment from PR #%d
Created by @%s
%s
`, id, author, kdep.String(false))
		} else if m.Event == Manual {
			prBody = fmt.Sprintf(`
# New deployment manually triggered
Created by @%s
%s
`, author, kdep.String(false))
		}

		m.upsertPR(
			ctx,
			branchName,
			renderedDeployment,
			kdep.Labels(),
			kdep.String(true),
			prBody,
			fmt.Sprintf("kubernetes/%s/%s/%s", kdep.Cluster, kdep.Tenant, kdep.Environment),
			lo.Ternary(author == "author", []string{}, []string{author}),
		)
	}

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
