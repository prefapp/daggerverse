package main

import (
	"context"
	"dagger/hydrate-orchestrator/internal/dagger"
	"encoding/json"
	"fmt"
	"strings"
)

// JSON Types

type ImageMatrix struct {
	Images []ImageData `json:"images"`
}

type ImageData struct {
	Tenant           string   `json:"tenant"`
	App              string   `json:"app"`
	Env              string   `json:"env"`
	ServiceNameList  []string `json:"service_name_list"`
	ImageKeys        []string `json:"image_keys"`
	Image            string   `json:"image"`
	Reviewers        []string `json:"reviewers"`
	Platform         string   `json:"platform"`
	Claim            string   `json:"claim"`
	Technology       string   `json:"technology"`
	RepositoryCaller string   `json:"repository_caller"`
}

// run-dispatch is the main entry point for the hydrate orchestrator.
func (m *HydrateOrchestrator) RunDispatch(
	ctx context.Context,
	// +optional
	// +default="{\"images\":[]}"
	newImagesMatrix string,
	// +required
	// +default="kubernetes"
	platformType string,
) *dagger.File {

	repositoryCaller, repoURL := m.getRepositoryCaller(newImagesMatrix)

	reviewers := m.getReviewers(newImagesMatrix)

	var summaryFile *dagger.File
	var err error

	switch platformType {
	case "kubernetes":
		summaryFile, err = m.GenerateKubernetesDeployments(
			ctx,
			newImagesMatrix,
			repositoryCaller,
			repoURL,
			reviewers,
		)
	case "tfworkspaces":
		summaryFile, err = m.GenerateTfWorkspacesDeployments(
			ctx,
			newImagesMatrix,
			repositoryCaller,
			repoURL,
			reviewers,
		)
	default:
		panic(fmt.Sprintf("Platform type %s not supported", platformType))
	}

	if err != nil {

		panic(err)
	}

	return summaryFile

}

// getRepositoryCaller extracts the repository caller and constructs the repository URL from the new images matrix.
// It assumes the first image in the matrix contains the repository caller.
// The repository caller is expected to be in the format "org/repository".
func (m *HydrateOrchestrator) getRepositoryCaller(newImagesMatrix string) (string, string) {
	var imagesMatrix ImageMatrix
	err := json.Unmarshal([]byte(newImagesMatrix), &imagesMatrix)

	if err != nil {
		panic(err)
	}

	org := strings.Split(m.Repo, "/")[0]

	repositoryCaller := imagesMatrix.Images[0].RepositoryCaller

	repoURL := fmt.Sprintf("https://github.com/%s/%s", org, repositoryCaller)

	return repositoryCaller, repoURL
}

// getReviewers extracts the reviewers from the new images matrix.
// It filters out any reviewers that are bots (identified by the suffix "[bot]") and
// ensures that there are no duplicate reviewers.
// The reviewers are expected to be listed under each image in the matrix.
// It returns a slice of unique reviewers.
func (m *HydrateOrchestrator) getReviewers(newImagesMatrix string) []string {
	var imagesMatrix ImageMatrix
	err := json.Unmarshal([]byte(newImagesMatrix), &imagesMatrix)

	if err != nil {
		panic(err)
	}

	reviewers := []string{}
	for _, image := range imagesMatrix.Images {
		// filter reviewers to avoid duplicates
		for _, reviewer := range image.Reviewers {
			if !strings.HasSuffix(reviewer, "[bot]") {
				// [bot] reviewers are not supported by the GitHub CLI
				if !contains(reviewers, reviewer) {
					reviewers = append(reviewers, reviewer)
				}
			} else {
				fmt.Printf("☢️ Skipping bot reviewer: %s\n", reviewer)
			}
		}
	}
	return reviewers
}

// contains checks if a string slice contains a specific item.
func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
