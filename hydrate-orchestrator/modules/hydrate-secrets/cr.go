package main

import (
	"context"
	"dagger/hydrate-secrets/internal/dagger"
	"strings"

	"gopkg.in/yaml.v3"
)

func (m *HydrateSecrets) GetCrsFileByClaimName(ctx context.Context, claimName string, dir *dagger.Directory) ([]*dagger.File, error) {

	entries, err := dir.Glob(ctx, "**.yaml")

	crs := []*dagger.File{}

	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		file := dir.File(entry)
		fileContent, err := file.Contents(ctx)
		if err != nil {
			return nil, err
		}

		cr := Cr{}

		err = yaml.Unmarshal([]byte(fileContent), &cr)

		if err != nil {
			return nil, err
		}

		if strings.Split(cr.Metadata.Annotations.ClaimRef, "/")[1] == claimName {
			crs = append(crs, file)
		}

	}

	return crs, nil

}
