package main

import (
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

func (d *Deployments) addDeployment(dep interface{}) {
	switch dep.(type) {
	case *KubernetesDeployment:

		kdep := dep.(*KubernetesDeployment)
		if !lo.ContainsBy(d.KubernetesDeployments, func(kd KubernetesDeployment) bool {
			return kd.Equals(*kdep)
		}) {
			d.KubernetesDeployments = append(d.KubernetesDeployments, *kdep)
		}
	default:
		panic(fmt.Sprintf("Unknown deployment type: %T", dep))
	}
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
			"Deployment in cluster: `%s`, tenant: `%s`, env: `%s`",
			kd.Cluster, kd.Tenant, kd.Environment,
		)
	} else {
		return "Deployment coordinates:" +
			fmt.Sprintf("\n\t* Cluster: `%s`", kd.Cluster) +
			fmt.Sprintf("\n\t* Tenant: `%s`", kd.Tenant) +
			fmt.Sprintf("\n\t* Environment: `%s`", kd.Environment)
	}
}

func (kd *KubernetesDeployment) Labels() []string {
	return []string{
		fmt.Sprintf("type/kubernetes"),
		fmt.Sprintf("cluster/%s", kd.Cluster),
		fmt.Sprintf("tenant/%s", kd.Tenant),
		fmt.Sprintf("env/%s", kd.Environment),
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
