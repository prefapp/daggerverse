package main

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
	return nil
}
