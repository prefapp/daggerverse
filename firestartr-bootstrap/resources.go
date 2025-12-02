package main

import (
	"context"
	"dagger/firestartr-bootstrap/internal/dagger"
	"fmt"
)

func (m *FirestartrBootstrap) PushCrsFiles(
	ctx context.Context,
	kindContainer *dagger.Container,

) error {

	tokenSecret, err := m.GenerateGithubToken(ctx)
	if err != nil {
		panic(err)
	}

	if m.Bootstrap.PushFiles.Claims.Push {
		claimsDir := kindContainer.
			Directory("/resources/claims").
			WithoutFile(fmt.Sprintf("claims/groups/%s-all.yaml", m.GhOrg))

		err := m.PushDirToRepo(
			ctx,
			claimsDir,
			m.Bootstrap.PushFiles.Claims.Repo,
			tokenSecret,
		)
		if err != nil {
			panic(err)
		}

		dotConfig := dag.Directory().
			WithDirectory(".config", m.ClaimsDotConfigDir)

		err = m.PushDirToRepo(
			ctx,
			dotConfig,
			m.Bootstrap.PushFiles.Claims.Repo,
			tokenSecret,
		)
		if err != nil {
			panic(err)
		}
	}

	if m.Bootstrap.PushFiles.Crs.Providers.Github.Push {
		crsDir := kindContainer.Directory("/resources/firestartr-crs/github")

		err := m.PushDirToRepo(
			ctx,
			crsDir,
			m.Bootstrap.PushFiles.Crs.Providers.Github.Repo,
			tokenSecret,
		)
		if err != nil {
			panic(err)
		}

		dotConfig := dag.Directory().
			WithDirectory(".config", m.CrsDotConfigDir)

		err = m.PushDirToRepo(
			ctx,
			dotConfig,
			m.Bootstrap.PushFiles.Crs.Providers.Github.Repo,
			tokenSecret,
		)
		if err != nil {
			panic(err)
		}
	}

	if m.Bootstrap.PushFiles.Crs.Providers.Terraform.Push {
		crsDir := kindContainer.Directory("/resources/firestartr-crs/infra")

		err := m.PushDirToRepo(
			ctx,
			crsDir,
			m.Bootstrap.PushFiles.Crs.Providers.Terraform.Repo,
			tokenSecret,
		)
		if err != nil {
			panic(err)
		}
	}

	return nil

}
