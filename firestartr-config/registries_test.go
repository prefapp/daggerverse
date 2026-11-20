package main

import (
	"context"
	"strings"
	"testing"
)

func TestRegistries(t *testing.T) {
	ctx := context.Background()
	dotFirestartr := testFixturesDir()
	schemas := testSchemaProvider()

	t.Run("Can get registries", func(t *testing.T) {
		registries, err := loadRegistries(ctx, dotFirestartr, schemas)
		if err != nil {
			t.Fatalf("Error loading registries: %s", err)
		}

		if len(registries) != 2 {
			t.Fatalf("Expected 2 registries, got %d", len(registries))
		}

		expected := map[string]struct {
			registry     string
			defaultValue bool
			auth         AuthStrategy
		}{
			"aws": {registry: "000000000000.dkr.ecr.eu-west-1.amazonaws.com", defaultValue: true, auth: AuthStrategyAWSOIDC},
			"az":  {registry: "acrtest.azurecr.io", defaultValue: false, auth: AuthStrategyAzureOIDC},
		}

		for _, reg := range registries {
			exp, ok := expected[reg.Name]
			if !ok {
				t.Fatalf("Unexpected registry %s", reg.Name)
			}
			if reg.Registry != exp.registry {
				t.Errorf("Expected registry %s for %s, got %s", exp.registry, reg.Name, reg.Registry)
			}
			if reg.Default != exp.defaultValue {
				t.Errorf("Expected default %v for %s, got %v", exp.defaultValue, reg.Name, reg.Default)
			}
			if reg.AuthStrategy != exp.auth {
				t.Errorf("Expected auth_strategy %s for %s, got %s", exp.auth, reg.Name, reg.AuthStrategy)
			}
			if len(reg.ImageTypes) == 0 {
				t.Errorf("Expected image_types for %s", reg.Name)
			}
			if reg.BasePaths.Services == "" {
				t.Errorf("Expected base_paths.services for %s", reg.Name)
			}
			delete(expected, reg.Name)
		}
		if len(expected) != 0 {
			t.Fatalf("Missing registries: %v", expected)
		}
	})

	t.Run("Accepts host without port and all auth strategies", func(t *testing.T) {
		strategies := []string{"aws_oidc", "azure_oidc", "generic", "ghcr", "dockerhub", "none"}
		for _, strategy := range strategies {
			dir := dag.Directory().WithNewFile("docker_registries/test.yaml", "name: test\nregistry: registry.example.com\nimage_types: [services]\ndefault: true\nauth_strategy: "+strategy+"\nbase_paths:\n  services: svc\n")
			registries, err := loadRegistries(ctx, dir, schemas)
			if err != nil {
				t.Fatalf("Expected auth strategy %s to be valid, got error: %s", strategy, err)
			}
			if len(registries) != 1 || string(registries[0].AuthStrategy) != strategy {
				t.Fatalf("Expected 1 registry with auth_strategy %s, got %#v", strategy, registries)
			}
		}
	})

	t.Run("Rejects registry with port", func(t *testing.T) {
		dir := dag.Directory().WithNewFile("docker_registries/test.yaml", "name: test\nregistry: registry.example.com:5000\nimage_types: [services]\ndefault: true\nauth_strategy: none\nbase_paths:\n  services: svc\n")
		_, err := loadRegistries(ctx, dir, schemas)
		if err == nil {
			t.Fatal("Expected error for registry with port")
		}
		if !strings.Contains(err.Error(), "docker_registries/test.yaml") {
			t.Errorf("Expected file path in error, got: %s", err)
		}
		if !strings.Contains(err.Error(), "docker_registries") {
			t.Errorf("Expected contract in error, got: %s", err)
		}
	})

	t.Run("Rejects unknown auth strategy", func(t *testing.T) {
		dir := dag.Directory().WithNewFile("docker_registries/test.yaml", "name: test\nregistry: registry.example.com\nimage_types: [services]\ndefault: true\nauth_strategy: typo\nbase_paths:\n  services: svc\n")
		_, err := loadRegistries(ctx, dir, schemas)
		if err == nil {
			t.Fatal("Expected error for unknown auth strategy")
		}
		if !strings.Contains(err.Error(), "docker_registries/test.yaml") {
			t.Errorf("Expected file path in error, got: %s", err)
		}
		if !strings.Contains(err.Error(), "auth_strategy") {
			t.Errorf("Expected auth_strategy in error, got: %s", err)
		}
	})

	t.Run("Rejects missing required fields instead of silently dropping", func(t *testing.T) {
		dir := dag.Directory().WithNewFile("docker_registries/test.yaml", "name: test\nregistry: registry.example.com\nimage_types: [services]\nauth_strategy: none\nbase_paths:\n  services: svc\n")
		_, err := loadRegistries(ctx, dir, schemas)
		if err == nil {
			t.Fatal("Expected error for missing default")
		}
		if !strings.Contains(err.Error(), "docker_registries/test.yaml") {
			t.Errorf("Expected file path in error, got: %s", err)
		}
		if !strings.Contains(err.Error(), "default") {
			t.Errorf("Expected missing default in error, got: %s", err)
		}
	})

	t.Run("Rejects missing base_paths.services", func(t *testing.T) {
		dir := dag.Directory().WithNewFile("docker_registries/test.yaml", "name: test\nregistry: registry.example.com\nimage_types: [services]\ndefault: true\nauth_strategy: none\nbase_paths:\n  charts: c\n")
		_, err := loadRegistries(ctx, dir, schemas)
		if err == nil {
			t.Fatal("Expected error for missing base_paths.services")
		}
		if !strings.Contains(err.Error(), "services") {
			t.Errorf("Expected services in error, got: %s", err)
		}
	})
}
