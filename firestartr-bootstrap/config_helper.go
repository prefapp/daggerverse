package main

import (
	"context"
	"dagger/firestartr-bootstrap/internal/dagger"
)

func getClaimsDotConfigDir(
	ctx context.Context,
	bootstrap interface{},
) (*dagger.Directory, error) {
	claimsDefaults, err := RenderDotConfigFile(
		ctx,
		dag.CurrentModule().
			Source().
			File("templates/claims_config/claims_defaults.tmpl"),
		bootstrap,
	)
	if err != nil {
		return nil, err
	}

	wetReposConfig, err := RenderDotConfigFile(
		ctx,
		dag.CurrentModule().
			Source().
			File("templates/claims_config/wet-repositories-config.tmpl"),
		bootstrap,
	)
	if err != nil {
		return nil, err
	}

	claimsDotConfigDir := dag.Directory().
		WithNewDirectory("/.config").
		WithNewFile("claims_defaults.yaml", claimsDefaults).
		WithNewFile("wet-repositories-config.yaml", wetReposConfig)

	return claimsDotConfigDir, nil
}

func getCrsDotConfigDir(
	ctx context.Context,
	bootstrap interface{},
	defaultsInterface CrsDefaultsData,
) (*dagger.Directory, error) {
	branchStrategies, err := RenderDotConfigFile(
		ctx,
		dag.CurrentModule().
			Source().
			File("templates/crs_config/branch_strategies.tmpl"),
		bootstrap,
	)
	if err != nil {
		return nil, err
	}

	expanderBranchStrategies, err := RenderDotConfigFile(
		ctx,
		dag.CurrentModule().
			Source().
			File("templates/crs_config/expander_branch_strategies.tmpl"),
		bootstrap,
	)
	if err != nil {
		return nil, err
	}

	groupDefaults, err := RenderDotConfigFile(
		ctx,
		dag.CurrentModule().
			Source().
			File("templates/crs_config/resources/defaults_github_group.tmpl"),
		defaultsInterface,
	)
	if err != nil {
		return nil, err
	}

	membersDefaults, err := RenderDotConfigFile(
		ctx,
		dag.CurrentModule().
			Source().
			File("templates/crs_config/resources/defaults_github_membership.tmpl"),
		defaultsInterface,
	)
	if err != nil {
		return nil, err
	}

	repoDefaults, err := RenderDotConfigFile(
		ctx,
		dag.CurrentModule().
			Source().
			File("templates/crs_config/resources/defaults_github_repository.tmpl"),
		defaultsInterface,
	)
	if err != nil {
		return nil, err
	}

	orgwebhookDefaults, err := RenderDotConfigFile(
		ctx,
		dag.CurrentModule().
			Source().
			File("templates/crs_config/resources/defaults_github_orgwebhook.tmpl"),
		defaultsInterface,
	)
	if err != nil {
		return nil, err
	}

	claimsDotConfigDir := dag.Directory().
		WithNewDirectory("/.config").
		WithNewFile("branch_strategies.yaml", branchStrategies).
		WithNewFile("expander_branch_strategies.yaml", expanderBranchStrategies).
		WithNewFile("resources/defaults_github_group.yaml", groupDefaults).
		WithNewFile("resources/defaults_github_membership.yaml", membersDefaults).
		WithNewFile("resources/defaults_github_repository.yaml", repoDefaults).
		WithNewFile("resources/defaults_github_orgwebhook.yaml", orgwebhookDefaults)

	return claimsDotConfigDir, nil
}
