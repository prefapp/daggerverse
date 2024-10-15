package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"
)

func (m *NotifyAndHydrateState) AddPrReferences(
	ctx context.Context,
	// Claims repository name
	// +required
	claimsRepo string,
	// Claim PR number
	// +required
	claimPrNumber string,
	// List of PRs
	// +required
	prs []Pr,

) (string, error) {

	prLinks := []string{}

	for _, pr := range prs {

		prLinks = append(prLinks, pr.Url)

	}

	const tpl = `
Related PRs:
{{range .}}
* {{.}}
{{end}}
`

	var buf bytes.Buffer

	t := template.Must(template.New("description").Parse(tpl))
	t.Execute(&buf, prLinks)

	return dag.Gh().Run(ctx, m.GhToken, strings.Join([]string{
		"pr",
		"comment",
		claimPrNumber,
		"--body",
		fmt.Sprintf("\"%s\"", buf.String()),
		"-R", claimsRepo,
	}, " "), GhRunOpts{DisableCache: true})

}

func (m *NotifyAndHydrateState) ClosePr(

	ctx context.Context,

	prNumber string,

	ghRepo string,

) (string, error) {

	command := strings.Join([]string{
		"pr",
		"close",
		prNumber,
		"-R",
		ghRepo,
	}, " ")

	return dag.
		Gh().
		Run(ctx,
			m.GhToken,
			command,
			GhRunOpts{DisableCache: true},
		)

}

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

func (m *NotifyAndHydrateState) GetRepoPrs(

	ctx context.Context,

	// Repository name ("<owner>/<repo>")
	ghRepo string,

) ([]Pr, error) {

	command := strings.Join([]string{"pr", "list", "--json", "headRefName", "--json", "url", "-L", "1000", "-R", ghRepo}, " ")

	content, err := dag.Gh().Run(ctx, m.GhToken, command, GhRunOpts{DisableCache: true})

	if err != nil {

		return nil, err
	}

	prs := []Pr{}

	json.Unmarshal([]byte(content), &prs)

	return prs, nil
}
