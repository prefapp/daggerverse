package main

import (
	"context"
	"dagger/hydrate-tfworkspaces/internal/dagger"
	"fmt"
	"strings"

	jsonpatch "github.com/evanphx/json-patch/v5"

	"sigs.k8s.io/yaml"
	sigsyaml "sigs.k8s.io/yaml"
)

func (m *HydrateTfworkspaces) PatchClaimWithNewImageValues(ctx context.Context, matrix ImageMatrix, appDir *dagger.Directory) (*dagger.Directory, error) {

	if len(matrix.Images) == 0 {

		return appDir, nil

	}

	imageData := matrix.Images[0]

	entries, err := appDir.Glob(ctx, "**.yaml")

	if err != nil {

		panic(err)
	}

	jsonObj := ""

	trappedEntry := ""

	for _, entry := range entries {

		fileContent, err := appDir.File(entry).Contents(ctx)

		if err != nil {

			return nil, err

		}

		claim := Claim{}

		err = yaml.Unmarshal([]byte(fileContent), &claim)

		if err != nil {

			return nil, err

		}

		if claim.Name == imageData.Platform {

			trappedEntry = entry

			tojson, err := yaml.YAMLToJSON([]byte(fileContent))

			if err != nil {

				return nil, err

			}

			jsonObj = string(tojson)

			break
		}

	}

	if jsonObj == "" {

		return nil, fmt.Errorf("no claim found for platform %s", imageData.Platform)

	}

	if len(entries) == 0 {

		return nil, fmt.Errorf("no claims found in app dir")

	}

	for _, imgKey := range imageData.ImageKeys {

		patchJSON := []byte(fmt.Sprintf(
			`[{"op": "add", "path": "/providers/terraform/values/%s", "value": "%s"}]`,
			imgKey,
			imageData.Image,
		))

		patch, err := jsonpatch.DecodePatch(patchJSON)

		if err != nil {

			panic(err)

		}

		modified, err := patch.Apply([]byte(jsonObj))

		jsonObj = string(modified)

		if err != nil {

			panic(err)

		}
	}

	modifiedYaml, err := yaml.JSONToYAML([]byte(jsonObj))

	if err != nil {

		return nil, err

	}

	appDir = appDir.
		WithoutFile(trappedEntry).
		WithNewFile(trappedEntry, string(modifiedYaml))

	return appDir, nil

}

func (m *HydrateTfworkspaces) PatchClaimWithPreviousImages(

	ctx context.Context,

	cr *Cr,

	appClaimsDir *dagger.Directory,

) (*dagger.Directory, error) {

	entries, err := appClaimsDir.Glob(ctx, "**.yaml")

	if err != nil {

		return nil, err

	}

	fmt.Printf(
		"🔍 Looking for claim %s to patch it with previous image\n",
		cr.Metadata.Annotations.ClaimRef,
	)

	for _, entry := range entries {

		fileContent, err := appClaimsDir.File(entry).Contents(ctx)

		if err != nil {

			panic(err)

		}

		claim := Claim{}

		err = yaml.Unmarshal([]byte(fileContent), &claim)

		if err != nil {

			panic(err)

		}

		if claim.Name == strings.Split(cr.Metadata.Annotations.ClaimRef, "/")[1] {

			fmt.Printf("🔍 Found claim %s\n", claim.Name)

			contentsFile, err := appClaimsDir.File(entry).Contents(ctx)

			if err != nil {

				return nil, err

			}

			patchedClaim, err := m.PatchClaim(
				cr.Metadata.Annotations.MicroServicePointer,
				cr.Metadata.Annotations.Image,
				contentsFile,
			)

			if err != nil {

				return nil, err

			}

			fmt.Printf("🔧 Patching claim %s with previous images\n", claim.Name)
			fmt.Printf("🍀 PATCHED CLAIM: %s\n", patchedClaim)
			appClaimsDir = appClaimsDir.
				WithoutFile(entry).
				WithNewFile(entry, patchedClaim)

		}

	}

	return appClaimsDir, nil
}

func (m *HydrateTfworkspaces) PatchClaim(

	path string,

	value string,

	yamlContent string,

) (string, error) {

	tojson, err := sigsyaml.YAMLToJSON([]byte(yamlContent))

	if err != nil {

		return "", err

	}

	patchJSON := []byte(fmt.Sprintf(
		`[{"op": "add", "path": "/providers/terraform/values/%s", "value": "%s"}]`,
		path,
		value,
	))

	patch, err := jsonpatch.DecodePatch(patchJSON)

	if err != nil {

		panic(err)

	}

	modifiedJson, err := patch.Apply([]byte(tojson))

	if err != nil {

		return "", err

	}

	modifiedYaml, err := sigsyaml.JSONToYAML(modifiedJson)

	if err != nil {

		return "", err

	}

	return string(modifiedYaml), nil

}
