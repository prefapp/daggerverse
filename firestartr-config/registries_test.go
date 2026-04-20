package main

import (
	"context"
	"testing"
)

func TestLegacyRegistries(t *testing.T) {

	ctx := context.Background()

	dotFirestartr := getDir("./fixtures/.firestartr")

	t.Run("Can get registries", func(t *testing.T) {

		registries, err := loadRegistries(ctx, dotFirestartr.Directory("/fixtures/.firestartr"))

		if err != nil {

			t.Fatalf("Error loading registries: %s", err)

		}

		if len(registries) != 1 {
			t.Fatalf("Expected 1 registry, got %d", len(registries))
		}

		expectedFirstRegistry := "000000000000.dkr.ecr.eu-west-1.amazonaws.com"

		if registries[0].Registry != expectedFirstRegistry {
			t.Errorf("Expected %s, got %s", expectedFirstRegistry, registries[0].Registry)
		}

	})
}
