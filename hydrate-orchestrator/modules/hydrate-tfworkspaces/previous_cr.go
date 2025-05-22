package main

import (
	"context"
	"strings"

	"gopkg.in/yaml.v3"
)

func (m *HydrateTfworkspaces) GetPreviousCr(ctx context.Context, claimName string) (*Cr, error) {

	entries, err := m.WetRepoDir.Glob(ctx, "**.yaml")

	if err != nil {

		return nil, err

	}

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

		if cr.Metadata.Annotations.MicroServicePointer != "" && cr.Metadata.Annotations.Image != "" {

			claimNameFromRef := strings.Split(cr.Metadata.Annotations.ClaimRef, "/")[1]

			if claimName == claimNameFromRef {

				return &cr, nil
			}

		}

	}

	return nil, nil
}
