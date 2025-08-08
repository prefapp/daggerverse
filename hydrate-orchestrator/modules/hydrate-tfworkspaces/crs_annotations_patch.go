package main

import (
	"context"
	"dagger/hydrate-tfworkspaces/internal/dagger"
	"fmt"
	"strings"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"gopkg.in/yaml.v3"
	sigsyaml "sigs.k8s.io/yaml"
)

func (m *HydrateTfworkspaces) AddAnnotationsToCr(

	ctx context.Context,

	claimName string,

	image string,

	path string,

	crsDir *dagger.Directory,

) (*dagger.Directory, error) {

	entries, err := crsDir.Glob(ctx, "**.yaml")

	if err != nil {

		return nil, err

	}

	for _, entry := range entries {

		fileContent, err := crsDir.File(entry).Contents(ctx)

		if err != nil {

			return nil, err

		}

		cr := Cr{}

		err = yaml.Unmarshal([]byte(fileContent), &cr)

		if err != nil {

			return nil, err

		}

		if strings.Split(cr.Metadata.Annotations.ClaimRef, "/")[1] == claimName {

			toJson, err := sigsyaml.YAMLToJSON([]byte(fileContent))

			if err != nil {

				return nil, err

			}

			patchJSON := []byte(fmt.Sprintf(
				`[{"op": "add", "path": "/metadata/annotations/firestartr.dev~1image", "value": "%s"},
				{"op": "add", "path": "/metadata/annotations/firestartr.dev~1microservice", "value": "%s"}]`,
				image,
				path,
			))

			patch, err := jsonpatch.DecodePatch(patchJSON)

			if err != nil {

				panic(err)

			}

			patched, err := patch.Apply([]byte(toJson))

			if err != nil {

				panic(err)

			}

			patchedToYaml, err := sigsyaml.JSONToYAML(patched)

			if err != nil {

				return nil, err

			}

			crsDir = crsDir.
				WithoutFile(entry).
				WithNewFile(entry, string(patchedToYaml))

			break
		}

	}

	return crsDir, nil

}

func (m *HydrateTfworkspaces) AddPrAnnotationToCr(

	ctx context.Context,

	claimName string,

	prNumber string,

	org string,

	repo string,

	crsDir *dagger.Directory,

) (*dagger.Directory, error) {

	entries, err := crsDir.Glob(ctx, "tfworkspaces/*.yaml")

	if err != nil {

		return nil, err

	}

	fmt.Printf("Beginning annotation process...")

	for _, entry := range entries {

		fileContent, err := crsDir.File(entry).Contents(ctx)

		if err != nil {

			return nil, err

		}

		cr := Cr{}

		err = yaml.Unmarshal([]byte(fileContent), &cr)

		if err != nil {

			return nil, err

		}

		fmt.Printf("Checking %s against %s", cr.Metadata.Annotations.ClaimRef, claimName)

		if strings.Split(cr.Metadata.Annotations.ClaimRef, "/")[1] == claimName {

			toJson, err := sigsyaml.YAMLToJSON([]byte(fileContent))

			fmt.Printf("🖊️ Adding annotation to %s\n", entry)

			fmt.Printf("🌀 To JSON: %s\n", toJson)

			if err != nil {

				return nil, err

			}

			annotationValue := sanitizePrString(fmt.Sprintf("%s/%s#%s", org, repo, prNumber))

			fmt.Printf("Adding annotation %s to %s\n", annotationValue, entry)

			patchJSON := []byte(fmt.Sprintf(
				`[{"op": "add", "path": "/metadata/annotations/firestartr.dev~1last-state-pr", "value": "%s"}]`,
				annotationValue,
			))

			fmt.Printf("｛🖋️} Patch JSON: %s\n", patchJSON)

			patch, err := jsonpatch.DecodePatch(patchJSON)

			if err != nil {

				panic(err)

			}

			patched, err := patch.Apply([]byte(toJson))

			if err != nil {

				panic(err)

			}

			patchedToYaml, err := sigsyaml.JSONToYAML(patched)

			if err != nil {

				return nil, err

			}

			crsDir = crsDir.
				WithoutFile(entry).
				WithNewFile(entry, string(patchedToYaml))

			break
		}

	}

	return crsDir, nil

}

func sanitizePrString(pr string) string {
	pr = strings.ReplaceAll(pr, "\n", "")
	pr = strings.ReplaceAll(pr, " ", "")
	pr = strings.ReplaceAll(pr, "\t", "")

	return pr
}
