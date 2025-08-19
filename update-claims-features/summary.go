package main

import (
	"context"
	"dagger/update-claims-features/internal/dagger"
)

type UpdateSummaryRow struct {
	// Claim name
	// +required
	Claim string `json:"claim"`
	// Status of the update
	// +required
	Status string `json:"status"`
}

type UpdateSummary struct {
	Items []UpdateSummaryRow `json:"items"`
}

func (s *UpdateSummary) addUpdateSummaryRow(claim string, status string) {
	s.Items = append(s.Items, UpdateSummaryRow{Claim: claim, Status: status})
}

// Function that converts a UpdateSummary object to a markdown table
func (s UpdateSummary) UpdateSummaryToMarkdownTable() string {

	if len(s.Items) == 0 {
		return "There are no updates to display"
	}

	table := "<table><tr><th>Updated claim</th><th>Status</th></tr>"

	for _, item := range s.Items {

		table += "<tr><td>" + item.Claim + "</td><td>" + item.Status + "</td></tr>"

	}

	table += "</table>"

	return table

}

// Function that creates a dagger.File object from a UpdateSummary object
func (m *UpdateClaimsFeatures) DeploymentSummaryToFile(ctx context.Context, updateSummary *UpdateSummary) *dagger.File {
	// Convert the UpdateSummary object to a markdown table
	markdownTable := updateSummary.UpdateSummaryToMarkdownTable()

	path := "/update-summary.md"
	// Create a dagger.File object with the markdown table as its content
	return dag.Directory().WithNewFile(path, markdownTable).File(path)
}
