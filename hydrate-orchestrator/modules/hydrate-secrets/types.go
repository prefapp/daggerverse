package main

type Artifact struct {
	Metadata Metadata `yaml:"metadata"`
}

type KubernetesResource struct {
	Kind     string `yaml:"kind"`
	Metadata struct {
		Name string `yaml:"name"`
	} `yaml:"metadata"`
}

type Config struct {
	Image    string     `yaml:"image"`
	Commands [][]string `yaml:"commands"`
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
	MicroServicePointer string `yaml:"firestartr.dev/microservice"`
	Image               string `yaml:"firestartr.dev/image"`
	ClaimRef            string `yaml:"firestartr.dev/claim-ref"`
}
