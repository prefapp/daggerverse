package main

import (
	"encoding/json"
	"reflect"
	"slices"
	"testing"
)

func TestProcessImagesMatrix(t *testing.T) {

	m := &HydrateOrchestrator{
		Repo: "org/repo",
	}

	matrix := ImageMatrix{
		Images: []ImageData{
			{
				Tenant:           "tenant",
				App:              "app",
				Env:              "env",
				ServiceNameList:  []string{"service"},
				Image:            "image",
				Reviewers:        []string{"reviewer"},
				Platform:         "cluster-name",
				Technology:       "kubernetes",
				RepositoryCaller: "caller",
			},
		},
	}

	// Serialize the matrix to JSON string
	newImagesMatrix, err := json.Marshal(matrix)
	if err != nil {
		t.Fatalf("failed to marshal matrix: %v", err)
	}

	deployments := m.processImagesMatrix(string(newImagesMatrix))

	expectedDpl := KubernetesAppDeployment{
		Deployment: Deployment{
			DeploymentPath: "kubernetes/cluster-name/tenant/env",
		},
		Cluster:     "cluster-name",
		Tenant:      "tenant",
		Environment: "env",
	}

	if len(deployments.KubernetesDeployments) != 1 {
		t.Fatalf("expected 1 deployment, got %d", len(deployments.KubernetesDeployments))
	}

	if !expectedDpl.Equals(deployments.KubernetesDeployments[0]) {
		t.Fatalf("expected %v, got %v", expectedDpl, deployments.KubernetesDeployments[0])
	}

}

func TestProcessImagesMatrixGroups(t *testing.T) {
	for _, secondEnv := range []string{"dev", "prod"} {
		t.Run(secondEnv, func(t *testing.T) {
			matrix := ImageMatrix{Images: []ImageData{
				{
					Tenant: "test-tenant", App: "sample-app", Env: "dev",
					Platform: "cluster-name", ServiceNameList: []string{"micro-a"},
					Image: "registry/image:tag", RepositoryCaller: "caller",
				},
				{
					Tenant: "test-tenant", App: "sample-app", Env: secondEnv,
					Platform: "cluster-name", ServiceNameList: []string{"micro-b"},
					Image: "registry/image:tag", RepositoryCaller: "caller",
				},
			}}
			input, err := json.Marshal(matrix)
			if err != nil {
				t.Fatal(err)
			}

			expected := map[string][]ImageData{
				"kubernetes/cluster-name/test-tenant/dev": {matrix.Images[0]},
			}
			secondPath := "kubernetes/cluster-name/test-tenant/" + secondEnv
			expected[secondPath] = append(expected[secondPath], matrix.Images[1])

			m := &HydrateOrchestrator{}
			deployments := m.processImagesMatrix(string(input)).KubernetesDeployments
			if len(deployments) != len(expected) {
				t.Fatalf("expected %d deployments, got %d", len(expected), len(deployments))
			}
			for _, deployment := range deployments {
				images, ok := expected[deployment.DeploymentPath]
				if !ok {
					t.Fatalf("unexpected deployment path: %s", deployment.DeploymentPath)
				}
				delete(expected, deployment.DeploymentPath)
				var combined ImageMatrix
				if err := json.Unmarshal([]byte(deployment.ImagesMatrix), &combined); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(combined.Images, images) {
					t.Errorf("expected images %v, got %v", images, combined.Images)
				}
				var services []string
				for _, image := range images {
					services = append(services, image.ServiceNameList...)
				}
				slices.Sort(services)
				slices.Sort(deployment.ServiceNames)
				if !slices.Equal(deployment.ServiceNames, services) {
					t.Errorf("expected services %v, got %v", services, deployment.ServiceNames)
				}
				first := images[0]
				if deployment.Cluster != first.Platform || deployment.Tenant != first.Tenant ||
					deployment.Environment != first.Env || deployment.RepositoryCaller != first.RepositoryCaller {
					t.Errorf("deployment metadata does not match first image: %+v", deployment)
				}
			}
		})
	}
}

func TestGetRepositoryCaller(t *testing.T) {

	m := &HydrateOrchestrator{
		Repo: "org/repo",
	}

	matrix := ImageMatrix{
		Images: []ImageData{
			{
				Tenant:           "tenant",
				App:              "app",
				Env:              "env",
				ServiceNameList:  []string{"service"},
				Image:            "image",
				Reviewers:        []string{"reviewer"},
				Platform:         "cluster-name",
				Technology:       "kubernetes",
				RepositoryCaller: "caller",
			},
		},
	}

	// Serialize the matrix to JSON string
	newImagesMatrix, err := json.Marshal(matrix)
	if err != nil {
		t.Fatalf("failed to marshal matrix: %v", err)
	}

	repositoryCaller, repoURL := m.getRepositoryCaller(string(newImagesMatrix))

	if repositoryCaller != "caller" {
		t.Fatalf("expected caller to be 'caller', got %s", repositoryCaller)
	}

	if repoURL != "https://github.com/org/caller" {
		t.Fatalf("expected repoURL to be 'https://github.com/org/caller', got %s", repoURL)
	}

}

func TestGetReviewers(t *testing.T) {
	m := &HydrateOrchestrator{
		Repo: "org/repo",
	}

	matrix := ImageMatrix{
		Images: []ImageData{
			{
				Tenant:           "tenant",
				App:              "app",
				Env:              "env",
				ServiceNameList:  []string{"service"},
				Image:            "image",
				Reviewers:        []string{"reviewer"},
				Platform:         "cluster-name",
				Technology:       "kubernetes",
				RepositoryCaller: "caller",
			},
		},
	}

	// Serialize the matrix to JSON string
	newImagesMatrix, err := json.Marshal(matrix)
	if err != nil {
		t.Fatalf("failed to marshal matrix: %v", err)
	}

	reviewers := m.getReviewers(string(newImagesMatrix))

	if len(reviewers) != 1 {
		t.Fatalf("expected reviewers to have 1 element, got %d", len(reviewers))
	}

	matrix.Images[0].Reviewers = append(matrix.Images[0].Reviewers, "reviewer2")

	// Serialize the matrix to JSON string
	newImagesMatrix, err = json.Marshal(matrix)
	if err != nil {
		t.Fatalf("failed to marshal matrix: %v", err)
	}

	reviewers = m.getReviewers(string(newImagesMatrix))

	if len(reviewers) != 2 {
		t.Fatalf("expected reviewers to have 2 elements, got %d", len(reviewers))
	}

	if reviewers[0] != "reviewer" {
		t.Fatalf("expected reviewers[0] to be 'reviewer', got %s", reviewers[0])
	}

	if reviewers[1] != "reviewer2" {
		t.Fatalf("expected reviewers[1] to be 'reviewer2', got %s", reviewers[1])
	}
}
