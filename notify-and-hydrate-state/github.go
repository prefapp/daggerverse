package main

import (
	"bytes"
	"context"
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
	// List of PR links
	// +required
	prLinks []string,
	// Github Token
	// +required
	githubToken *Secret,

) (string, error) {

	const tpl = `
Related PRs:
{{range .}}
* {{.}}
{{end}}
`

	var buf bytes.Buffer

	t := template.Must(template.New("description").Parse(tpl))
	t.Execute(&buf, prLinks)

	return dag.Gh().Run(ctx, githubToken, strings.Join([]string{
		"issue",
		"-R", claimsRepo,
		"edit", claimPrNumber,
		"-b", fmt.Sprintf("\"%s\"", buf.String()),
	}, " "))

}
