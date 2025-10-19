package main

import (
	"context"
	"testing"
)

func TestHydrateConfigExists(t *testing.T) {

	t.Run("TestHydrateConfigShouldNotExists", func(t *testing.T) {

		dir := dag.Directory()

		ctx := context.Background()

		exists, err := hydrateConfigFileExists(ctx, dir)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if exists {
			t.Errorf("Expected hydrate config file to not exist, but it does")
		}

	})

	t.Run("TestHydrateConfigShouldExists", func(t *testing.T) {
		dir := dag.Directory().
			WithNewFile(".github/hydrate_tfworkspaces_config.yaml", "test-content")

		ctx := context.Background()

		exists, err := hydrateConfigFileExists(ctx, dir)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if !exists {
			t.Errorf("Expected hydrate config file to exist, but it does not")
		}
	})

}
