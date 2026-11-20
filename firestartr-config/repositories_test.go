package main

import (
	"context"
	"strings"
	"testing"
)

func TestRepositories(t *testing.T) {
	ctx := context.Background()
	dotFirestartr := testFixturesDir()
	schemas := testSchemaProvider()

	t.Run("Can get repositories from fixtures", func(t *testing.T) {
		repositories, err := loadRepositories(ctx, dotFirestartr, schemas)
		if err != nil {
			t.Fatalf("Error loading repositories: %s", err)
		}

		if len(repositories) != 3 {
			t.Fatalf("Expected 3 repositories, got %d", len(repositories))
		}

		expected := map[string]string{
			"argo":      "https://argoproj.github.io/argo-helm",
			"http-repo": "http://charts.example.com/stable",
			"oci-repo":  "oci://registry.example.com/charts",
		}

		for _, repo := range repositories {
			url, ok := expected[repo.Name]
			if !ok {
				t.Fatalf("Unexpected repository %s", repo.Name)
			}
			if repo.Url != url {
				t.Fatalf("Expected %s, got %s", url, repo.Url)
			}
			delete(expected, repo.Name)
		}
		if len(expected) != 0 {
			t.Fatalf("Missing repositories: %v", expected)
		}
	})

	t.Run("Accepts HTTP, HTTPS and OCI URLs", func(t *testing.T) {
		for _, tc := range []struct {
			name string
			url  string
		}{
			{name: "https", url: "https://charts.example.com/stable"},
			{name: "http", url: "http://charts.example.com/stable"},
			{name: "oci", url: "oci://registry.example.com/charts"},
			{name: "https-prefix-only", url: "https://"},
		} {
			dir := dag.Directory().WithNewFile("helm_repositories/repo.yaml", "name: repo\nurl: "+tc.url+"\n")
			repos, err := loadRepositories(ctx, dir, schemas)
			if err != nil {
				t.Fatalf("Expected url %s to be valid, got error: %s", tc.url, err)
			}
			if len(repos) != 1 || repos[0].Url != tc.url {
				t.Fatalf("Unexpected repositories for %s: %#v", tc.url, repos)
			}
		}
	})

	t.Run("Rejects invalid URL scheme", func(t *testing.T) {
		dir := dag.Directory().WithNewFile("helm_repositories/repo.yaml", "name: repo\nurl: ftp://charts.example.com\n")
		_, err := loadRepositories(ctx, dir, schemas)
		if err == nil {
			t.Fatal("Expected error for unsupported URL scheme")
		}
		if !strings.Contains(err.Error(), "helm_repositories/repo.yaml") {
			t.Errorf("Expected file path in error, got: %s", err)
		}
		if !strings.Contains(err.Error(), "url") {
			t.Errorf("Expected url in error, got: %s", err)
		}
	})

	t.Run("Rejects empty name", func(t *testing.T) {
		dir := dag.Directory().WithNewFile("helm_repositories/repo.yaml", "name: \"\"\nurl: https://charts.example.com\n")
		_, err := loadRepositories(ctx, dir, schemas)
		if err == nil {
			t.Fatal("Expected error for empty name")
		}
		if !strings.Contains(err.Error(), "name") {
			t.Errorf("Expected name in error, got: %s", err)
		}
	})
}
