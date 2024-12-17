package main

import (
	"encoding/json"
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
	BaseFolder       string   `json:"base_folder"`
	RepositoryCaller string   `json:"repository_caller"`
}

func (m *HydrateOrchestrator) RunDispatch(
	// +optional
	// +default="{\"images\":[]}"
	newImagesMatrix string,
) {

}

func (m *HydrateOrchestrator) processImagesMatrix(
	updatedDeployments string,
) *Deployments {
	result := &Deployments{
		KubernetesDeployments: []KubernetesDeployment{},
	}

	var imagesMatrix ImageMatrix
	err := json.Unmarshal([]byte(updatedDeployments), &imagesMatrix)

	if err != nil {
		panic(err)
	}

	for _, image := range imagesMatrix.Images {

		// At the moment the dispatch does not send the cluster so we extract it from the base folder
		cluster := strings.Split(image.BaseFolder, "/")[1]
		deploymentPath := filepath.Join(
			"kubernetes",
			cluster,
			image.Tenant,
			image.Env,
		)

		kdep := KubernetesDeployment{
			Deployment: Deployment{
				DeploymentPath: deploymentPath,
			},
			Cluster:     cluster,
			Tenant:      image.Tenant,
			Environment: image.Env,
		}

		result.addDeployment(kdep)

	}

	return result
}
