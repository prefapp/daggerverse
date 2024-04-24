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

) *Container {

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

	return m.ConfigGitContainer(ctx, token).
		WithMountedDirectory("/repo", wetRepositoryDir).
		WithWorkdir("/repo").
		WithExec([]string{"checkout", "-b", prBranch}).
		WithExec([]string{"add", fileName}).
		WithExec([]string{"commit", "-m", "Automated commit for CR " + cr.Metadata.Name}).
		WithExec([]string{"push", "origin", prBranch})

	command := strings.Join([]string{"pr", "create", "-H", prBranch, "-R", wetRepoName}, " ")

	dag.Gh().Run(ctx, token, command, GhRunOpts{DisableCache: true})

}

func (m *NotifyAndHydrateState) ConfigGitContainer(

	ctx context.Context,

	token *Secret,

) *Container {

	plainTextToken, err := token.Plaintext(ctx)

	if err != nil {

		panic(err)

	}

	gitConfigContent := "https://firestartr:" + plainTextToken + "@github.com"

	return dag.Container().
		From("alpine/git").
		WithExec([]string{
			"config",
			"--global",
			"url." + gitConfigContent + ".insteadOf",
			"https://github.com",
		}).
		WithExec([]string{
			"config",
			"--global",
			"user.email",
			"firestartr-bot@firestartr.dev",
		}).
		WithExec([]string{
			"config",
			"--global",
			"user.name",
			"firestartr-bot",
		})

}

func (m *NotifyAndHydrateState) Test(

	ctx context.Context,

) *Container {

	return dag.Container().
		From("alpine/git")
}
