package main

import (
	"context"
	"dagger/hydrate-tfworkspaces/internal/dagger"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

func (m *HydrateTfworkspaces) GetCrFileByClaimName(ctx context.Context, claimName string, dir *dagger.Directory) (*dagger.File, error) {

	entries, err := dir.Glob(ctx, "**.yaml")

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

			return file, nil

		}

	}

	return nil, fmt.Errorf("cr from claim name %s not found in redered crs", claimName)

}
