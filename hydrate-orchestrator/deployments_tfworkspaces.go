package main

import (
	"context"
	"dagger/hydrate-orchestrator/internal/dagger"
	"encoding/json"
	"fmt"
)

func (m *HydrateOrchestrator) GenerateTfWorkspacesDeployments(
	ctx context.Context,
	newImagesMatrix string,
	repositoryCaller string,
	repoURL string,
	reviewers []string,
) (*dagger.File, error) {
	deployments := m.processImagesMatrixForTfworkspaces(newImagesMatrix)

	kind := "FirestartrTerraformWorkspace"

	summary := &DeploymentSummary{
		Items: []DeploymentSummaryRow{},
	}

	for _, tfDep := range deployments.TfWorkspaceDeployments {

		branchName := fmt.Sprintf("%s-kubernetes-%s", repositoryCaller, tfDep.ClaimName)

		renderedDeployment, err := dag.HydrateTfworkspaces(
			m.ValuesStateDir,
			m.WetStateDir,
			m.DotFirestartr,
		).Render(ctx, tfDep.ClaimName, m.App, dagger.HydrateTfworkspacesRenderOpts{
			NewImagesMatrix: tfDep.ImagesMatrix,
		})

		if err != nil {

			summary.addDeploymentSummaryRow(
				tfDep.DeploymentPath,
				fmt.Sprintf("Failed: %s", err.Error()),
			)

			continue
		}

		prBody := fmt.Sprintf(`
# New deployment from new image in repository [*%s*](%s)
%s
`, repositoryCaller, repoURL, tfDep.String(false))

		globPattern := fmt.Sprintf("%s/%s/%s/%s", "tfworkspaces", tfDep.ClaimName, "*", "*")

		prLink, err := m.upsertPR(
			ctx,
			branchName,
			&renderedDeployment[0],
			tfDep.Labels(),
			tfDep.String(true),
			prBody,
			fmt.Sprintf("tfworkspaces/%s/%s/%s", tfDep.ClaimName, tfDep.Tenant, tfDep.Environment),
			reviewers,
		)

		if err != nil {

			summary.addDeploymentSummaryRow(
				tfDep.DeploymentPath,
				fmt.Sprintf("Failed: %s", err.Error()),
			)

			continue

		}

		if m.AutomergeFileExists(ctx, globPattern) {

			fmt.Printf("Automerge file found, merging PR %s\n", prLink)

			if prLink == "" {

				summary.addDeploymentSummaryRow(
					tfDep.DeploymentPath,
					"Failed: PR link is empty, cannot merge PR",
				)

				continue

			}

			err := m.MergePullRequest(ctx, prLink)

			if err != nil {

				summary.addDeploymentSummaryRow(
					tfDep.DeploymentPath,
					fmt.Sprintf("Failed: %s", err.Error()),
				)

				continue

			}

			summary.addDeploymentSummaryRow(
				fmt.Sprintf("%s/%s.%s.yaml",
					tfDep.DeploymentPath,
					kind,
					tfDep.ClaimName,
				),
				fmt.Sprintf(
					"Success, pr merged: <a href=\"%s\">%s</a>",
					prLink,
					prLink,
				),
			)

		} else {

			fmt.Println("Automerge file does not exist, skipping automerge")

			summary.addDeploymentSummaryRow(
				fmt.Sprintf("%s/%s.%s.yaml",
					tfDep.DeploymentPath,
					kind,
					tfDep.ClaimName,
				),
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

func (m *HydrateOrchestrator) processImagesMatrixForTfworkspaces(
	updatedDeployments string,
) *Deployments {
	result := &Deployments{
		TfWorkspaceDeployments: []TfWorkspaceDeployment{},
	}

	var imagesMatrix ImageMatrix
	err := json.Unmarshal([]byte(updatedDeployments), &imagesMatrix)

	if err != nil {
		panic(err)
	}

	for _, image := range imagesMatrix.Images {

		uniqueImage := []ImageData{image}

		uniqueImageMatrix := ImageMatrix{
			Images: uniqueImage,
		}

		jsonUniqueImage, err := json.Marshal(uniqueImageMatrix)

		if err != nil {

			panic(err)

		}

		kdep := &TfWorkspaceDeployment{
			Deployment: Deployment{
				DeploymentPath: "tfworkspaces",
			},
			ClaimName:    image.Platform,
			ImagesMatrix: string(jsonUniqueImage),
			Tenant:       image.Tenant,
			Environment:  image.Env,
		}

		result.addDeployment(kdep)

	}

	return result
}
