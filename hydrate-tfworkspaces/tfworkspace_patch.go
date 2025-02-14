package main

import (
	"context"
	"dagger/hydrate-tfworkspaces/internal/dagger"
	"encoding/json"
	"fmt"

	jsonpatch "github.com/evanphx/json-patch/v5"

	"sigs.k8s.io/yaml"
)

func (m *HydrateTfworkspaces) UpdateKeyInTfWorkspace(ctx context.Context, jsonMatrix string, appDir *dagger.Directory) (*dagger.Directory, error) {

	matrix := ImageMatrix{}

	err := json.Unmarshal([]byte(jsonMatrix), &matrix)

	if err != nil {

		return nil, err

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

		return nil, fmt.Errorf("no claim found in app dir with platform %s", imageData.Platform)
	}

	if len(entries) == 0 {

		return nil, fmt.Errorf("no claims found in app dir with id %s", appDir.ID)

	}

	for _, imgKey := range imageData.ImageKeys {

		patchJSON := []byte(fmt.Sprintf(`[{"op": "add", "path": "/providers/terraform/values/%s", "value": "%s"}]`, imgKey, imageData.Image))

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
