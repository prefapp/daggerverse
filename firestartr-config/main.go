package main

import (
	"context"
	"dagger/firestartr-config/internal/dagger"
)

type FirestartrConfig struct {
	Apps             []FirestartrApp
	Registries       []Registry
	Repositories     []Repository
	Platforms        []Platform
	DotFirestartrDir *dagger.Directory
}

func New(
	ctx context.Context,

	// The path to the .firestartr directory
	dotFirestartr *dagger.Directory,

	// Schema release/version to validate against. Accepts npm semver versions
	// (e.g. "v3.2.0" or "3.2.0") or dist-tags (e.g. "latest").
	// +optional
	// +default="v3.2.0"
	schemaVersion string,

	// Schema source selector when schemaDirectory is not provided: empty/"npm"
	// resolves from the npm registry (default package
	// @firestartr/firestartr-claims_schemas), an HTTP(S) URL is treated as an
	// npm registry/mirror base. For offline/pre-fetched schemas, pass
	// schemaDirectory instead.
	// +optional
	schemaBaseURL string,

	// Optional directory containing pre-fetched schemas at
	// <schemaVersion>/<contract>.schema.json, read via the Dagger API. When
	// provided, it takes precedence over schemaBaseURL (portable offline fallback).
	// +optional
	schemaDirectory *dagger.Directory,

) (*FirestartrConfig, error) {
	schemas, err := newSchemaProvider(schemaVersion, schemaBaseURL, schemaDirectory)
	if err != nil {
		return nil, err
	}

	registries, err := loadRegistries(ctx, dotFirestartr, schemas)
	if err != nil {
		return nil, err
	}

	repositories, err := loadRepositories(ctx, dotFirestartr, schemas)
	if err != nil {
		return nil, err
	}

	apps, appsErr := loadApps(ctx, dotFirestartr, schemas)
	if appsErr != nil {
		return nil, appsErr
	}

	platforms, platformsErr := loadPlatforms(ctx, dotFirestartr, schemas)
	if platformsErr != nil {
		return nil, platformsErr
	}

	return &FirestartrConfig{
		DotFirestartrDir: dotFirestartr,
		Registries:       registries,
		Repositories:     repositories,
		Apps:             apps,
		Platforms:        platforms,
	}, nil
}
