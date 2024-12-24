package main

import (
	"context"
	"dagger/end-to-end-state-app/internal/dagger"
	"fmt"
	"path"
	"path/filepath"

	"github.com/google/uuid"
)

type EndToEndStateApp struct{}

func (m *EndToEndStateApp) Run(
	ctx context.Context,
	// Microservice directory
	//+required
	dir *dagger.Directory,
	// Microservice repository name
	//+optional
	//+default="firestartr-test/e2e-sample-app-micro-a"
	repoName string,
	// GitHub token
	//+required
	ghToken *dagger.Secret,
) {

	e2eDirPath := "/tmp"
	e2eFilePath := "contents/e2e.txt"

	// Create a file with random content
	file := dag.Container().From("alpine:latest").
		WithExec([]string{
			"sh", "-c", fmt.Sprintf("mkdir -p %s && echo \"%s\" > %s", filepath.Dir(filepath.Join(e2eDirPath, e2eFilePath)), uuid.NewString(), filepath.Join(e2eDirPath, e2eFilePath)),
		}).
		File(path.Join("/tmp", e2eFilePath))

	dag.Gh().Container(dagger.GhContainerOpts{
		Token:   ghToken,
		Plugins: []string{"prefapp/gh-commit"},
	}).
		WithDirectory(
			e2eDirPath,
			dir,
			dagger.ContainerWithDirectoryOpts{
				Exclude: []string{".git"},
			}).
		WithFile(path.Join(e2eDirPath, e2eFilePath), file).
		WithWorkdir(e2eDirPath).
		WithExec([]string{
			"gh",
			"commit",
			"-R", repoName,
			"-m", "Update e2e file",
			"-b", "dev",
		}).Sync(ctx)

}
