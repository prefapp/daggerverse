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
	// Author of the PR
	// +optional
	// +default="author"
	author string,
	// Glob Pattern
	// +required
	globPattern string,
	// Aritfact ref. This param could be used to reference the artifact that triggered the deployment
	// It contains the image tag, sha, etc.
	// +optional
	// +default=""
	artifactRef string,
) *dagger.File {

	m.ArtifactRef = artifactRef

	branchInfo := m.getBranchInfo(ctx)

	summary := &DeploymentSummary{
		Items: []DeploymentSummaryRow{},
	}

	deployments := m.processDeploymentGlob(ctx, m.ValuesStateDir, globPattern)

	for _, kdep := range deployments.KubernetesDeployments {

		branchName := fmt.Sprintf("kubernetes-%s-%s-%s", kdep.Cluster, kdep.Tenant, kdep.Environment)

		renderedDeployment, err := dag.HydrateKubernetes(
			m.ValuesStateDir,
			m.WetStateDir,
			m.DotFirestartr,
			dagger.HydrateKubernetesOpts{
				HelmConfigDir: m.AuthDir,
			},
		).Render(ctx, m.App, kdep.Cluster, dagger.HydrateKubernetesRenderOpts{
			Tenant: kdep.Tenant,
			Env:    kdep.Environment,
		})

		if err != nil {
			summary.addDeploymentSummaryRow(
				kdep.DeploymentPath,
				err.Error(),
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

		_, err = m.upsertPR(
			ctx,
			branchName,
			&renderedDeployment[0],
			kdep.Labels(),
			kdep.String(true),
			prBody,
			kdep.DeploymentPath,
			lo.Ternary(author == "author", []string{}, []string{author}),
		)

		if err != nil {
			summary.addDeploymentSummaryRow(
				kdep.DeploymentPath,
				fmt.Sprintf("Failed: %s", err.Error()),
			)

		} else {
			summary.addDeploymentSummaryRow(
				kdep.DeploymentPath,
				"Success",
			)
		}
	}

	for _, kdep := range deployments.KubernetesSysDeployments {
		branchName := fmt.Sprintf("kubernetes-sys-services-%s-%s", kdep.Cluster, kdep.SysServiceName)

		renderedDeployment, err := dag.HydrateKubernetes(
			m.ValuesStateDir,
			m.WetStateDir,
			m.DotFirestartr,
			dagger.HydrateKubernetesOpts{
				HelmConfigDir: m.AuthDir,
				RenderType:    "sys-services",
			},
		).Render(ctx, kdep.SysServiceName, kdep.Cluster)

		if err != nil {
			summary.addDeploymentSummaryRow(kdep.DeploymentPath, err.Error())

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

		_, err = m.upsertPR(
			ctx,
			branchName,
			&renderedDeployment[0],
			kdep.Labels(),
			kdep.String(true),
			prBody,
			kdep.DeploymentPath,
			lo.Ternary(author == "author", []string{}, []string{author}),
		)

		if err != nil {
			summary.addDeploymentSummaryRow(
				kdep.DeploymentPath,
				fmt.Sprintf("Failed: %s", err.Error()),
			)

		} else {
			summary.addDeploymentSummaryRow(
				kdep.DeploymentPath,
				"Success",
			)
		}

	}

	for _, tfDep := range deployments.TfWorkspaceDeployments {

		renderedDep, err := dag.
			HydrateTfworkspaces(
				m.ValuesStateDir,
				m.WetStateDir,
				m.DotFirestartr,
			).
			Render(ctx, tfDep.ClaimName, m.App)

		if err != nil {
			summary.addDeploymentSummaryRow(
				tfDep.DeploymentPath,
				err.Error(),
			)

			continue
		}

		branchName := fmt.Sprintf("tfworkspaces-%s", tfDep.ClaimName)

		prBody := fmt.Sprintf(`
# New deployment manually triggered
Created by @%s from %s within commit [%s](%s)
%s
`,
			author,
			branchInfo.Name,
			branchInfo.SHA,
			fmt.Sprintf("https://github.com/%s/commit/%s", m.Repo, branchInfo.SHA),
			tfDep.String(false),
		)

		_, err = m.upsertPR(
			ctx,
			branchName,
			&renderedDep[0],
			tfDep.Labels(),
			tfDep.String(true),
			prBody,
			tfDep.DeploymentPath,
			lo.Ternary(author == "author", []string{}, []string{author}),
		)

		if err != nil {

			summary.addDeploymentSummaryRow(
				tfDep.DeploymentPath,
				fmt.Sprintf("Failed: %s", err.Error()),
			)

		} else {
			summary.addDeploymentSummaryRow(
				tfDep.DeploymentPath,
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

	for _, kdep := range deployments.KubernetesDeployments {

		renderedDeployment, err := dag.HydrateKubernetes(
			m.ValuesStateDir,
			m.WetStateDir,
			m.DotFirestartr,
			dagger.HydrateKubernetesOpts{
				HelmConfigDir: m.AuthDir,
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

	for _, kdep := range deployments.KubernetesSysDeployments {

		renderedDeployment, err := dag.HydrateKubernetes(
			m.ValuesStateDir,
			m.WetStateDir,
			m.DotFirestartr,
			dagger.HydrateKubernetesOpts{
				HelmConfigDir: m.AuthDir,
				RenderType:    "sys-services",
			},
		).Render(ctx, kdep.SysServiceName, kdep.Cluster)

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
	globPattern string,

) *Deployments {

	affected_files, err := valuesStateDir.Glob(ctx, globPattern)

	if len(affected_files) == 0 {
		panic(
			fmt.Sprintf("error: your input glob pattern %s did not match any files", globPattern),
		)
	}

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
		KubernetesDeployments:    []KubernetesAppDeployment{},
		KubernetesSysDeployments: []KubernetesSysDeployment{},
	}

	for _, deployment := range deployments {

		dirs := splitPath(deployment)

		if len(dirs) == 0 {
			panic(fmt.Sprintf("Invalid deployment path provided (dir count is zeri): %s", deployment))
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

		case "kubernetes-sys-services":
			// Process kubernetes sys service deployment
			if lo.Contains([]string{"repositories.yaml", "environments.yaml"}, dirs[1]) {
				continue
			}
			kdep := kubernetesSysDepFromStr(deployment)
			result.addDeployment(kdep)
		case "tfworkspaces":
			// Process terraform workspace deployment
			tfDep := &TfWorkspaceDeployment{
				Deployment: Deployment{
					DeploymentPath: deployment,
				},
				ClaimName: m.ArtifactRef,
			}

			result.addDeployment(tfDep)
		}

	}

	return result

}
