package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/samber/lo"
)

/*
struct to hold the updated deployments
*/

type Deployments struct {
	KubernetesDeployments []KubernetesDeployment
}

type Deployment struct {
	DeploymentPath string
}

/*
kubernetes specific deployment struct
*/

type KubernetesDeployment struct {
	Deployment
	Cluster     string
	Tenant      string
	Environment string
}

// Check if two KubernetesDeployments are equal
func (kd *KubernetesDeployment) Equals(other KubernetesDeployment) bool {
	return kd.DeploymentPath == other.DeploymentPath &&
		kd.Cluster == other.Cluster &&
		kd.Tenant == other.Tenant &&
		kd.Environment == other.Environment
}

// Process updated deployments and return all unique deployments after validating and processing them
func (m *HydrateOrchestrator) ProcessUpdatedDeployments(
	ctx context.Context,
	// List of updated deployments in JSON format
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
			if !lo.ContainsBy(result.KubernetesDeployments, func(d KubernetesDeployment) bool {
				return d.Equals(*dep)
			}) {
				result.KubernetesDeployments = append(result.KubernetesDeployments, *dep)
			}
		}

	}

	return result

}

func kubernetesDepFromStr(deployment string) *KubernetesDeployment {

	dirs := splitPath(deployment)

	if len(dirs) < 4 {
		panic(fmt.Sprintf("Invalid kubernetes deployment path provided: %s", deployment))
	}

	// In this case the modified file is kubernetes/<cluster>/<tenant>/<env>.yaml
	if len(dirs) == 4 {

		envFile := filepath.Base(deployment)
		env := strings.TrimSuffix(envFile, filepath.Ext(envFile))

		return &KubernetesDeployment{
			Deployment: Deployment{
				DeploymentPath: strings.Join(append(dirs[0:3], env), string(os.PathSeparator)),
			},
			Cluster:     dirs[1],
			Tenant:      dirs[2],
			Environment: env,
		}

	} else {
		return &KubernetesDeployment{
			Deployment: Deployment{
				DeploymentPath: strings.Join(dirs[0:4], string(os.PathSeparator)),
			},
			Cluster:     dirs[1],
			Tenant:      dirs[2],
			Environment: dirs[3],
		}
	}

}

func splitPath(path string) []string {
	return strings.Split(path, string(os.PathSeparator))
}
