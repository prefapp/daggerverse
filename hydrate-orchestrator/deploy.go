package main

import (
	"context"
	"dagger/hydrate-orchestrator/internal/dagger"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
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
			var e *dagger.ExecError
			if errors.As(err, &e) {
					fmt.Println("FAILED")
					fmt.Println("STDOUT:", e.Stdout)
					fmt.Println("STDERR:", e.Stderr)
					fmt.Println("EXIT CODE:", e.ExitCode)
			}
			summary.addDeploymentSummaryRow(
				kdep.DeploymentPath,
				fmt.Sprintf("Failed: %s", "STDOUT" + e.Stdout + "." + "STDERR" + e.Stderr),
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
			summary.addDeploymentSummaryRow(
				kdep.DeploymentPath,
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
				fmt.Sprintf("Failed: %s", err.Error()),
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

		prLink, err := m.upsertPR(
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

			continue

		}

		// https://github.com/org/app-repo/pull/8
		// parts:    [https:, , github.com, org, app-repo, pull, 8]
		// positions:  0     1       2        3     4        5   6
		prNumber := strings.Split(prLink, "/")[6]
		repo := strings.Split(prLink, "/")[4]
		org := strings.Split(prLink, "/")[3]
		fmt.Printf("🔗 Getting PR number from PR link\n")
		fmt.Printf("PR link: %s\n", prLink)
		fmt.Printf("PR number: %s\n", prNumber)
		fmt.Printf("Repo: %s\n", repo)
		fmt.Printf("Org: %s\n", org)

		updatedDir := dag.HydrateTfworkspaces(
			m.ValuesStateDir,
			&renderedDep[0],
			m.DotFirestartr,
		).AddPrAnnotationToCr(
			tfDep.ClaimName,
			prNumber,
			org,
			repo,
			&renderedDep[0],
		)

		contentsDirPath := "/contents"

		_, err = dag.Gh(dagger.GhOpts{
			Version: m.GhCliVersion,
		}).Container(dagger.GhContainerOpts{
			Token:   m.GhToken,
			Plugins: []string{"prefapp/gh-commit"},
		}).WithDirectory(contentsDirPath, updatedDir, dagger.ContainerWithDirectoryOpts{
			Exclude: []string{".git"},
		}).WithWorkdir(contentsDirPath).
			WithEnvVariable("CACHE_BUSTER", time.Now().String()).
			WithExec([]string{
				"gh",
				"commit",
				"-R", m.Repo,
				"-b", branchName,
				"-m", "Update deployments",
				"--delete-path", fmt.Sprintf("tfworkspaces/%s/%s/%s", tfDep.ClaimName, tfDep.Tenant, tfDep.Environment),
			}).
			Sync(ctx)

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

	for _, secDep := range deployments.SecretsDeployment {
		branchName := fmt.Sprintf("secrets-%s-%s", secDep.Tenant, secDep.Environment)

		renderedDeployment, err := dag.HydrateSecrets(
			m.ValuesStateDir,
			m.WetStateDir,
			m.DotFirestartr,
		).Render(ctx, m.App, secDep.Tenant, secDep.Environment)

		if err != nil {
			summary.addDeploymentSummaryRow(
				secDep.DeploymentPath,
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
			secDep.String(false),
		)

		_, err = m.upsertPR(
			ctx,
			branchName,
			&renderedDeployment[0],
			secDep.Labels(),
			secDep.String(true),
			prBody,
			secDep.DeploymentPath,
			lo.Ternary(author == "author", []string{}, []string{author}),
		)

		if err != nil {
			summary.addDeploymentSummaryRow(
				secDep.DeploymentPath,
				fmt.Sprintf("Failed: %s", err.Error()),
			)

		} else {
			summary.addDeploymentSummaryRow(
				secDep.DeploymentPath,
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

	deployments := m.processUpdatedDeployments(ctx, updatedDeployments)

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

	return m.processUpdatedDeployments(ctx, string(jsonString))
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
		KubernetesDeployments:    []KubernetesAppDeployment{},
		KubernetesSysDeployments: []KubernetesSysDeployment{},
		SecretsDeployment:        []SecretsDeployment{},
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
			tfDep := &TfWorkspaceDeployment{
				Deployment: Deployment{
					DeploymentPath: deployment,
				},
				ClaimName: m.ArtifactRef,
			}

			if strings.Trim(m.ArtifactRef, " ") == "" && m.Event == Manual {
				panic(fmt.Sprintf("error: your input artifact ref %s is empty", m.ArtifactRef))
			}

			if m.ArtifactRef != "" && strings.HasSuffix(deployment, ".yaml") {
				content, err := m.ValuesStateDir.File(deployment).Contents(ctx)
				if err != nil {
					panic(err)
				}
				claim := &Claim{}
				yaml.Unmarshal([]byte(content), claim)

				if claim.Name == m.ArtifactRef {
					result.addDeployment(tfDep)
				}
			} else {
				result.addDeployment(tfDep)
			}
		case "secrets":
			secDep := secretsDepFromStr(deployment)

			result.addDeployment(secDep)
		}

	}

	return result

}
