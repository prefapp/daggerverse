package main

import (
	"context"
	"fmt"
	"strings"
)

func (m *NotifyAndHydrateState) CloseOrphanPrs(
	ctx context.Context,
	prNumber string,
	orphanPrs []Pr,
	wetRepo string,
) {
	fmt.Printf("💡 Closing orphan PRs for PR %s\n", prNumber)
	fmt.Printf("💡 Orphan PRs: %v\n", orphanPrs)
	fmt.Printf("💡 Wet repo: %s\n", wetRepo)

	for _, orphanPr := range orphanPrs {

		m.ClosePr(ctx, orphanPr.Number, wetRepo)

	}
}

func isChildPr(parentPrNumber string, pr Pr) bool {

	splitted := strings.Split(pr.HeadRefName, "-")

	isChild := (len(splitted) > 1 &&

		splitted[len(splitted)-1] == parentPrNumber)

	return isChild
}

func isAutomatedPr(pr Pr) bool {

	return strings.HasPrefix(pr.HeadRefName, "automated/")

}
