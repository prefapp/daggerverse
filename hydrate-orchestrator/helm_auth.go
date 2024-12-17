package main

import (
	"context"
	"dagger/hydrate-orchestrator/internal/dagger"
	"fmt"
	"slices"
)

type HelmAuth struct {
	Username  string
	Password  *dagger.Secret
	Registry  string
	NeedsAuth bool
}

func (m *HydrateOrchestrator) GetHelmAuth(ctx context.Context, authDir *dagger.Directory) HelmAuth {

	needsAuth, err := m.NeedsHelmAuth(ctx, authDir)

	if err != nil {
		panic(err)
	}

	if !needsAuth {
		return HelmAuth{
			NeedsAuth: false,
		}
	}

	username := getFileContent(ctx, authDir.File("helm-auth/username"))

	password := getFileContent(ctx, authDir.File("helm-auth/password"))

	registry := getFileContent(ctx, authDir.File("helm-auth/registry"))

	return HelmAuth{
		Username:  username,
		Password:  dag.SetSecret("pass", password),
		Registry:  registry,
		NeedsAuth: true,
	}

}

func getFileContent(ctx context.Context, file *dagger.File) string {

	content, err := file.Contents(ctx)

	if err != nil {
		panic(fmt.Sprintf("Failed to read file: %v", err))
	}

	return content
}

func (m *HydrateOrchestrator) NeedsHelmAuth(ctx context.Context, authDir *dagger.Directory) (bool, error) {

	entries, err := authDir.Glob(ctx, "**/helm-auth/*")

	if err != nil {
		panic(err)
	}

	for _, file := range []string{"username", "password", "registry"} {

		if slices.Contains(entries, "helm-auth/"+file) {
			continue
		}

		return false, nil
	}

	return true, nil

}
