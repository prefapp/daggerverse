package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
- kubernetes specific deployment struct
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

func (kd *KubernetesDeployment) String(summary bool) string {

	if summary {

		return fmt.Sprintf(
			"Deployment in cluster: %s, tenant: %s, env: %s",
			kd.Cluster, kd.Tenant, kd.Environment,
		)
	} else {
		return fmt.Sprintf(`
Deployment:
* Cluster: %s
* Tenant: %s
* Environment: %s
		`,
			kd.Cluster, kd.Tenant, kd.Environment,
		)
	}
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
