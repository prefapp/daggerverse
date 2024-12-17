package main

import (
	"context"
	"dagger/hydrate-orchestrator/internal/dagger"
	"encoding/json"
	"fmt"
	"strings"
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
		dag.HydrateKubernetes(
			m.ValuesStateDir,
			m.WetStateDir,
		).Render(m.App, kdep.Cluster, kdep.Tenant, kdep.Environment)

	}

	for _, deployment := range deployments {

			branchName := fmt.Sprintf("%s-%s-%s-%s", depType, cluster, tenant, env)

			prExists := m.CheckPrExists(ctx, repo, branchName, ghToken)
			if !prExists {

				m.CreateRemoteBranch(ctx, wetRepoDir, branchName, ghToken)
			}

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
