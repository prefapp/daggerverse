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

type PrFiles struct {
	AddedModified []string
	Deleted       []string
}

func (m *NotifyAndHydrateState) GetPrChangedFiles(

	ctx context.Context,

	claimsRepo *Directory,

) (PrFiles, error) {

	result := PrFiles{
		AddedModified: []string{},
		Deleted:       []string{},
	}

	c, err := dag.
		Container().
		From("alpine/git").
		WithMountedDirectory("/repo", claimsRepo).
		WithWorkdir("/repo").
		Sync(ctx)

	amResp, err := c.
		WithExec([]string{
			"diff",
			"origin/" + m.ClaimsDefaultBranch,
			"-M90%",
			"--name-only",
			"--diff-filter=AM",
		}).
		Stdout(ctx)

	if err != nil {

		return result, err

	}

	for _, line := range strings.Split(amResp, "\n") {

		if line == "" {

			continue

		}

		result.AddedModified = append(result.AddedModified, line)

	}

	dResp, err := c.
		WithExec([]string{
			"diff",
			"origin/" + m.ClaimsDefaultBranch,
			"-M90%",
			"--name-only",
			"--diff-filter=D",
		}).
		Stdout(ctx)

	if err != nil {

		return result, err

	}

	for _, line := range strings.Split(dResp, "\n") {

		if line == "" {

			continue

		}

		result.Deleted = append(result.Deleted, line)

	}

	return result, nil
}

func (m *NotifyAndHydrateState) GetAffectedClaims(ctx context.Context,

	ghRepo string,

	prNumber string,

	claimsDir *Directory,

) ([]string, error) {

	fsLog("Getting affected claims")

	prFiles, err := m.GetPrChangedFiles(ctx, claimsDir)

	if err != nil {

		return nil, err

	}

	claimsByYamlChanges := m.
		FilterClaimsByYamlChanges(
			ctx,
			claimsDir,
			prFiles.Deleted,
			prFiles.AddedModified,
			ghRepo,
		)

	claimsByTfChanges := m.
		FilterClaimsByTfChanges(ctx,
			claimsDir,
			prFiles.AddedModified,
		)

	claims := slices.Compact(append(claimsByTfChanges, claimsByYamlChanges...))

	return claims, nil
}

func (m *NotifyAndHydrateState) FilterClaimsByYamlChanges(

	ctx context.Context,

	claimsDir *Directory,

	deletedFiles []string,

	addedOrModifiedFiles []string,

	ghRepo string,

) []string {

	result := []string{}

	for _, file := range addedOrModifiedFiles {

		if !isYaml(file) {

			continue

		}

		claimName := m.ReadClaimNameFromFile(ctx, claimsDir, file)

		result = append(result, claimName)

	}

	for _, file := range deletedFiles {

		if !isYaml(file) {

			continue
		}

		claimName := m.
			GetClaimNameFromDefaultBranch(
				ctx,
				file,
				ghRepo,
			)

		result = append(
			result,
			claimName,
		)
	}

	return result
}

func (*NotifyAndHydrateState) ReadClaimNameFromFile(ctx context.Context, claimsDir *Directory, file string) string {

	contents, err := claimsDir.
		File(file).
		Contents(ctx)

	jsonContents, err := yaml.
		YAMLToJSON([]byte(contents))

	if err != nil {

		panic(err)

	}

	return gjson.
		Get(string(jsonContents), "name").
		String()

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

func isYaml(file string) bool {
	return filepath.Ext(file) == ".yaml" || filepath.Ext(file) == ".yml"
}
