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

func (m *NotifyAndHydrateState) ClosePr(ctx context.Context, prNumber int, ghRepo string) (string, error) {

	command := strings.Join([]string{
		"pr",
		"close",
		fmt.Sprintf("%d", prNumber),
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

func (m *NotifyAndHydrateState) GetRepoPrs(ctx context.Context, ghRepo string) ([]Pr, error) {

	command := strings.Join([]string{
		"pr",
		"list",
		"--json",
		"headRefName",
		"--json",
		"number,url",
		"-L",
		"1000",
		"-R",
		ghRepo},
		" ")

	content, err := dag.Gh().Run(ctx, m.GhToken, command, GhRunOpts{DisableCache: true})

	if err != nil {

		return nil, err
	}

	prs := []Pr{}

	json.Unmarshal([]byte(content), &prs)

	return prs, nil
}

func (m *NotifyAndHydrateState) CloseOrphanPrs(
	ctx context.Context,
	prNumber string,
	orphanPrs []Pr,
	wetRepo string,
) {

	fsLog(fmt.Sprintf("💡 Closing orphan PRs for PR %s\n", prNumber))

	fsLog(fmt.Sprintf("💡 Orphan PRs: %v\n", orphanPrs))

	fsLog(fmt.Sprintf("💡 Wet repo: %s\n", wetRepo))

	for _, orphanPr := range orphanPrs {

		m.ClosePr(ctx, orphanPr.Number, wetRepo)

	}
}
