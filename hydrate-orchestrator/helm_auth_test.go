package main

import (
	"context"
	"testing"
)

func TestNeedsHelmAuth(t *testing.T) {

	ctx := context.Background()

	m := HydrateOrchestrator{}

	emptyDir := dag.Directory()

	needsAuth, err := m.NeedsHelmAuth(ctx, emptyDir)

	if err != nil {
		t.Errorf("Error: %v", err)
	}

	if needsAuth {
		t.Errorf("Expected no auth needed")
	}

}
