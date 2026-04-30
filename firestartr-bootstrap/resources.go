package main

import (
	"context"
	"dagger/firestartr-bootstrap/internal/dagger"
	"fmt"
	"strings"
)

func (m *FirestartrBootstrap) PushBootstrapFiles(
	ctx context.Context,
	kindContainer *dagger.Container,
) error {

	tokenSecret, err := m.GenerateGithubToken(ctx)
	if err != nil {
		return err
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
			return err
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
			return err
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
			return err
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
			return err
		}
	}

	if m.Bootstrap.PushFiles.Crs.Providers.Terraform.Push {
		crsDir := kindContainer.Directory("/resources/firestartr-crs/infra")

		// Exclude non terraform CRs
		terraformDir := crsDir.Filter(dagger.DirectoryFilterOpts{
			Include: []string{"FirestartrTerraformWorkspace.*"},
		})

		err := m.PushDirToRepo(
			ctx,
			terraformDir,
			m.Bootstrap.PushFiles.Crs.Providers.Terraform.Repo,
			tokenSecret,
		)
		if err != nil {
			if strings.Contains(err.Error(), "nothing to commit") {
				fmt.Println("No terraform CRs to push, skipping...")
				return nil
			}
			return err
		}
	}

	if m.Bootstrap.PushFiles.Crs.Providers.Secrets.Push {
		crsDir := kindContainer.Directory("/resources/firestartr-crs/infra")

		// Exclude non secret CRs
		secretsDir := crsDir.Filter(dagger.DirectoryFilterOpts{
			Include: []string{"ExternalSecret.*"},
		})

		err := m.PushDirToRepo(
			ctx,
			secretsDir,
			m.Bootstrap.PushFiles.Crs.Providers.Secrets.Repo,
			tokenSecret,
		)
		if err != nil {
			return err
		}
	}

	if m.Bootstrap.PushFiles.DotFirestartr.Push {
		dotFirestartrDir := dag.CurrentModule().Source().Directory("./dot-firestartr")

		err := m.PushDirToRepo(
			ctx,
			dotFirestartrDir,
			m.Bootstrap.PushFiles.DotFirestartr.Repo,
			tokenSecret,
		)
		if err != nil {
			return err
		}
	}

	return nil

}
