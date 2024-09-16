package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"slices"
	"strings"

	"path/filepath"

	"github.com/tidwall/gjson"
	"sigs.k8s.io/yaml"
)

func (m *NotifyAndHydrateState) GetPrChangedFiles(

	ctx context.Context,

	claimsRepo *Directory,

) ([]string, error) {

	result := []string{}

	resp, err := dag.
		Container().
		From("alpine/git").
		WithMountedDirectory("/repo", claimsRepo).
		WithWorkdir("/repo").
		WithExec([]string{
			"diff",
			"origin/" + m.ClaimsDefaultBranch,
			"-M90%",
			"--name-only",
		}).
		Stdout(ctx)

	if err != nil {

		return nil, err

	}

	for _, line := range strings.Split(resp, "\n") {

		result = append(result, line)

	}

	return result, nil
}

func (m *NotifyAndHydrateState) GetAffectedClaims(ctx context.Context,

	ghRepo string,

	prNumber string,

	claimsDir *Directory,

) ([]string, error) {

	prFiles, err := m.GetPrChangedFiles(ctx, claimsDir)

	if err != nil {

		return nil, err

	}

	claimsByYamlChanges := m.FilterClaimsByYamlChanges(ctx, claimsDir, prFiles, ghRepo)

	claimsByTfChanges := m.FilterClaimsByTfChanges(ctx, claimsDir, prFiles)

	claims := slices.Compact(append(claimsByTfChanges, claimsByYamlChanges...))

	fmt.Printf("Affected claims: %v\n", claims)

	return claims, nil
}

func (m *NotifyAndHydrateState) FilterClaimsByYamlChanges(

	ctx context.Context,

	claimsDir *Directory,

	prFiles []string,

	ghRepo string,

) []string {

	fmt.Printf("prFiles: %v\n", prFiles)

	entries, err := claimsDir.
		WithoutDirectory(".git").
		WithoutDirectory(".github").
		WithoutDirectory(".config").
		Glob(ctx, "**/*.yaml")

	fmt.Printf("entries: %v\n", entries)

	if err != nil {

		panic(err)

	}

	if err != nil {

		panic(err)

	}

	affectedClaims := []string{}

	for _, file := range prFiles {

		for _, entry := range entries {

			if strings.Contains(entry, file) {

				// get contents of the file
				contents, err := claimsDir.File(entry).Contents(ctx)

				if err != nil {

					panic(err)

				}

				jsonContents, err := yaml.YAMLToJSON([]byte(contents))

				claimName := gjson.Get(string(jsonContents), "name")

				affectedClaims = append(affectedClaims, claimName.String())

			} else {

				yamlContent, err := m.GetFileContentFromDefaultBranch(ctx, ghRepo, file)

				if err != nil {

					fmt.Printf("CR not found in the main branch: %s\n", file)

				} else {

					jsonContentFromMain, err := yaml.YAMLToJSON([]byte(yamlContent))

					if err != nil {

						panic(err)

					}

					claimName := gjson.Get(string(jsonContentFromMain), "name")

					affectedClaims = append(affectedClaims, claimName.String())

				}
			}
		}

	}

	fmt.Printf("Affected claims: %v\n", affectedClaims)

	return affectedClaims

}

func (m *NotifyAndHydrateState) FilterClaimsByTfChanges(

	ctx context.Context,

	claimsDir *Directory,

	prFiles []string,

) []string {

	entries, err := claimsDir.
		WithoutDirectory(".git").
		WithoutDirectory(".github").
		WithoutDirectory(".config").
		Glob(ctx, "**/*.tf")

	if err != nil {

		panic(err)

	}

	affectedClaims := []string{}

	for _, file := range prFiles {

		for _, entry := range entries {

			if strings.Contains(entry, file) {

				tfDirPath := filepath.Dir(entry)

				tfDir := claimsDir.Directory(tfDirPath)

				tfDirEntries, err := tfDir.Glob(ctx, "*.yaml")

				if err != nil {

					panic(err)

				}

				if len(tfDirEntries) > 1 {

					panic(fmt.Errorf("More than one yaml file in the directory: %s", tfDirPath))

				} else if len(tfDirEntries) == 0 {

					continue

				}

				contents, err := tfDir.File(tfDirEntries[0]).Contents(ctx)

				if err != nil {

					panic(err)

				}

				jsonContents, err := yaml.YAMLToJSON([]byte(contents))

				source := gjson.Get(string(jsonContents), "providers.terraform.source")

				module := gjson.Get(string(jsonContents), "providers.terraform.module")

				if strings.ToLower(source.String()) == "inline" && !module.Exists() {

					claimName := gjson.Get(string(jsonContents), "name")

					affectedClaims = append(affectedClaims, claimName.String())
				}
			}
		}

	}

	fmt.Printf("Affected claims: %v\n", affectedClaims)

	return affectedClaims

}

// # GitHub CLI api
// # https://cli.github.com/manual/gh_api

//	gh api \
//	  -H "Accept: application/vnd.github+json" \
//	  -H "X-GitHub-Api-Version: 2022-11-28" \
//	  /repos/OWNER/REPO/contents/PATH
func (m *NotifyAndHydrateState) GetFileContentFromDefaultBranch(

	ctx context.Context,

	// +default="claims"
	repo string,

	// +default="claims/tfworkspaces/test-module-a.yaml"
	path string,

) (string, error) {

	endpoint := fmt.Sprintf(
		"/repos/%s/contents/%s",
		repo,
		path,
	)

	command := strings.Join([]string{
		"api",
		"-H \"Accept: application/vnd.github+json\"",
		"-H \"X-GitHub-Api-Version: 2022-11-28\"",
		endpoint,
	},
		" ",
	)

	jsonResp, err := dag.
		Gh().
		Run(
			ctx,
			m.GhToken,
			command,
			GhRunOpts{DisableCache: true},
		)

	if err != nil {

		return "", err
	}

	b64Content := gjson.
		Get(string(jsonResp), "content").
		String()

	return base64Decode(b64Content), nil
}

func base64Decode(str string) string {
	data, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		panic(err)
	}
	return string(data)
}
