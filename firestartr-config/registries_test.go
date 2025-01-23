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

		if len(registries) != 2 {
			t.Errorf("Expected 2 registries, got %d", len(registries))
		}

		expectedFirstRegistry := "000000000000.dkr.ecr.eu-west-1.amazonaws.com"

		if registries[0].Registry != expectedFirstRegistry {
			t.Errorf("Expected %s, got %s", expectedFirstRegistry, registries[0].Registry)
		}

		expectedSecondRegistry := "xxxxxxxxxx.azurecr.io"

		if registries[1].Registry != expectedSecondRegistry {
			t.Errorf("Expected %s, got %s", expectedSecondRegistry, registries[1].Registry)
		}
	})
}
