package main

import (
	"context"
	"strings"
)

func (m *NotifyAndHydrateState) CreatePr(

	ctx context.Context,

	file *File,

	wetRepositoryDir *Directory,

	wetRepoName string,

	action string,

	claimPrNumber string,

	token *Secret,

) *Directory {

	fileName, err := file.Name(ctx)

	if err != nil {

		panic(err)

	}

	switch action {
	case "create":
	case "update":
		wetRepositoryDir.WithFile(fileName, file)
	case "delete":
		wetRepositoryDir.WithoutFile(fileName)
	}

	cr, err := m.unmarshalCr(ctx, file)

	if err != nil {

		panic(err)

	}

	prBranch := "automated/" + cr.Metadata.Name + "-" + claimPrNumber

	// dag.Git().
	// 	Load(wetRepositoryDir).
	// 	WithCommand([]string{"checkout", "-b", prBranch}).
	// 	WithCommand([]string{"add", fileName}).
	// 	WithCommand([]string{"commit", "-m", "Update " + fileName}).
	// 	WithCommand([]string{"push"})

	command := strings.Join([]string{"pr", "create", "-H", prBranch, "-R", wetRepoName}, " ")

	dag.Gh().Run(ctx, token, command, GhRunOpts{DisableCache: true})

	return wetRepositoryDir

}
