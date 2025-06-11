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
	ClaimsToUpdate       []string
	FeaturesToUpdate     []string
	VersionConstraint    string
	Automerge            bool
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

type Claim struct {
	Kind          string            `yaml:"kind,omitempty"`
	Version       string            `yaml:"version,omitempty"`
	Type          string            `yaml:"type,omitempty"`
	Lifecycle     string            `yaml:"lifecycle,omitempty"`
	System        string            `yaml:"system,omitempty"`
	Name          string            `yaml:"name,omitempty"`
	Description   string            `yaml:"description,omitempty"`
	Owner         string            `yaml:"owner,omitempty"`
	Annotations   map[string]string `yaml:"annotations,omitempty"`
	Providers     Providers         `yaml:"providers,omitempty"`
	PlatformOwner string            `yaml:"platformOwner,omitempty"`
	MaintainedBy  []string          `yaml:"maintainedBy,omitempty"`
	Profile       Profile           `yaml:"profile,omitempty"`
}

type Providers struct {
	Github Github `yaml:"github,omitempty"`
}

type Profile struct {
	DisplayName string `yaml:"displayName,omitempty"`
	Email       string `yaml:"email,omitempty"`
	Picture     string `yaml:"picture,omitempty"`
}

type Github struct {
	Description         string         `yaml:"description,omitempty"`
	Name                string         `yaml:"name,omitempty"`
	Org                 string         `yaml:"org,omitempty"`
	Visibility          string         `yaml:"visibility,omitempty"`
	DefaultBranch       string         `yaml:"defaultBranch,omitempty"`
	OrgPermissions      string         `yaml:"orgPermissions,omitempty"`
	Technology          Technology     `yaml:"technology,omitempty"`
	AdditionalBranches  []Branch       `yaml:"additionalBranches,omitempty"`
	BranchStrategy      BranchStrategy `yaml:"branchStrategy,omitempty"`
	AllowSquashMerge    bool           `yaml:"allowSquashMerge,omitempty"`
	AllowMergeCommit    bool           `yaml:"allowMergeCommit,omitempty"`
	AllowRebaseMerge    bool           `yaml:"allowRebaseMerge,omitempty"`
	DeleteBranchOnMerge bool           `yaml:"deleteBranchOnMerge,omitempty"`
	AutoInit            bool           `yaml:"autoInit,omitempty"`
	ArchiveOnDestroy    bool           `yaml:"archiveOnDestroy,omitempty"`
	AllowUpdateBranch   bool           `yaml:"allowUpdateBranch,omitempty"`
	HasIssues           bool           `yaml:"hasIssues,omitempty"`
	Pages               Pages          `yaml:"pages,omitempty"`
	Actions             Actions        `yaml:"actions,omitempty"`
	Overrides           Overrides      `yaml:"overrides,omitempty"`
	Features            []Feature      `yaml:"features,omitempty"`
}

type Technology struct {
	Stack   string `yaml:"stack,omitempty"`
	Version string `yaml:"version,omitempty"`
}

type BranchStrategy struct {
	Name          string `yaml:"name,omitempty"`
	DefaultBranch string `yaml:"defaultBranch,omitempty"`
}

type Pages struct {
	Cname  string `yaml:"cname,omitempty"`
	Source Source `yaml:"source,omitempty"`
}

type Source struct {
	Branch string `yaml:"branch,omitempty"`
	Path   string `yaml:"path,omitempty"`
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
