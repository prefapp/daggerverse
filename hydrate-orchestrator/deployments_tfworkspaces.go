package main

import (
	"context"
	"dagger/hydrate-orchestrator/internal/dagger"
	"encoding/json"
	"fmt"
	"strings"
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
				extractErrorMessage(err),
			)

			continue
		}

		prBody := fmt.Sprintf(`
# New deployment from new image in repository [*%s*](%s)
%s
`, repositoryCaller, repoURL, tfDep.String(false))

		globPattern := fmt.Sprintf("%s/%s/%s/%s", "tfworkspaces", tfDep.ClaimName, "*", "*")

		labels := []LabelInfo{
			{
				Name:        "plan",
				Color:       "7E7C7A",
				Description: "Run terraform plan",
			},
		}

		output, err := m.upsertPR(
			ctx,
			branchName,
			&renderedDeployment[0],
			labels,
			tfDep.String(true),
			prBody,
			fmt.Sprintf("tfworkspaces/%s/%s/%s", tfDep.ClaimName, tfDep.Tenant, tfDep.Environment),
			reviewers,
			DEPLOYMENT_BRANCH_NAME,
		)

		if err != nil {

			if output != "" {
				summary.addDeploymentSummaryRow(
					tfDep.DeploymentPath,
					output,
				)

				continue
			}

			summary.addDeploymentSummaryRow(
				tfDep.DeploymentPath,
				extractErrorMessage(err),
			)

			continue

		}

		parts := strings.Split(output, "/")
		if !strings.HasPrefix(output, "https://github.com/") || len(parts) < 7 {
			summary.addDeploymentSummaryRow(
				tfDep.DeploymentPath,
				fmt.Sprintf("Invalid PR URL format: %s", output),
			)
			continue
		}

		// https://github.com/org/app-repo/pull/8
		// parts:    [https:, , github.com, org, app-repo, pull, 8]
		// positions:  0     1       2        3     4        5   6
		prNumber := parts[6]
		repo := parts[4]
		org := parts[3]
		fmt.Printf("🔗 Getting PR number from PR link\n")
		fmt.Printf("PR link: %s\n", output)
		fmt.Printf("PR number: %s\n", prNumber)
		fmt.Printf("Repo: %s\n", repo)
		fmt.Printf("Org: %s\n", org)

		updatedDir := dag.HydrateTfworkspaces(
			m.ValuesStateDir,
			&renderedDeployment[0],
			m.DotFirestartr,
		).AddPrAnnotationToCr(
			tfDep.ClaimName,
			prNumber,
			org,
			repo,
			&renderedDeployment[0],
		)

		_, err = dag.Gh().Commit(
			updatedDir,
			branchName,
			"Update deployments",
			m.GhToken,
			dagger.GhCommitOpts{
				BaseBranch: DEPLOYMENT_BRANCH_NAME,
				DeletePath: fmt.Sprintf("tfworkspaces/%s/%s/%s", tfDep.ClaimName, tfDep.Tenant, tfDep.Environment),
			},
		).Sync(ctx)

		if err != nil {
			summary.addDeploymentSummaryRow(
				tfDep.DeploymentPath,
				extractErrorMessage(err),
			)

			continue
		}

		if m.AutomergeFileExists(ctx, globPattern) {

			fmt.Printf("AUTO_MERGE file found, merging PR %s\n", output)

			if output == "" {

				summary.addDeploymentSummaryRow(
					tfDep.DeploymentPath,
					"Failed: PR link is empty, cannot merge PR",
				)

				continue

			}

			err := m.MergePullRequest(ctx, output)

			if err != nil {

				summary.addDeploymentSummaryRow(
					tfDep.DeploymentPath,
					extractErrorMessage(err),
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
					output,
					output,
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
					output,
					output,
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
			ClaimName:    image.Claim,
			ImagesMatrix: string(jsonUniqueImage),
			Tenant:       image.Tenant,
			Environment:  image.Env,
		}

		result.addDeployment(kdep)

	}

	return result
}
