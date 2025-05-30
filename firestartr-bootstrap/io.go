package main

import (
	"context"
	"dagger/firestartr-bootstrap/internal/dagger"
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type Claim struct {
	Kind string `yaml:"kind"`
	Name string `yaml:"name"`
}

type Cr struct {
	Kind     string   `yaml:"kind"`
	Metadata Metadata `yaml:"metadata"`
}

type Metadata struct {
	Name string `yaml:"name"`
}

func (m *FirestartrBootstrap) SplitRenderedClaimsInFiles(renderedContent string) (*dagger.Directory, error) {

	fmt.Printf("Rendered Content: %s\n", renderedContent)

	claims := regexp.MustCompile(`(?m)^(---\n)`).Split(string(renderedContent), -1)

	dir := dag.Directory()

	for _, manifest := range claims {
		claim := Claim{}

		err := yaml.Unmarshal([]byte(manifest), &claim)
		if err != nil {
			return nil, err
		}

		if claim.Kind == "" && claim.Name == "" {
			continue
		}

		fileName := fmt.Sprintf("%s.yaml", claim.Name)

		manifest = "---\n" + manifest

		pathFile := fmt.Sprintf("claims/%s/%s", getPathByKind(claim.Kind), fileName)

		dir = dir.WithNewFile(pathFile, manifest)
	}

	return dir, nil
}

func getPathByKind(kind string) string {
	mapKindPath := map[string]string{
		"ComponentClaim": "components",
		"GroupClaims":    "groups",
	}

	if path, ok := mapKindPath[kind]; ok {
		return path
	} else {
		panic(fmt.Sprintf("Unknown kind: %s", kind))
	}
}

func (m *FirestartrBootstrap) SplitRenderedCrsInFiles(renderedContent string) (*dagger.Directory, error) {
	fmt.Printf("Rendered Content: %s\n", renderedContent)

	claims := regexp.MustCompile(`(?m)^(---\n)`).Split(string(renderedContent), -1)

	dir := dag.Directory()

	for _, manifest := range claims {
		cr := Cr{}

		err := yaml.Unmarshal([]byte(manifest), &cr)
		if err != nil {
			return nil, err
		}

		if cr.Kind == "" && cr.Metadata.Name == "" {
			continue
		}

		fileName := fmt.Sprintf("%s.%s.yml", cr.Kind, cr.Metadata.Name)

		manifest = "---\n" + manifest

		dir = dir.WithNewFile(fileName, manifest)
	}

	return dir, nil
}

func loadCredsFile(ctx context.Context, creds *dagger.Secret) (*CredsFile, error) {
	credsContent, err := creds.Plaintext(ctx)
	if err != nil {
		return nil, err
	}

	credsFile := &CredsFile{}

	err = yaml.Unmarshal([]byte(credsContent), credsFile)
	if err != nil {
		return nil, err
	}

	escaped := strings.ReplaceAll(credsFile.GithubApp.Pem, "\n", "\\n")

	credsFile.GithubApp.RawPem = escaped

	return credsFile, nil
}
