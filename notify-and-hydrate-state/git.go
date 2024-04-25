package main

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func (m *NotifyAndHydrateState) CreatePrsFromDiff(

	ctx context.Context,

	diff *DiffResult,

	wetRepositoryDir *Directory,

	wetRepoName string,

	claimPrNumber string,

	prs []PrBranchName,

) []string {

	createdPrs := []string{}

	for _, file := range diff.AddedFiles {

		pr, err := m.CreatePr(ctx, file, wetRepositoryDir, wetRepoName, "create", claimPrNumber, prs)

		if err != nil {

			panic(err)

		}

		createdPrs = append(createdPrs, pr)

	}

	for _, file := range diff.ModifiedFiles {

		pr, err := m.CreatePr(ctx, file, wetRepositoryDir, wetRepoName, "update", claimPrNumber, prs)

		if err != nil {

			panic(err)

		}

		createdPrs = append(createdPrs, pr)

	}

	for _, file := range diff.DeletedFiles {

		pr, err := m.CreatePr(ctx, file, wetRepositoryDir, wetRepoName, "delete", claimPrNumber, prs)

		if err != nil {

			panic(err)

		}

		createdPrs = append(createdPrs, pr)

	}

	return createdPrs

}

func (m *NotifyAndHydrateState) CreatePr(

	ctx context.Context,

	file *File,

	wetRepositoryDir *Directory,

	wetRepoName string,

	action string,

	claimPrNumber string,

	prs []PrBranchName,

) (string, error) {

	fileName, err := file.Name(ctx)

	if err != nil {

		panic(err)

	}

	switch action {
	case "create":
	case "update":
		wetRepositoryDir = wetRepositoryDir.WithFile(fileName, file)
	case "delete":
		wetRepositoryDir = wetRepositoryDir.WithoutFile(fileName)
	}

	cr, err := m.unmarshalCr(ctx, file)

	if err != nil {

		panic(err)

	}

	prBranch := "automated/" + cr.Metadata.Name + "-" + claimPrNumber

	m.ConfigGitContainer(ctx).
		WithMountedDirectory("/repo", wetRepositoryDir).
		WithWorkdir("/repo").
		WithEnvVariable("CACHEBUSTER", time.Now().String()).
		WithExec([]string{"checkout", "-b", prBranch}).
		WithExec([]string{"add", fileName}).
		WithExec([]string{"commit", "-m", "Automated commit for CR " + cr.Metadata.Name}).
		WithExec([]string{"push", "origin", prBranch, "--force"}).
		Stdout(ctx)

	prTitle := fmt.Sprintf("\"hydrate: %s from  %s/%s#%s\"", cr.Metadata.Name, strings.Split(wetRepoName, "/")[0], "claims", claimPrNumber)

	prBody := fmt.Sprintf("\"hydrated: %s\"", cr.Metadata.Name)

	return m.CreatePrIfNotExists(ctx, prBranch, wetRepoName, prTitle, prBody, prs)

}

func (m *NotifyAndHydrateState) ConfigGitContainer(

	ctx context.Context,

) *Container {

	plainTextToken, err := m.GhToken.Plaintext(ctx)

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

func (m *NotifyAndHydrateState) CreatePrIfNotExists(

	ctx context.Context,

	branch string,

	repo string,

	title string,

	body string,

	prs []PrBranchName,

) (string, error) {

	for _, pr := range prs {

		if pr.HeadRefName == branch {

			return pr.HeadRefName, nil

		}

	}

	command := strings.Join([]string{
		"pr",
		"create",
		"-H",
		branch,
		"-R",
		repo,
		"-t",
		title,
		"-b",
		body,
	}, " ")

	return dag.Gh().Run(ctx, m.GhToken, command, GhRunOpts{DisableCache: true})

}
