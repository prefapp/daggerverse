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

func (m *HydrateTfworkspaces) AddAnnotations(

	ctx context.Context,

	claimName string,

	image string,

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
				claimName,
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
