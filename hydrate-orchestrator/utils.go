package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Deployments struct {
	KubernetesDeployments []KubernetesDeployment
}

type Deployment struct {
	DeploymentPath string
}

type KubernetesDeployment struct {
	Deployment
	Cluster     string
	Tenant      string
	Environment string
}

func (m *HydrateOrchestrator) ProcessUpdatedDeployments(
	ctx context.Context,
	// List of updated deployments
	// +required
	updatedDeployments string,
) *Deployments {
	// Load the updated deployments from JSON string using gojq
	var deployments []string
	err := json.Unmarshal([]byte(updatedDeployments), &deployments)

	if err != nil {
		panic(err)
	}

	result := &Deployments{
		KubernetesDeployments: []KubernetesDeployment{},
	}

	for _, deployment := range deployments {

		dirs := splitPath(deployment)

		if len(dirs) == 0 {
			panic(fmt.Sprintf("Invalid deployment path provided: %s", deployment))
		}

		deploymentType := dirs[0]

		switch deploymentType {
		case "kubernetes":
			// Process kubernetes deployment
			dep := kubernetesDepFromStr(deployment)
			result.KubernetesDeployments = append(result.KubernetesDeployments, *dep)
		}

	}

	return result

}

func kubernetesDepFromStr(deployment string) *KubernetesDeployment {

	dirs := splitPath(deployment)

	if len(dirs) < 4 {
		panic(fmt.Sprintf("Invalid kubernetes deployment path provided: %s", deployment))
	}

	return &KubernetesDeployment{
		Deployment: Deployment{
			DeploymentPath: deployment,
		},
		Cluster:     dirs[1],
		Tenant:      dirs[2],
		Environment: dirs[3],
	}

}

func splitPath(path string) []string {
	return strings.Split(path, string(os.PathSeparator))
}
