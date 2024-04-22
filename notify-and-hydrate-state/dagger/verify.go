func (m *NotifyAndHydrateState) VerifyPrs(

	ctx context.Context,

	// Github token secret
	ghToken Secret,

	// PR number ("<owner>/<repo>#<pr-number>")
	claimsPr string,

	// Repository name ("<owner>/<repo>")
	ghRepo string,

	// CRs to verify
	crs *[]File,

) bool {

	currentPrNumber := strings.Split(claimsPr, "#")[1]

	prs, err := m.GetRepoPrsByBranchName(ctx, &ghToken, ghRepo)

	if err != nil {

		panic(err)

	}

	for _, cr := range *crs {

		CrHasPendingPr, err := m.CrHasPendingPr(
			ctx,
			prs,
			currentPrNumber,
			ghRepo,
			&cr,
		)

		if err != nil {

			panic(err)

		}

		if CrHasPendingPr {

			return false

		}

	}

	return true
}

func (m *NotifyAndHydrateState) CrHasPendingPr(

	ctx context.Context,

	prs []PrBranchName,

	currentPrNumber string,

	ghRepo string,

	cr *File,

) (bool, error) {

	content, err := cr.Contents(ctx)

	if err != nil {

		return false, err

	}

	crObj := CR{}

	yaml.Unmarshal([]byte(content), &crObj)

	crHasPendingPr := false

	// Iterate the PRs to find the PR that matches the current PR number
	for _, pr := range prs {

		prExistsForCr := strings.Contains(pr.HeadRefName, crObj.Metadata.Name)

		if prExistsForCr {

			// Check if the PR number matches the current PR number
			// automated/<metadata-name>-<uuid>-<pr-number>
			uniqueValidPr := "automated/" + crObj.Metadata.Name + "-" + currentPrNumber

			if uniqueValidPr != pr.HeadRefName {
				crHasPendingPr = true
			}

		}

	}

	return crHasPendingPr, nil

}

type PrBranchName struct {
	HeadRefName string `json:"headRefName"`
}

func (m *NotifyAndHydrateState) GetRepoPrsByBranchName(

	ctx context.Context,

	ghToken *Secret,

	// Repository name ("<owner>/<repo>")
	ghRepo string,

) ([]PrBranchName, error) {

	command := strings.Join(

		[]string{
			"pr",
			"list",
			"--json",
			"headRefName",
			"-R",
			ghRepo,
		}, " ",
	)

	content, err := dag.
		Gh().
		Run(ctx, ghToken, command, GhRunOpts{DisableCache: true})

	if err != nil {

		return nil, err
	}

	prs := []PrBranchName{}

	json.Unmarshal(
		[]byte(content),
		&prs,
	)

	return prs, nil
}
