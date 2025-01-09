package main

import (
	"context"
	"dagger/hydrate-orchestrator/internal/dagger"
)

type DeploymentSummaryRow struct {
	// Deployment path
	// +required
	DeploymentPath string `json:"deploymentPath"`
	// Status of the deployment
	// +required
	Status string `json:"status"`
	// Description of the deployment
	// +required
	Description string `json:"description"`
}

type DeploymentSummary struct {
	Items []DeploymentSummaryRow `json:"items"`
}

func (s *DeploymentSummary) addDeploymentSummaryRow(deploymentPath string, status string, description string) {
	s.Items = append(s.Items, DeploymentSummaryRow{DeploymentPath: deploymentPath, Status: status, Description: description})
}

// Function that converts a DeploymentSummary object to a markdown table
func (s DeploymentSummary) DeploymentSummaryToMarkdownTable() string {

	if len(s.Items) == 0 {
		return "There are no deployments to display"
	}

	// Markdown table header
	table := "| Deployment Path | Status |\n"
	table += "| --- | --- |\n"

	// Iterate over each item in the DeploymentSummary object
	for _, item := range s.Items {
		table += "|" + item.DeploymentPath + "|" + item.Status + "|\n"
	}

	return table

}

// Function that creates a dagger.File object from a DeploymentSummary object
func (m *HydrateOrchestrator) DeploymentSummaryToFile(ctx context.Context, deploymentSummary *DeploymentSummary) *dagger.File {

	// Convert the DeploymentSummary object to a markdown table
	markdownTable := deploymentSummary.DeploymentSummaryToMarkdownTable()

	path := "/deployment-summary.md"
	// Create a dagger.File object with the markdown table as its content
	return dag.Directory().WithNewFile(path, markdownTable).File(path)
}
