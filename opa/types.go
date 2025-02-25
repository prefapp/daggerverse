package main

import "dagger/opa/internal/dagger"

type ClaimsDataRules struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	RegoFile    string `yaml:"regoFile"`
	File        *dagger.File
	ApplyTo     []ApplicableRule `yaml:"applyTo"`
}

type ApplicableRule struct {
	App          string `yaml:"app"`
	Name         string `yaml:"name"`
	Kind         string `yaml:"kind"`
	ResourceType string `yaml:"resourceType"`
	Environment  string `yaml:"env"`
	Tenant       string `yaml:"tenant"`
	Platform     string `yaml:"platform"`
}

type Claim struct {
	Name         string `yaml:"name"`
	Kind         string `yaml:"kind"`
	ResourceType string `yaml:"resourceType"`
}

type ClaimClassification struct {
	Name         string
	Kind         string
	Environment  string
	ResourceType string
	Tenant       string
	Platform     string
	App          string
	File         *dagger.File
}
