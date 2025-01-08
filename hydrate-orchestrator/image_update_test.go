package main

import (
	"encoding/json"
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
				BaseFolder:       "kubernetes/cluster-name",
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

	expectedDpl := KubernetesDeployment{
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
				BaseFolder:       "base",
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
				BaseFolder:       "base",
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
