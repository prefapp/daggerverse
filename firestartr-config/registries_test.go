package main

import (
	"context"
	"testing"
)

func TestRegistries(t *testing.T) {

	ctx := context.Background()

	dotFirestartr := getDir("./fixtures/.firestartr")

	t.Run("Can get registries", func(t *testing.T) {

		registries, err := loadRegistries(ctx, dotFirestartr.Directory("/fixtures/.firestartr"))

		if err != nil {

			t.Errorf("Error loading registries: %s", err)

		}

		if len(registries) != 1 {
			t.Errorf("Expected 1 registry, got %d", len(registries))
		}

		expectedFirstRegistry := "xxxxxxxxxx.azurecr.io"

		if registries[0].Url != expectedFirstRegistry {
			t.Errorf("Expected %s, got %s", expectedFirstRegistry, registries[0].Url)
		}

	})
}
