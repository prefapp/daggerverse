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

	fmt.Printf("🍄 PR files affected: %v\n", prFiles)

	if err != nil {

		return nil, err

	}

	claimsByYamlChanges := m.FilterClaimsByYamlChanges(ctx, claimsDir, prFiles, ghRepo)

	fmt.Print("🍄 Claims by yaml changes: %v\n", claimsByYamlChanges)

	claimsByTfChanges := m.FilterClaimsByTfChanges(ctx, claimsDir, prFiles)

	fmt.Print("🍄 Claims by tf changes: %v\n", claimsByTfChanges)

	claims := slices.Compact(append(claimsByTfChanges, claimsByYamlChanges...))

	fmt.Printf("🍄 Affected claims: %v\n", claims)

	return claims, nil
}

func (m *NotifyAndHydrateState) FilterClaimsByYamlChanges(

	ctx context.Context,

	claimsDir *Directory,

	prFiles []string,

	ghRepo string,

) []string {

	fmt.Printf("prFiles: %v\n", prFiles)

	affectedClaims := []string{}

	for _, file := range prFiles {

		if !strings.HasSuffix(file, ".yaml") && !strings.HasSuffix(file, ".yml") {

			continue
		}

		contents, err := claimsDir.
			File(file).
			Contents(ctx)

		if err != nil {

			claimName := m.
				GetClaimNameFromDefaultBranch(
					ctx,
					file,
					ghRepo,
				)

			affectedClaims = append(
				affectedClaims,
				claimName,
			)

		} else {

			jsonContents, err := yaml.YAMLToJSON([]byte(contents))

			if err != nil {

				panic(err)

			}

			claimName := gjson.Get(string(jsonContents), "name")

			affectedClaims = append(affectedClaims, claimName.String())
		}

	}

	return affectedClaims
}

func (m *NotifyAndHydrateState) GetClaimNameFromDefaultBranch(ctx context.Context, file string, ghRepo string) string {

	fmt.Printf(
		"Claim not found in pr, getting from default branch: %s\n",
		file,
	)

	yamlContent, err := m.
		GetFileContentFromDefaultBranch(
			ctx,
			ghRepo,
			file,
		)

	jsonContentFromMain, err := yaml.YAMLToJSON([]byte(yamlContent))

	if err != nil {

		panic(err)

	}

	return gjson.
		Get(string(jsonContentFromMain), "name").
		String()

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
