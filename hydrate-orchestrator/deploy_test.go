package main

import (
	"encoding/json"
	"testing"
)

func TestProcessedUpdatedDeployments(t *testing.T) {

	m := &HydrateOrchestrator{}

	// should ignore these paths
	paths := []string{"kubernetes/environments.yaml", "kubernetes/repositories.yaml"}

	pathsJSON, err := json.Marshal(paths)
	if err != nil {
		t.Fatalf("Error marshalling paths to JSON: %v", err)
	}

	deployments := m.processUpdatedDeployments(string(pathsJSON))

	if len(deployments.KubernetesDeployments) != 0 {
		t.Errorf("Expected 0 deployments, got %v", len(deployments.KubernetesDeployments))
	}

	// should process this path
	paths = []string{"kubernetes/environments.yaml", "kubernetes/sample-cluster/sample-tenant/sample-env/foo.yaml"}

	pathsJSON, err = json.Marshal(paths)
	if err != nil {
		t.Fatalf("Error marshalling paths to JSON: %v", err)
	}

	deployments = m.processUpdatedDeployments(string(pathsJSON))

	if len(deployments.KubernetesDeployments) != 1 {
		t.Errorf("Expected 1 deployment, got %v", len(deployments.KubernetesDeployments))
	}

	expectedDp := KubernetesAppDeployment{
		Deployment: Deployment{
			DeploymentPath: "kubernetes/sample-cluster/sample-tenant/sample-env",
		},
		Cluster:     "sample-cluster",
		Tenant:      "sample-tenant",
		Environment: "sample-env",
	}

	if expectedDp.Equals(deployments.KubernetesDeployments[0]) == false {
		t.Errorf("Expected %v, got %v", expectedDp, deployments.KubernetesDeployments[0])
	}

	// Should detect <cluster>/<tenant>/<env>.yaml as valid deployment path
	paths = []string{"kubernetes/sample-cluster/sample-tenant/sample-env.yaml"}

	pathsJSON, err = json.Marshal(paths)
	if err != nil {
		t.Fatalf("Error marshalling paths to JSON: %v", err)
	}

	deployments = m.processUpdatedDeployments(string(pathsJSON))

	if len(deployments.KubernetesDeployments) != 1 {
		t.Errorf("Expected 1 deployment, got %v", len(deployments.KubernetesDeployments))
	}

	if expectedDp.Equals(deployments.KubernetesDeployments[0]) == false {
		t.Errorf("Expected %v, got %v", expectedDp, deployments.KubernetesDeployments[0])
	}

	// Should detect multiple values files as a single deployment
	paths = []string{"kubernetes/sample-cluster/sample-tenant/sample-env/foo.yaml", "kubernetes/sample-cluster/sample-tenant/sample-env/bar.yaml"}

	pathsJSON, err = json.Marshal(paths)
	if err != nil {
		t.Fatalf("Error marshalling paths to JSON: %v", err)
	}

	deployments = m.processUpdatedDeployments(string(pathsJSON))

	if len(deployments.KubernetesDeployments) != 1 {
		t.Errorf("Expected 1 deployment, got %v", len(deployments.KubernetesDeployments))
	}

	if expectedDp.Equals(deployments.KubernetesDeployments[0]) == false {
		t.Errorf("Expected %v, got %v", expectedDp, deployments.KubernetesDeployments[0])
	}

	// should detect multiple deployments
	paths = []string{"kubernetes/sample-cluster/sample-tenant/sample-env/foo.yaml", "kubernetes/sample-cluster/another-tenant/sample-env/foo.yaml"}

	newDp := KubernetesAppDeployment{
		Deployment: Deployment{
			DeploymentPath: "kubernetes/sample-cluster/another-tenant/sample-env",
		},
		Cluster:     "sample-cluster",
		Tenant:      "another-tenant",
		Environment: "sample-env",
	}

	pathsJSON, err = json.Marshal(paths)
	if err != nil {
		t.Fatalf("Error marshalling paths to JSON: %v", err)
	}

	deployments = m.processUpdatedDeployments(string(pathsJSON))

	if len(deployments.KubernetesDeployments) != 2 {
		t.Errorf("Expected 2 deployments, got %v", len(deployments.KubernetesDeployments))
	}

	if expectedDp.Equals(deployments.KubernetesDeployments[0]) == false {
		t.Errorf("Expected %v, got %v", expectedDp, deployments.KubernetesDeployments[0])
	}

	if newDp.Equals(deployments.KubernetesDeployments[1]) == false {
		t.Errorf("Expected %v, got %v", newDp, deployments.KubernetesDeployments[1])
	}

}
