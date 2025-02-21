package main

import (
	"context"
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
	Technology       string   `json:"technology"`
	RepositoryCaller string   `json:"repository_caller"`
}

func (m *HydrateOrchestrator) RunDispatch(
	ctx context.Context,
	// +optional
	// +default="{\"images\":[]}"
	newImagesMatrix string,
	// +required
	// +default="kubernetes"
	platformType string,
) {
	repositoryCaller, repoURL := m.getRepositoryCaller(newImagesMatrix)

	reviewers := m.getReviewers(newImagesMatrix)

	switch platformType {
	case "kubernetes":
		m.GenerateKubernetesDeployments(
			ctx,
			newImagesMatrix,
			repositoryCaller,
			repoURL,
			reviewers,
		)
	case "tfworkspaces":
		m.GenerateTfWorkspacesDeployments(
			ctx,
			newImagesMatrix,
			repositoryCaller,
			repoURL,
			reviewers,
		)
	default:
		panic(fmt.Sprintf("Platform type %s not supported", platformType))
	}

}

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

func (m *HydrateOrchestrator) getReviewers(newImagesMatrix string) []string {
	var imagesMatrix ImageMatrix
	err := json.Unmarshal([]byte(newImagesMatrix), &imagesMatrix)

	if err != nil {
		panic(err)
	}

	reviewers := []string{}
	for _, image := range imagesMatrix.Images {
		reviewers = append(reviewers, image.Reviewers...)
	}
	return reviewers
}
