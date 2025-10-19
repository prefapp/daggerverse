package main

import (
	"context"
	"testing"
)

func TestFilterByParentPr(t *testing.T) {

	m := &NotifyAndHydrateState{}

	parentNumber := "1"

	prs := []Pr{
		{
			HeadRefName: "automated/b-1",
		},
		{
			HeadRefName: "automated/b-2",
		},
	}

	filteredPrs, err := m.FilterByParentPr(
		context.Background(),
		parentNumber,
		prs,
	)

	if err != nil {

		t.Errorf("Error: %v", err)

	}

	if len(filteredPrs) != 1 {

		t.Errorf("Expected 1, got %d", len(filteredPrs))

	}

	if filteredPrs[0].HeadRefName != "automated/b-1" {

		t.Errorf("Expected automated/b-1, got %s", filteredPrs[0].HeadRefName)

	}

}
