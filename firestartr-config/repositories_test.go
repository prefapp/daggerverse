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

			t.Errorf("Error loading repositories: %s", err)

		}

		if len(repositories) != 2 {
			t.Errorf("Expected 2 repositories, got %d", len(repositories))
		}

		expectedFirstRegistry := "000000000000.dkr.ecr.eu-west-1.amazonaws.com"

		if repositories[0].Url != expectedFirstRegistry {
			t.Errorf("Expected %s, got %s", expectedFirstRegistry, repositories[0].Url)
		}

		expectedSecondRegistry := "xxxxxxxxxx.azurecr.io"

		if repositories[1].Url != expectedSecondRegistry {
			t.Errorf("Expected %s, got %s", expectedSecondRegistry, repositories[1].Url)
		}
	})
}
