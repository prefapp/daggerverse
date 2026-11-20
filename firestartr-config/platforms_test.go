package main

import (
	"context"
	"strings"
	"testing"
)

func TestPlatforms(t *testing.T) {
	ctx := context.Background()
	dotFirestartr := testFixturesDir()
	schemas := testSchemaProvider()

	t.Run("Can load platforms from yaml and yml", func(t *testing.T) {
		platforms, err := loadPlatforms(ctx, dotFirestartr, schemas)
		if err != nil {
			t.Fatalf("Error loading platforms: %s", err)
		}
		if len(platforms) != 2 {
			t.Fatalf("Expected 2 platforms, got %d", len(platforms))
		}

		byName := map[string]Platform{}
		for _, p := range platforms {
			byName[p.Name] = p
		}

		cluster, ok := byName["cluster-name"]
		if !ok {
			t.Fatal("Expected cluster-name platform")
		}
		if cluster.PlatformType != "kubernetes" {
			t.Errorf("Expected type kubernetes, got %s", cluster.PlatformType)
		}
		if len(cluster.Tenants) != 1 || cluster.Tenants[0] != "test-tenant" {
			t.Errorf("Unexpected tenants: %#v", cluster.Tenants)
		}
		if len(cluster.Envs) != 2 {
			t.Errorf("Expected 2 envs, got %d", len(cluster.Envs))
		}
		if len(cluster.AllowedClaims) != 1 {
			t.Errorf("Expected 1 allowed claim, got %d", len(cluster.AllowedClaims))
		}

		azure, ok := byName["azure-example-platform"]
		if !ok {
			t.Fatal("Expected azure-example-platform platform")
		}
		if azure.PlatformType != "tfworkspaces" {
			t.Errorf("Expected type tfworkspaces, got %s", azure.PlatformType)
		}
	})

	t.Run("Rejects platform missing required fields", func(t *testing.T) {
		dir := dag.Directory().WithNewFile("platforms/p.yaml", "name: x\ntenants: [t]\nenvs: [e]\n")
		_, err := loadPlatforms(ctx, dir, schemas)
		if err == nil {
			t.Fatal("Expected error for missing type")
		}
		if !strings.Contains(err.Error(), "platforms/p.yaml") {
			t.Errorf("Expected file path in error, got: %s", err)
		}
		if !strings.Contains(err.Error(), "type") {
			t.Errorf("Expected type in error, got: %s", err)
		}
	})

	t.Run("Rejects platform with empty tenants", func(t *testing.T) {
		dir := dag.Directory().WithNewFile("platforms/p.yaml", "type: k8s\nname: x\ntenants: []\nenvs: [e]\n")
		_, err := loadPlatforms(ctx, dir, schemas)
		if err == nil {
			t.Fatal("Expected error for empty tenants")
		}
		if !strings.Contains(err.Error(), "tenants") {
			t.Errorf("Expected tenants in error, got: %s", err)
		}
	})
}
