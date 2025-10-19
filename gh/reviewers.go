package main

import (
	"fmt"
	"slices"
	"strings"
)

func (m *Gh) filterReviewers(reviewers []string) []string {
	filteredReviewers := []string{}

	for _, reviewer := range reviewers {
		// filter reviewers to avoid duplicates
		if !strings.HasSuffix(reviewer, "[bot]") {
			// [bot] reviewers are not supported by the GitHub API, so we skip them.
			// We also check if the reviewer is already in the list to avoid duplicates.
			if !slices.Contains(filteredReviewers, reviewer) {
				filteredReviewers = append(filteredReviewers, reviewer)
			}
		} else {
			fmt.Printf("☢️ Skipping bot reviewer: %s\n", reviewer)
		}
	}

	return filteredReviewers
}
