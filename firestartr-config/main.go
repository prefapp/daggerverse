// A generated module for FirestartrConfig functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
	"dagger/firestartr-config/internal/dagger"
)

type FirestartrConfig struct {
	Apps             []FirestartrApp
	Registries       []Registry
	Platforms        []Platform
	DotFirestartrDir *dagger.Directory
}

func New(

	ctx context.Context,

	// The path to the values directory, where the helm values are stored
	dotFirestartr *dagger.Directory,

) (*FirestartrConfig, error) {

	registries, err := loadRegistries(ctx, dotFirestartr)

	if err != nil {

		return nil, err

	}

	apps, appsErr := loadApps(ctx, dotFirestartr)

	if appsErr != nil {

		return nil, appsErr

	}

	platforms, platformsErr := loadPlatforms(ctx, dotFirestartr)

	if platformsErr != nil {

		return nil, platformsErr

	}

	return &FirestartrConfig{

		DotFirestartrDir: dotFirestartr,

		Registries: registries,

		Apps: apps,

		Platforms: platforms,
	}, nil
}
