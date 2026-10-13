package main

import (
	"dagger/update-claims-features/internal/dagger"
	"gopkg.in/yaml.v3"
)

type UpdateClaimsFeatures struct {
	Repo                      string
	Org                       string
	GhToken                   *dagger.Secret
	CustomFeaturesRepoGhToken *dagger.Secret
	PrefappGhToken            *dagger.Secret
	GhCliVersion              string
	ClaimsDirPath             string
	ClaimsDir                 *dagger.Directory
	DefaultBranch             string
	ClaimsToUpdate            []string
	FeaturesToUpdate          []string
	VersionConstraint         string
	Automerge                 bool
	LocalGhCliPath            *dagger.File
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

// loadedClaim holds a claim both as its original yaml.Node tree (used to
// re-serialize preserving field order) and as a map[string]any (used by the
// processing and validation logic).
type loadedClaim struct {
	node *yaml.Node
	data map[string]any
}
