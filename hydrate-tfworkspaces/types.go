package main

type ImageMatrix struct {
	Images []ImageData `json:"images"`
}

// JSON Types
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
