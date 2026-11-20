package main

import (
	"context"
	"strings"
	"testing"
)

func TestApps(t *testing.T) {
	ctx := context.Background()
	dotFirestartr := testFixturesDir()
	schemas := testSchemaProvider()

	t.Run("Can load app and deserialize valid service lists", func(t *testing.T) {
		apps, err := loadApps(ctx, dotFirestartr, schemas)
		if err != nil {
			t.Fatalf("Error loading apps: %s", err)
		}
		if len(apps) != 1 {
			t.Fatalf("Expected 1 app, got %d", len(apps))
		}
		app := apps[0]
		if app.Name != "sample-app" {
			t.Errorf("Expected name sample-app, got %s", app.Name)
		}
		if app.StateRepo != "org/state-app-sample" {
			t.Errorf("Expected state_repo org/state-app-sample, got %s", app.StateRepo)
		}
		if len(app.Services) != 1 {
			t.Fatalf("Expected 1 service, got %d", len(app.Services))
		}
		svc := app.Services[0]
		if svc.Repo != "org/build-and-dispatch" {
			t.Errorf("Expected repo org/build-and-dispatch, got %s", svc.Repo)
		}
		if len(svc.ServiceNames) != 2 || svc.ServiceNames[0] != "micro-a" || svc.ServiceNames[1] != "micro-b" {
			t.Errorf("Unexpected service names: %#v", svc.ServiceNames)
		}
	})

	t.Run("Rejects app missing services", func(t *testing.T) {
		dir := dag.Directory().WithNewFile("apps/app.yaml", "name: sample\nstate_repo: org/repo\n")
		_, err := loadApps(ctx, dir, schemas)
		if err == nil {
			t.Fatal("Expected error for missing services")
		}
		if !strings.Contains(err.Error(), "apps/app.yaml") {
			t.Errorf("Expected file path in error, got: %s", err)
		}
		if !strings.Contains(err.Error(), "services") {
			t.Errorf("Expected services in error, got: %s", err)
		}
	})

	t.Run("Rejects service missing required repo", func(t *testing.T) {
		dir := dag.Directory().WithNewFile("apps/app.yaml", "name: sample\nstate_repo: org/repo\nservices:\n  - service_names: [a]\n")
		_, err := loadApps(ctx, dir, schemas)
		if err == nil {
			t.Fatal("Expected error for missing repo")
		}
		if !strings.Contains(err.Error(), "repo") {
			t.Errorf("Expected repo in error, got: %s", err)
		}
	})

	t.Run("Rejects service with empty service_names", func(t *testing.T) {
		dir := dag.Directory().WithNewFile("apps/app.yaml", "name: sample\nstate_repo: org/repo\nservices:\n  - repo: org/x\n    service_names: []\n")
		_, err := loadApps(ctx, dir, schemas)
		if err == nil {
			t.Fatal("Expected error for empty service_names")
		}
		if !strings.Contains(err.Error(), "service_names") {
			t.Errorf("Expected service_names in error, got: %s", err)
		}
	})
}
