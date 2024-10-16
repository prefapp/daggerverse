package main

import (
	"context"
	"dagger/notify-and-hydrate-state/internal/dagger"
	"fmt"
	"strings"
	"time"
)

func (m *NotifyAndHydrateState) UpsertPrsFromDiff(

	ctx context.Context,

	diff *DiffResult,

	wetRepositoryDir *dagger.Directory,

	wetRepoName string,

	claimPrNumber string,

	prList []Pr,

) (PrsResult, error) {

	createdOrUpdatedPrs := []Pr{}

	orphanPrs := make([]Pr, len(prList))

	copy(orphanPrs, prList)

	for _, file := range diff.AddedFiles {

		pr, err := m.UpsertPr(ctx, file, wetRepositoryDir, wetRepoName, "create", claimPrNumber, prList)

		if err != nil {

			panic(err)

		}

		createdOrUpdatedPrs = append(createdOrUpdatedPrs, pr)

		orphanPrs = removeOrphan(orphanPrs, pr)

	}

	for _, file := range diff.ModifiedFiles {

		pr, err := m.UpsertPr(ctx, file, wetRepositoryDir, wetRepoName, "update", claimPrNumber, prList)

		if err != nil {

			panic(err)

		}

		createdOrUpdatedPrs = append(createdOrUpdatedPrs, pr)

		orphanPrs = removeOrphan(orphanPrs, pr)
	}

	for _, file := range diff.DeletedFiles {

		pr, err := m.UpsertPr(ctx, file, wetRepositoryDir, wetRepoName, "delete", claimPrNumber, prList)

		if err != nil {

			panic(err)

		}

		createdOrUpdatedPrs = append(createdOrUpdatedPrs, pr)

		orphanPrs = removeOrphan(orphanPrs, pr)

	}

	return PrsResult{Orphans: orphanPrs, Prs: createdOrUpdatedPrs}, nil

}

func removeOrphan(orphanPrs []Pr, pr Pr) []Pr {

	for i, orphanPr := range orphanPrs {

		if orphanPr.Url == pr.Url {

			orphanPrs = append(orphanPrs[:i], orphanPrs[i+1:]...)

		}

	}

	return orphanPrs
}

func (m *NotifyAndHydrateState) UpsertPr(

	ctx context.Context,

	file *dagger.File,

	wetRepositoryDir *dagger.Directory,

	wetRepoName string,

	action string,

	claimPrNumber string,

	prs []Pr,

) (Pr, error) {

	createdOrUpdatedPr := Pr{}

	fileName, err := file.Name(ctx)

	if err != nil {

		panic(err)

	}

	switch action {
	case "create":
		wetRepositoryDir = wetRepositoryDir.WithFile(fileName, file)
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

	gitContainer, err := m.ConfigGitContainer(ctx).
		WithEnvVariable("CACHEBUSTER", time.Now().String()).
		WithMountedDirectory("/repo", wetRepositoryDir).
		WithWorkdir("/repo").
		WithExec([]string{"git", "checkout", "-b", prBranch}).
		WithExec([]string{"git", "add", fileName}).
		WithExec([]string{"git", "commit", "-m", "Automated commit for CR " + cr.Metadata.Name}).
		WithExec([]string{"git", "push", "origin", prBranch, "--force"}).
		Sync(ctx)

	if err != nil {

		panic(err)

	}

	wetRepositoryDir = gitContainer.Directory("/repo")

	prTitle := fmt.Sprintf("\"hydrate: %s \"", cr.Metadata.Name)

	prBody := fmt.Sprintf("\"changes come from %s/%s#%s\"", strings.Split(wetRepoName, "/")[0], "claims", claimPrNumber)

	prLink, err := m.CreatePrIfNotExists(ctx, prBranch, wetRepoName, prTitle, prBody, prs)

	createdOrUpdatedPr.Url = prLink
	createdOrUpdatedPr.HeadRefName = prBranch

	if err != nil {

		panic(err)

	}

	wetRepositoryDir = m.CmdAnnotateCrPr(
		ctx,
		prLink,
		prLink,
		wetRepositoryDir,
		fileName,
	)

	gitContainer.
		WithMountedDirectory("/repo", wetRepositoryDir).
		WithWorkdir("/repo").
		WithExec([]string{"git", "checkout", prBranch}).
		WithExec([]string{"git", "add", fileName}).
		WithExec([]string{"git", "commit", "-m", "Automated commit for CR " + cr.Metadata.Name}).
		WithExec([]string{"git", "push", "origin", prBranch, "--force"}).
		Stdout(ctx)

	return createdOrUpdatedPr, nil
}

func (m *NotifyAndHydrateState) ConfigGitContainer(

	ctx context.Context,

) *dagger.Container {

	plainTextToken, err := m.GhToken.Plaintext(ctx)

	if err != nil {

		panic(err)

	}

	gitConfigContent := "https://firestartr:" + plainTextToken + "@github.com"

	return dag.Container().
		From("alpine/git").
		WithExec([]string{
			"git",
			"config",
			"--global",
			"url." + gitConfigContent + ".insteadOf",
			"https://github.com",
		}).
		WithExec([]string{
			"git",
			"config",
			"--global",
			"user.email",
			"firestartr-bot@firestartr.dev",
		}).
		WithExec([]string{
			"git",
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

	prs []Pr,

) (string, error) {

	for _, pr := range prs {

		if pr.HeadRefName == branch {

			return pr.Url, nil

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

	return dag.Gh().Run(ctx, m.GhToken, command, dagger.GhRunOpts{DisableCache: true})

}
