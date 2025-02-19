package main

type ImageMatrix struct {
	Images []ImageData `json:"images"`
}

type ImageData struct {
	Tenant           string   `json:"tenant"`
	App              string   `json:"app"`
	Env              string   `json:"env"`
	ServiceNameList  []string `json:"service_name_list"`
	ImageKeys        []string `json:"image_keys"`
	Image            string   `json:"image"`
	Reviewers        []string `json:"reviewers"`
	Platform         string   `json:"platform"`
	Technology       string   `json:"technology"`
	RepositoryCaller string   `json:"repository_caller"`
}

type Claim struct {
	Name string `yaml:"name"`
}

type Cr struct {
	Kind       string   `yaml:"kind"`
	Metadata   Metadata `yaml:"metadata"`
	ApiVersion string   `yaml:"apiVersion"`
}

type Metadata struct {
	Annotations Annotations `yaml:"annotations"`
}

type Annotations struct {
	MicroService string `yaml:"firestartr.dev/microservice"`
	Image        string `yaml:"firestartr.dev/image"`
	ClaimRef     string `yaml:"firestartr.dev/claim-ref"`
}

type Config struct {
	Image string `yaml:"image"`
}
