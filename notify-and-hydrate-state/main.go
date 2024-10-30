package main

import (
	"context"
	"dagger/notify-and-hydrate-state/internal/dagger"
	"fmt"
	"strings"
)

type NotifyAndHydrateState struct {
	FirestarterImage            string
	FirestarterImageTag         string
	GithubAppID                 string
	GithubInstallationID        string
	GithubPrefappInstallationID string
	GithubPrivateKey            *dagger.Secret
	GithubOrganization          string
	GhToken                     *dagger.Secret
	ClaimsDefaultBranch         string // +default="main"
}

func New(
	// +optional
	// +default="latest-slim"
	firestarterImageTag string,
	// +optional
	// +default="ghcr.io/prefapp/gitops-k8s"
	firestarterImage string,
	// +required
	// Github application ID
	githubAppID string,
	// +required
	// Github installation ID
	githubInstallationID string,
	// +required
	// Github prefapp installation ID
	githubPrefappInstallationID string,
	// +required
	// Github private key
	githubPrivateKey *dagger.Secret,
	// +required
	// Github organization
	githubOrganization string,
	// +required
	// Github token
	ghToken *dagger.Secret,
	// +default="main"
	claimsDefaultBranch string,

) *NotifyAndHydrateState {

	return &NotifyAndHydrateState{

		FirestarterImage: firestarterImage,

		FirestarterImageTag: firestarterImageTag,

		GithubAppID: githubAppID,

		GithubInstallationID: githubInstallationID,

		GithubPrefappInstallationID: githubPrefappInstallationID,

		GithubPrivateKey: githubPrivateKey,

		GithubOrganization: githubOrganization,

		GhToken: ghToken,

		ClaimsDefaultBranch: claimsDefaultBranch,
	}

}

func (m *NotifyAndHydrateState) Workflow(
	ctx context.Context,
	// Claims repository name
	// +required
	claimsRepo string,
	// Wet repository name
	// +required
	wetRepo string,
	// Claims directory
	// +required
	claimsDir *dagger.Directory,
	// Previous CRs directory
	// +required
	crsDir *dagger.Directory,
	// Provider to render
	// +required
	provider string,
	// Claims PR
	// +required
	claimsPr string,

) DiffResult {

	claimsPrNumber := strings.Split(claimsPr, "#")[1]

	newCrsDir := m.CmdHydrate(claimsRepo, claimsDir, crsDir, provider)

	affectedClaims, err := m.GetAffectedClaims(ctx, claimsRepo, claimsPrNumber, claimsDir)

	if err != nil {

		panic(fmt.Errorf("failed to get affected claims: %w", err))

	}

	diff := m.CompareDirs(ctx, crsDir, newCrsDir, affectedClaims)

	fsLog(fmt.Sprintf("Compared dirs has the diff: %+v", diff))

	previousPrs, err := m.GetRepoPrs(ctx, wetRepo)

	logPrs("Previous PRs", previousPrs)

	if err != nil {

		panic(fmt.Errorf("failed to get PRs: %w", err))

	}

	isValid, err := m.Verify(ctx, claimsPr, wetRepo, append(
		append(diff.DeletedFiles, diff.AddedFiles...),
		diff.ModifiedFiles...), previousPrs)

	if !isValid {

		fmt.Printf("<---- DEBUG ---->")
		fmt.Printf("Claims pr number: %s", claimsPrNumber)
		fmt.Printf("Claims repo: %s", claimsRepo)
		fmt.Printf("Error: %s", err)
		fmt.Printf("<---- DEBUG ---->")

		res, err := dag.Gh().Run(
			ctx,
			m.GhToken,
			strings.Join([]string{
				"gh",
				"pr",
				"comment",
				claimsPrNumber,
				"--body",
				fmt.Sprintf("Failed to verify hydrate process: %s", err),
				"-R", claimsRepo,
			}, " "),

			dagger.GhRunOpts{DisableCache: true},
		)

		if err != nil {

			panic(fmt.Errorf("failed to run gh command: %w", err))

		}

		fmt.Printf("Comment response: %s", res)

		panic(fmt.Errorf("failed to verify: %w", err))

	}

	childPreviousPrs, err := m.FilterByParentPr(
		ctx,
		claimsPrNumber,
		previousPrs,
	)

	logPrs("Child previous PRs", childPreviousPrs)

	if err != nil {

		panic(fmt.Errorf("failed to filter by parent PR: %w", err))

	}

	result, err := m.UpsertPrsFromDiff(
		ctx,
		&diff,
		crsDir,
		wetRepo,
		claimsPrNumber,
		childPreviousPrs,
	)

	if err != nil {

		panic(fmt.Errorf("failed to upsert PRs from diff: %w", err))

	}

	logPrs("orphan PRs", result.Orphans)

	m.CloseOrphanPrs(
		ctx,
		claimsPrNumber,
		result.Orphans,
		wetRepo,
	)

	logPrs("Created or updated PRs", result.Prs)

	_, err = m.AddPrReferences(ctx, claimsRepo, claimsPrNumber, result.Prs)

	if err != nil {

		panic(fmt.Errorf("failed to add PR references: %w", err))

	}

	return diff
}
