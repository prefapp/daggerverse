package main

import "dagger/notify-and-hydrate-state/internal/dagger"

type DiffResult struct {
	AddedFiles []*dagger.File

	DeletedFiles []*dagger.File

	ModifiedFiles []*dagger.File

	UnmodifiedFiles []*dagger.File
}

type PrsResult struct {
	Orphans []Pr

	Prs []Pr
}

type PrFiles struct {
	AddedModified []string
	Deleted       []string
}

type Metadata struct {
	Name string `yaml:"name"`
}

type Cr struct {
	Metadata Metadata
}

type Pr struct {
	HeadRefName string `json:"headRefName"`
	Url         string `json:"url"`
	Number      int    `json:"number"`
}
