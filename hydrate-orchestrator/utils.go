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
	KubernetesDeployments    []KubernetesAppDeployment
	KubernetesSysDeployments []KubernetesSysDeployment
	TfWorkspaceDeployments   []TfWorkspaceDeployment
	SecretsDeployment        []SecretsDeployment
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
	case *TfWorkspaceDeployment:
		if !lo.ContainsBy(d.TfWorkspaceDeployments, func(tfd TfWorkspaceDeployment) bool {
			return tfd.Equals(*dep)
		}) {
			d.TfWorkspaceDeployments = append(d.TfWorkspaceDeployments, *dep)
		}
	case *SecretsDeployment:
		if !lo.ContainsBy(d.SecretsDeployment, func(sd SecretsDeployment) bool {
			return sd.Equals(*dep)
		}) {
			d.SecretsDeployment = append(d.SecretsDeployment, *dep)
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
	Cluster          string
	Tenant           string
	Environment      string
	ImagesMatrix     string
	ServiceNames     []string
	RepositoryCaller string
	Image            string
}

/*
- TfWorkspaceDeployment specific deployment struct
*/
type TfWorkspaceDeployment struct {
	Deployment
	ClaimName    string
	Tenant       string
	Environment  string
	ImagesMatrix string
}

type SecretsDeployment struct {
	Deployment
	Tenant      string
	Environment string
}

func (sd *SecretsDeployment) Equals(other SecretsDeployment) bool {
	return sd.DeploymentPath == other.DeploymentPath &&
		sd.Tenant == other.Tenant &&
		sd.Environment == other.Environment
}

func (tfd *TfWorkspaceDeployment) Equals(other TfWorkspaceDeployment) bool {
	return tfd.DeploymentPath == other.DeploymentPath &&
		tfd.ClaimName == other.ClaimName
}

func (tfd *TfWorkspaceDeployment) String(summary bool) string {
	if summary {
		return fmt.Sprintf(
			"TFWorkspace deployment: `%s`",
			tfd.ClaimName,
		)
	} else {
		return "Deployment coordinates:" +
			fmt.Sprintf("\n\t* Claim: `%s`", tfd.ClaimName)
	}
}

func (sd *SecretsDeployment) String(summary bool) string {
	if summary {
		return fmt.Sprintf(
			"Secrets deployment tenant: `%s`, env: `%s`",
			sd.Tenant, sd.Environment,
		)
	} else {
		return "Deployment coordinates:" +
			fmt.Sprintf("\n\t* Tenant: `%s`", sd.Tenant) +
			fmt.Sprintf("\n\t* Environment: `%s`", sd.Environment)
	}
}

func (tfd *TfWorkspaceDeployment) Labels() []string {
	return []string{
		"plan",
	}
}

func (sd *SecretsDeployment) Labels() []string {
	return []string{
		"type/secrets",
		fmt.Sprintf("tenant/%s", sd.Tenant),
		fmt.Sprintf("env/%s", sd.Environment),
	}
}

// Check if two KubernetesAppDeployment are equal
func (kd *KubernetesAppDeployment) Equals(other KubernetesAppDeployment) bool {
	return kd.DeploymentPath == other.DeploymentPath &&
		kd.Cluster == other.Cluster &&
		kd.Tenant == other.Tenant &&
		kd.Environment == other.Environment
}

func (kd *KubernetesAppDeployment) String(summary bool) string {
	serviceNames := kd.ServiceNames
	if len(serviceNames) == 0 {
		serviceNames = []string{"unknown-service"}
	}
	repo := kd.RepositoryCaller
	if repo == "" {
		repo = "unknown-repo"
	}
	image := kd.Image
	if image == "" {
		image = "unknown-image"
	}

	if summary {
		serviceName := serviceNames[0]
		if len(serviceNames) > 1 {
			serviceName = strings.Join(serviceNames, ",")
		}

		return fmt.Sprintf(
			"Deployment of `%s` in cluster: `%s`, tenant: `%s`, env: `%s`",
			serviceName, kd.Cluster, kd.Tenant, kd.Environment,
		)
	} else {
		servicesList := ""
		for _, svc := range serviceNames {
			servicesList += fmt.Sprintf("  - %s\n", svc)
		}
		return fmt.Sprintf("\n\t* Repository: `%s`", repo) +
			fmt.Sprintf("\n\t* Services updated: `%s`", servicesList) +
			fmt.Sprintf("\n\t* New image: `%s`", image) +
			"\n\tDeployment coordinates:" +
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
	Cluster        string
	SysServiceName string
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

	fmt.Printf("kubernetesDepFromStr dirs: %v\n", dirs)
	fmt.Printf("kubernetesDepFromStr len(dirs): %d\n", len(dirs))
	// In this case the modified file is kubernetes/<cluster>/<tenant>/<env>.yaml
	if len(dirs) == 4 {
		fmt.Printf("kubernetesDepFromStr dirs are 4: %v\n", dirs)
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
		fmt.Printf("dirs are more than 4: %v\n", dirs)
		return &KubernetesAppDeployment{
			Deployment: Deployment{
				DeploymentPath: strings.Join(dirs[0:4], string(os.PathSeparator)),
			},
			Cluster:     dirs[1],
			Tenant:      dirs[2],
			Environment: dirs[3],
		}
	}

	panic(fmt.Sprintf("Invalid app deployment path provided: %s", deployment))

}

func kubernetesSysDepFromStr(deployment string) *KubernetesSysDeployment {

	dirs := splitPath(deployment)

	// In this case the modified file is kubernetes/<cluster>/<sys-service>/values.yaml
	fmt.Printf("kubernetesSysDepFromStr dirs: %v\n", dirs)
	fmt.Printf("kubernetesSysDepFromStr len(dirs): %d\n", len(dirs))
	if len(dirs) >= 3 {
		fmt.Printf("dirs are 3 or more: %v\n", dirs)
		sysServiceName := dirs[2]
		if len(dirs) == 3 &&
			(filepath.Ext(sysServiceName) == ".yaml" || filepath.Ext(sysServiceName) == ".yml") {
			sysServiceName = strings.TrimSuffix(sysServiceName, filepath.Ext(sysServiceName))
		}

		return &KubernetesSysDeployment{
			Deployment: Deployment{
				DeploymentPath: strings.Join(dirs[0:3], string(os.PathSeparator)),
			},
			Cluster:        dirs[1],
			SysServiceName: sysServiceName,
		}
	}

	panic(fmt.Sprintf("Invalid sys-service deployment path provided: %s", deployment))
}

func secretsDepFromStr(deployment string) *SecretsDeployment {

	dirs := splitPath(deployment)

	fmt.Printf("secretsDepFromStr dirs: %v\n", dirs)
	fmt.Printf("secretsDepFromStr len(dirs): %d\n", len(dirs))
	if len(dirs) == 3 {
		fmt.Printf("secretsDepFromStr dirs are 3: %v\n", dirs)
		return &SecretsDeployment{
			Deployment: Deployment{
				DeploymentPath: strings.Join(dirs, string(os.PathSeparator)),
			},
			Tenant:      dirs[1],
			Environment: dirs[2],
		}

	}

	panic(fmt.Sprintf("Invalid secrets deployment path provided: %s", deployment))
}

func splitPath(path string) []string {

	// remove "/" a the end
	if strings.HasSuffix(path, string(os.PathSeparator)) {
		path = strings.TrimSuffix(path, string(os.PathSeparator))
	}

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
		Token:   m.GhToken,
		Version: m.GhCliVersion,
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
