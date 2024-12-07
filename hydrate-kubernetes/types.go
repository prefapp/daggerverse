package main

type Artifact struct {
	Metadata Metadata `yaml:"metadata"`
}

type Metadata struct {
	Annotations Annotations `yaml:"annotations"`
}

type Annotations struct {
	MicroService string `yaml:"firestartr.dev/microservice"`

	Image string `yaml:"firestartr.dev/image"`
}

type ImageMatrix struct {
	Images []ImageData `json:"images"`
}

// JSON Types
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

type DepsFile struct {
	Dependencies []string `yaml:"deps"`
}

type KubernetesResource struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name string `yaml:"name"`
	} `yaml:"metadata"`
}
