package main

import (
	"context"
	"testing"
)

func TestRepositories(t *testing.T) {

	ctx := context.Background()

	dotFirestartr := getDir("./fixtures/.firestartr")

	t.Run("Can get repositories", func(t *testing.T) {

		repositories, err := loadRepositories(ctx, dotFirestartr.Directory("/fixtures/.firestartr"))

		if err != nil {

			t.Fatalf("Error loading repositories: %s", err)

		}

		if len(repositories) != 1 {
			t.Fatalf("Expected 1 repository, got %d", len(repositories))
		}

		if repositories[0].Url != "https://argoproj.github.io/argo-helm" {
			t.Fatalf("Expected %s, got %s", "https://argoproj.github.io/argo-helm", repositories[0].Url)
		}

		if repositories[0].Name != "argo" {
			t.Fatalf("Expected %s, got %s", "argo", repositories[0].Name)
		}
	})
}
