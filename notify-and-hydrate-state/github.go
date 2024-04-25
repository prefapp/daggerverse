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

) (string, error) {

	//panic(claimsRepo + "---" + claimPrNumber + "---" + strings.Join(prLinks, "---"))

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
		"issue",
		"-R", claimsRepo,
		"edit", claimPrNumber,
		"-b", fmt.Sprintf("\"%s\"", buf.String()),
	}, " "), GhRunOpts{DisableCache: true})

}
