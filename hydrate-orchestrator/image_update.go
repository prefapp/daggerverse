package main

import (
	"context"
	"dagger/hydrate-orchestrator/internal/dagger"
	"encoding/json"
	"fmt"
	"path/filepath"
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
	Image            string   `json:"image"`
	Reviewers        []string `json:"reviewers"`
	Platform         string   `json:"platform"`
	Technology       string   `json:"technology"`
	RepositoryCaller string   `json:"repository_caller"`
}

func (m *HydrateOrchestrator) RunDispatch(
	ctx context.Context,
	// Workflow run id
	// +required
	id string,
	// +optional
	// +default="{\"images\":[]}"
	newImagesMatrix string,
	// // Pr that triggered the render
	// // +required
	// workflowRun int,
) {

	deployments := m.processImagesMatrix(newImagesMatrix)

	repositoryCaller, repoURL := m.getRepositoryCaller(newImagesMatrix)

	reviewers := m.getReviewers(newImagesMatrix)

	for _, kdep := range deployments.KubernetesDeployments {

		branchName := fmt.Sprintf("%s-kubernetes-%s-%s-%s", repositoryCaller, kdep.Cluster, kdep.Tenant, kdep.Environment)

		renderedDeployment, err := dag.HydrateKubernetes(
			m.ValuesStateDir,
			m.WetStateDir,
			m.DotFirestartr,
			dagger.HydrateKubernetesOpts{
				HelmConfigDir: m.AuthDir,
			},
		).Render(ctx, m.App, kdep.Cluster, dagger.HydrateKubernetesRenderOpts{
			Tenant:          kdep.Tenant,
			Env:             kdep.Environment,
			NewImagesMatrix: kdep.ImagesMatrix,
		})

		if err != nil {
			panic(err)
		}

		prBody := fmt.Sprintf(`
# New deployment from new image in repository [*%s*](%s)
%s
`, repositoryCaller, repoURL, kdep.String(false))

		globPattern := fmt.
			Sprintf("%s/%s/%s/%s", "kubernetes", kdep.Cluster, kdep.Tenant, kdep.Environment)

		prLink, err := m.upsertPR(
			ctx,
			"",
			branchName,
			&renderedDeployment[0],
			kdep.Labels(),
			kdep.String(true),
			prBody,
			fmt.Sprintf("kubernetes/%s/%s/%s", kdep.Cluster, kdep.Tenant, kdep.Environment),
			reviewers,
		)

		if err != nil {

			panic(err)
		}

		if m.AutomergeFileExists(ctx, globPattern) {

			fmt.Printf("Automerge file found, merging PR %s\n", prLink)

			if prLink == "" {

				panic("PR link is empty, cannot merge PR")

			}

			err := m.MergePullRequest(ctx, prLink)

			if err != nil {

				panic(err)

			}

		} else {

			fmt.Println("Automerge file does not exist, skipping automerge")

		}

	}

}

func (m *HydrateOrchestrator) processImagesMatrix(
	updatedDeployments string,
) *Deployments {
	result := &Deployments{
		KubernetesDeployments: []KubernetesAppDeployment{},
	}

	var imagesMatrix ImageMatrix
	err := json.Unmarshal([]byte(updatedDeployments), &imagesMatrix)

	if err != nil {
		panic(err)
	}

	for _, image := range imagesMatrix.Images {

		// At the moment the dispatch does not send the cluster so we extract it from the base folder

		deploymentPath := filepath.Join(
			"kubernetes",
			image.Platform,
			image.Tenant,
			image.Env,
		)

		uniqueImage := []ImageData{image}

		uniqueImageMatrix := ImageMatrix{
			Images: uniqueImage,
		}

		jsonUniqueImage, err := json.Marshal(uniqueImageMatrix)

		if err != nil {

			panic(err)

		}

		kdep := &KubernetesAppDeployment{
			Deployment: Deployment{
				DeploymentPath: deploymentPath,
			},
			Cluster:      image.Platform,
			Tenant:       image.Tenant,
			Environment:  image.Env,
			ImagesMatrix: string(jsonUniqueImage),
		}

		result.addDeployment(kdep)

	}

	return result
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
