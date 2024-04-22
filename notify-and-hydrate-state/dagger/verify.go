package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type Metadata struct {
	Name string `yaml:"name"`
}

type Cr struct {
	Metadata Metadata
}

// This function verifies that the Open PRs for the CRs are not pending
// automated PRs.
func (m *NotifyAndHydrateState) Verify(

	ctx context.Context,

	// Github token secret
	ghToken *Secret,

	// PR number ("<owner>/<repo>#<pr-number>")
	claimsPr string,

	// Repository name ("<owner>/<repo>")
	ghRepo string,

	// CRs to verify
	crs *[]File,

) (bool, error) {

	currentPrNumber := strings.Split(claimsPr, "#")[1]

	prs, err := m.GetRepoPrs(ctx, ghToken, ghRepo)

	if err != nil {

		return false, fmt.Errorf("failed to get PRs: %w", err)

	}

	for _, cr := range *crs {

		crInstance, err := m.unmarshalCr(ctx, &cr)

		if err != nil {

			return false, fmt.Errorf("failed to get CR instance: %w", err)

		}

		crHasPendingPr, err := m.CrHasPendingPr(ctx, prs, currentPrNumber, ghRepo, &crInstance)

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

func (m *NotifyAndHydrateState) CrHasPendingPr(ctx context.Context,

	prs []PrBranchName,

	currentPrNumber string,

	ghRepo string,

	cr *Cr,

) (bool, error) {

	for _, pr := range prs {

		if strings.Contains(pr.HeadRefName, cr.Metadata.Name) {

			// Pr format: automated/<metadata-name>-<uuid>-<pr-number>
			uniqueValidPr := "automated/" + cr.Metadata.Name + "-" + currentPrNumber

			if uniqueValidPr != pr.HeadRefName {

				return true, nil
			}

		}

	}

	return false, nil

}

type PrBranchName struct {
	HeadRefName string `json:"headRefName"`
}

func (m *NotifyAndHydrateState) GetRepoPrs(

	ctx context.Context,

	ghToken *Secret,

	// Repository name ("<owner>/<repo>")
	ghRepo string,

) ([]PrBranchName, error) {

	command := strings.Join([]string{"pr", "list", "--json", "headRefName", "-R", ghRepo}, " ")

	content, err := dag.Gh().Run(ctx, ghToken, command, GhRunOpts{DisableCache: true})

	if err != nil {

		return nil, err
	}

	prs := []PrBranchName{}

	json.Unmarshal([]byte(content), &prs)

	return prs, nil
}
