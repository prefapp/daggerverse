package main

type HelmValues struct {
	Labels         Labels         `yaml:"labels"`
	Deploy         Deploy         `yaml:"deploy"`
	ServiceAccount ServiceAccount `yaml:"serviceaccount"`
	RoleRules      []RoleRule     `yaml:"roleRules"`
	Secret         Secret         `yaml:"secret"`
	Config         Config         `yaml:"config"`
}

type Labels struct {
	App     string `yaml:"app"`
	Concern string `yaml:"concern"`
}

type Deploy struct {
	Replicas      int         `yaml:"replicas"`
	Image         Image       `yaml:"image"`
	Command       []string    `yaml:"command"`
	ContainerPort int         `yaml:"containerPort"`
	Probes        Probes      `yaml:"probes"`
	Resources     interface{} `yaml:"resources"`
	VolumeMounts  interface{} `yaml:"volumeMounts"`
	Volumes       interface{} `yaml:"volumes"`
}

type Image struct {
	Name       string `yaml:"name"`
	Tag        string `yaml:"tag"`
	PullPolicy string `yaml:"pullPolicy"`
}

type Probes struct {
	Liveness  interface{} `yaml:"liveness"`
	Readiness interface{} `yaml:"readiness"`
	Startup   interface{} `yaml:"startup"`
}

type ServiceAccount struct {
	Annotations interface{} `yaml:"annotations"`
}

type RoleRule struct {
	APIGroups []string `yaml:"apiGroups"`
	Resources []string `yaml:"resources"`
	Verbs     []string `yaml:"verbs"`
}

type Secret struct {
	Type string            `yaml:"type"`
	Data map[string]string `yaml:"data"`
}

type Config struct {
	Data map[string]string `yaml:"data"`
}

type KubernetesSecret struct {
	ApiVersion string `yaml:"apiVersion"`

	Kind string `yaml:"kind"`

	Metadata Metadata `yaml:"metadata"`

	Type string `yaml:"type"`

	Data map[string]string `yaml:"data"`
}
