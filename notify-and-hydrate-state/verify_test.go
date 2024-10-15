package main

import (
	"testing"
)

func TestCrHasPendingPr(t *testing.T) {

	m := &NotifyAndHydrateState{}

	cr := &Cr{
		Metadata: Metadata{
			Name: "notify-and-hydrate-state-uuid",
		},
	}

	prs := []Pr{
		{
			HeadRefName: "automated/notify-and-hydrate-state-uuid-1",
		},
	}

	hasPendingPr, err := m.CrHasPendingPr(prs, "1", cr)

	if err != nil {
		t.Errorf("Error: %v", err)
	}

	if hasPendingPr {

		t.Errorf("Expected false, got true")

	}

}
