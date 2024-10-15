package main

import (
	"context"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

func (m *NotifyAndHydrateState) Verify(

	ctx context.Context,

	// PR number ("<owner>/<repo>#<pr-number>")
	claimsPr string,

	// Repository name ("<owner>/<repo>")
	ghRepo string,

	// CRs to verify
	crs []*File,

	prs []Pr,

) (bool, error) {

	currentPrNumber := strings.Split(claimsPr, "#")[1]

	for _, cr := range crs {

		crInstance, err := m.unmarshalCr(ctx, cr)

		if err != nil {

			return false, fmt.Errorf("failed to get CR instance: %w", err)

		}

		crHasPendingPr, err := m.CrHasPendingPr(prs, currentPrNumber, &crInstance)

		if err != nil {

			return false, fmt.Errorf("failed to check if CR has pending PR: %w", err)

		}

		if crHasPendingPr {

			return false, fmt.Errorf("The CR %s has a pending PR", crInstance.Metadata.Name)

		}

	}

	return true, nil
}

func (*NotifyAndHydrateState) unmarshalCr(ctx context.Context, cr *File) (Cr, error) {

	crInstance := Cr{}

	contents, err := cr.Contents(ctx)

	if err != nil {

		return crInstance, fmt.Errorf("failed to get CR contents: %w", err)

	}

	yaml.Unmarshal([]byte(contents), &crInstance)

	return crInstance, nil
}

func (*NotifyAndHydrateState) CrHasPendingPr(

	prs []Pr,

	currentPrNumber string,

	cr *Cr,

) (bool, error) {

	for _, pr := range prs {

		if strings.Contains(pr.HeadRefName, cr.Metadata.Name) && !strings.Contains(pr.HeadRefName, "-plan") {

			// Pr format: automated/<metadata-name>-<uuid>-<pr-number>
			uniqueValidPr := "automated/" + cr.Metadata.Name + "-" + currentPrNumber

			if uniqueValidPr != pr.HeadRefName {

				return true, nil
			}

		}

	}

	return false, nil

}
