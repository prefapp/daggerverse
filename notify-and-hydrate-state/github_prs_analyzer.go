package main

import (
	"context"
	"strings"
)

func (m *NotifyAndHydrateState) FilterByParentPr(

	ctx context.Context,

	parentPrNumber string,

	prs []Pr,

) ([]Pr, error) {

	filteredPrs := []Pr{}

	for _, pr := range prs {

		if isAutomatedPr(pr) && isChildPr(parentPrNumber, pr) {

			filteredPrs = append(filteredPrs, pr)

		}

	}

	return filteredPrs, nil

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
