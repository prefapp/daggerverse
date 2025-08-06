package main

import (
	"fmt"
	"strings"
)

func (m *Gh) filterReviewers(reviewers []string) []string {
	filteredReviewers := []string{}

	for _, reviewer := range reviewers {
		// filter reviewers to avoid duplicates
		if !strings.HasSuffix(reviewer, "[bot]") {
			// [bot] reviewers are not supported by the GitHub API, so we skip them.
			// We also check if the reviewer is already in the list to avoid duplicates.
			if !contains(filteredReviewers, reviewer) {
				filteredReviewers = append(filteredReviewers, reviewer)
			}
		} else {
			fmt.Printf("☢️ Skipping bot reviewer: %s\n", reviewer)
		}
	}

	return filteredReviewers
}

func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
