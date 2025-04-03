package main

import "dagger/update-claims-features/internal/dagger"

type UpdateClaimsFeatures struct {
	Repo                 string
	GhToken              *dagger.Secret
	PrefappGhToken       *dagger.Secret
	GhCliVersion         string
	ClaimsDirPath        string
	ClaimsDir            *dagger.Directory
	DefaultBranch        string
	ComponentsFolderName string
	ClaimToUpdate        string
	FeatureToUpdate      string
	VersionConstraint    string
	Automerge            bool
}

type ReleasesList struct {
	TagName string `json:"tagName"`
}

type Claim struct {
	Kind      string    `yaml:"kind,omitempty"`
	Version   string    `yaml:"version,omitempty"`
	Type      string    `yaml:"type,omitempty"`
	Lifecycle string    `yaml:"lifecycle,omitempty"`
	System    string    `yaml:"system,omitempty"`
	Name      string    `yaml:"name,omitempty"`
	Providers Providers `yaml:"providers,omitempty"`
	Owner     string    `yaml:"owner,omitempty"`
}

type Providers struct {
	Github Github `yaml:"github,omitempty"`
}

type Github struct {
	Description        string         `yaml:"description,omitempty"`
	Name               string         `yaml:"name,omitempty"`
	Org                string         `yaml:"org,omitempty"`
	Visibility         string         `yaml:"visibility,omitempty"`
	AdditionalBranches []Branch       `yaml:"additionalBranches,omitempty"`
	BranchStrategy     BranchStrategy `yaml:"branchStrategy,omitempty"`
	Actions            Actions        `yaml:"actions,omitempty"`
	Overrides          Overrides      `yaml:"overrides,omitempty"`
	Features           []Feature      `yaml:"features,omitempty"`
}

type BranchStrategy struct {
	Name          string `yaml:"name,omitempty"`
	DefaultBranch string `yaml:"defaultBranch,omitempty"`
}

type Actions struct {
	Oidc OIDC `yaml:"oidc,omitempty"`
}

type OIDC struct {
	UseDefault       bool     `yaml:"useDefault"`
	IncludeClaimKeys []string `yaml:"includeClaimKeys,omitempty"`
}

type Feature struct {
	Name    string            `yaml:"name,omitempty"`
	Version string            `yaml:"version,omitempty"`
	Args    map[string]string `yaml:"args,omitempty"`
}

type Branch struct {
	Name   string `yaml:"name,omitempty"`
	Orphan bool   `yaml:"orphan"`
}

type Overrides struct {
	AdditionalAdmins      []string `yaml:"additionalAdmins,omitempty"`
	AdditionalMaintainers []string `yaml:"additionalMaintainers,omitempty"`
	AdditionalWriters     []string `yaml:"additionalWriters,omitempty"`
	AdditionalReaders     []string `yaml:"additionalReaders,omitempty"`
}
