package main

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"path/filepath"

	"github.com/tidwall/gjson"
	"sigs.k8s.io/yaml"
)

func (m *NotifyAndHydrateState) GetPrChangedFiles(

	ctx context.Context,

	ghRepo string,

	prNumber string,

) ([]string, error) {

	command := strings.Join([]string{"pr", "view", prNumber, "--json", "files", "-R", ghRepo}, " ")

	content, err := dag.Gh().Run(ctx, m.GhToken, command, GhRunOpts{DisableCache: true})

	if err != nil {

		panic(err)
	}

	files := []string{}

	for _, fileName := range gjson.Get(content, "files.#.path").Array() {

		files = append(files, fileName.String())
	}

	fmt.Printf("PR changed files: %v\n", files)

	return files, nil
}

func (m *NotifyAndHydrateState) GetAffectedClaims(ctx context.Context,

	ghRepo string,

	prNumber string,

	claimsDir *Directory,

) ([]string, error) {

	prFiles, err := m.GetPrChangedFiles(ctx, ghRepo, prNumber)

	if err != nil {

		return nil, err

	}

	claimsByYamlChanges := m.FilterClaimsByYamlChanges(ctx, claimsDir, prFiles)

	claimsByTfChanges := m.FilterClaimsByTfChanges(ctx, claimsDir, prFiles)

	claims := slices.Compact(append(claimsByTfChanges, claimsByYamlChanges...))

	return claims, nil
}

func (m *NotifyAndHydrateState) FilterClaimsByYamlChanges(

	ctx context.Context,

	claimsDir *Directory,

	prFiles []string,

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
