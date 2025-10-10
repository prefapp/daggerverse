package main

import "dagger/update-claims-features/internal/dagger"

type UpdateClaimsFeatures struct {
	Repo              string
	Org               string
	GhToken           *dagger.Secret
	PrefappGhToken    *dagger.Secret
	GhCliVersion      string
	ClaimsDirPath     string
	ClaimsDir         *dagger.Directory
	DefaultBranch     string
	ClaimsToUpdate    []string
	FeaturesToUpdate  []string
	VersionConstraint string
	Automerge         bool
	LocalGhCliPath    *dagger.File
}

type Pr struct {
	HeadRefName string `json:"headRefName"`
	Url         string `json:"url"`
	Number      int    `json:"number"`
	State       string `json:"state"`
}

type ReleasesList struct {
	TagName string `json:"tagName"`
}

type ReleaseBody struct {
	Body string `json:"body"`
}
