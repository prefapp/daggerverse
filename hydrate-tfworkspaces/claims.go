package main

import (
	"context"

	"gopkg.in/yaml.v3"
)

func (m *HydrateTfworkspaces) GetAppClaimNames(

	ctx context.Context,

) ([]string, error) {

	coordinatesPath := "app-claims/tfworkspaces"

	appClaimsDir := m.ValuesDir.Directory(coordinatesPath)

	claimNamesFromAppDir := []string{}

	entries, err := appClaimsDir.Glob(ctx, "**.yaml")

	if err != nil {

		return nil, err

	}

	for _, entry := range entries {

		claim := Claim{}

		fileContent, err := appClaimsDir.File(entry).Contents(ctx)

		if err != nil {

			return nil, err

		}

		err = yaml.Unmarshal([]byte(fileContent), &claim)

		if err != nil {

			return nil, err

		}

		claimNamesFromAppDir = append(claimNamesFromAppDir, claim.Name)

	}

	return claimNamesFromAppDir, nil

}
