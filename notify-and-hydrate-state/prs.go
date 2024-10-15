package main

import (
	"context"
	"fmt"
	"strings"
)

func (m *NotifyAndHydrateState) CloseOrphanPrs(
	ctx context.Context,
	prNumber string,
	upsertedPrs []Pr,
	wetRepo string,
) {

	childPrs, err := m.FilterByParentPr(
		ctx,
		prNumber,
		upsertedPrs,
	)

	if err != nil {

		panic(fmt.Errorf("failed to filter by parent PR: %w", err))

	}

	for _, childPr := range childPrs {

		if isOrphanPr(childPr, upsertedPrs) {

			m.ClosePr(ctx, wetRepo, string(childPr.Number))

		}
	}
}

func isOrphanPr(pr Pr, consideredPrs []Pr) bool {

	for _, consideredPr := range consideredPrs {

		if consideredPr.Url == pr.Url {

			return false

		}
	}

	return true

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
