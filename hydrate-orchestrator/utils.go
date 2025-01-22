package main

import (
	"context"
	"dagger/hydrate-orchestrator/internal/dagger"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/samber/lo"
)

/*
struct to hold the updated deployments
*/

type Deployments struct {
	KubernetesDeployments []KubernetesAppDeployment
	KubernetesSysDeployments []KubernetesSysDeployment
}

func (d *Deployments) addDeployment(dep interface{}) {
	switch dep := dep.(type) {
	case *KubernetesAppDeployment:

		kdep := dep
		if !lo.ContainsBy(d.KubernetesDeployments, func(kd KubernetesAppDeployment) bool {
			return kd.Equals(*kdep)
		}) {
			d.KubernetesDeployments = append(d.KubernetesDeployments, *kdep)
			
		}
	case *KubernetesSysDeployment:
		if !lo.ContainsBy(d.KubernetesSysDeployments, func(kd KubernetesSysDeployment) bool {
			return kd.Equals(*dep)
		}) {
			d.KubernetesSysDeployments = append(d.KubernetesSysDeployments, *dep)
		}

	default:
		panic(fmt.Sprintf("Unknown deployment type: %T", dep))
	}
}

type Deployment struct {
	DeploymentPath string
}

/*
- kubernetes app specific deployment struct
*/

type KubernetesAppDeployment struct {
	Deployment
	Cluster     string
	Tenant      string
	Environment string
}

// Check if two KubernetesAppDeployment are equal
func (kd *KubernetesAppDeployment) Equals(other KubernetesAppDeployment) bool {
	return kd.DeploymentPath == other.DeploymentPath &&
		kd.Cluster == other.Cluster &&
		kd.Tenant == other.Tenant &&
		kd.Environment == other.Environment
}

func (kd *KubernetesAppDeployment) String(summary bool) string {

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

func (kd *KubernetesAppDeployment) Labels() []string {
	return []string{
		"type/kubernetes",
		fmt.Sprintf("cluster/%s", kd.Cluster),
		fmt.Sprintf("tenant/%s", kd.Tenant),
		fmt.Sprintf("env/%s", kd.Environment),
	}
}

/*
- Kubernetes sys service specific deployment struct
*/

type KubernetesSysDeployment struct {
	Deployment
	Cluster     	string
	SysServiceName 	string
}

// Check if two KubernetesSysDeployments are equal
func (kd *KubernetesSysDeployment) Equals(other KubernetesSysDeployment) bool {
	return kd.DeploymentPath == other.DeploymentPath &&
		kd.Cluster == other.Cluster &&
		kd.SysServiceName == other.SysServiceName
}

func (kd *KubernetesSysDeployment) String(summary bool) string {
	
	if summary {

		return fmt.Sprintf(
			"Deployment in cluster: `%s`, sys service: `%s`",
			kd.Cluster, kd.SysServiceName,
		)
	} else {
		return "Deployment coordinates:" +
			fmt.Sprintf("\n\t* Cluster: `%s`", kd.Cluster) +
			fmt.Sprintf("\n\t* Sys Service: `%s`", kd.SysServiceName)
	}
}

func (kd *KubernetesSysDeployment) Labels() []string {
	return []string{
		"type/kubernetes",
		fmt.Sprintf("cluster/%s", kd.Cluster),
		fmt.Sprintf("sys-service/%s", kd.SysServiceName),
	}
}


func kubernetesDepFromStr(deployment string) *KubernetesAppDeployment {

	dirs := splitPath(deployment)

	// In this case the modified file is kubernetes/<cluster>/<tenant>/<env>.yaml
	if len(dirs) == 4 {

		envFile := filepath.Base(deployment)
		env := strings.TrimSuffix(envFile, filepath.Ext(envFile))

		return &KubernetesAppDeployment{
			Deployment: Deployment{
				DeploymentPath: strings.Join(append(dirs[0:3], env), string(os.PathSeparator)),
			},
			Cluster:     dirs[1],
			Tenant:      dirs[2],
			Environment: env,
		}

	} else if len(dirs) > 4 {
		return &KubernetesAppDeployment{
			Deployment: Deployment{
				DeploymentPath: strings.Join(dirs[0:4], string(os.PathSeparator)),
			},
			Cluster:     dirs[1],
			Tenant:      dirs[2],
			Environment: dirs[3],
		}
	}

	panic(fmt.Sprintf("Invalid deployment path provided: %s", deployment))

}

func kubernetesSysDepFromStr(deployment string) *KubernetesSysDeployment {

	dirs := splitPath(deployment)

	// In this case the modified file is kubernetes/<cluster>/<sys-service>/values.yaml

	if len(dirs) > 3 {
		
		return &KubernetesSysDeployment{
			Deployment: Deployment{
				DeploymentPath: strings.Join(dirs[0:3], string(os.PathSeparator)),
			},
			Cluster: dirs[1],
			SysServiceName: dirs[2],
		}
	}

	panic(fmt.Sprintf("Invalid deployment path provided: %s", deployment))
}

func splitPath(path string) []string {
	return strings.Split(path, string(os.PathSeparator))
}

type BranchInfo struct {
	Name string
	SHA  string
}

func (m *HydrateOrchestrator) getBranchInfo(
	ctx context.Context,
) *BranchInfo {

	gitDirPath := "/git_dir"
	ctr := dag.Gh().Container(dagger.GhContainerOpts{
		Token: m.GhToken,
	}).
		WithDirectory(gitDirPath, m.ValuesStateDir).
		WithWorkdir(gitDirPath).
		WithEnvVariable("CACHE_BUSTER", time.Now().String())

	branch, err := ctr.WithExec([]string{
		"git",
		"branch",
		"--show-current",
	}).Stdout(ctx)

	branch = strings.TrimSpace(branch)

	if err != nil {
		panic(err)
	}

	sha, err := ctr.WithExec([]string{
		"git",
		"rev-parse",
		branch,
	}).Stdout(ctx)

	if err != nil {
		panic(err)
	}

	sha = strings.TrimSpace(sha)

	return &BranchInfo{
		Name: strings.TrimSpace(branch),
		SHA:  sha,
	}
}
