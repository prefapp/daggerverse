package main

import (
	"context"
	"strings"

	"gopkg.in/yaml.v3"
)

func (m *HydrateTfworkspaces) GetPreviousImagesFromCrs(ctx context.Context, matrix ImageMatrix) ([]Cr, error) {

	entries, err := m.WetRepoDir.Glob(ctx, "**.yaml")

	if err != nil {

		return nil, err

	}

	crs := []Cr{}

	for _, entry := range entries {

		fileContent, err := m.WetRepoDir.File(entry).Contents(ctx)

		if err != nil {

			return nil, err

		}

		cr := Cr{}

		err = yaml.Unmarshal([]byte(fileContent), &cr)

		if err != nil {

			return nil, err

		}

		if cr.Metadata.Annotations.MicroService != "" && cr.Metadata.Annotations.Image != "" {

			claimName := strings.Split(cr.Metadata.Annotations.ClaimRef, "/")[1]

			if len(matrix.Images) > 0 && claimName == matrix.Images[0].Platform {

				continue
			}

			crs = append(crs, cr)

		}

	}

	return crs, nil
}
