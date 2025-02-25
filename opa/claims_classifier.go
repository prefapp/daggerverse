package main

import (
	"context"
	"dagger/opa/internal/dagger"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

func (m *Opa) ClassifyClaims(ctx context.Context, claimsDir *dagger.Directory) ([]ClaimClassification, error) {

	var entries []string

	for _, ext := range []string{".yml", ".yaml"} {

		extEntries, err := claimsDir.Glob(ctx, fmt.Sprintf("*/*/*/*%s", ext))

		if err != nil {

			return nil, err

		}

		entries = append(entries, extEntries...)
	}

	var classifications []ClaimClassification

	for _, entry := range entries {

		file := claimsDir.File(entry)

		contents, err := file.Contents(ctx)

		if err != nil {

			return nil, err

		}

		claim := &Claim{}

		err = yaml.Unmarshal([]byte(contents), claim)

		if err != nil {

			return nil, err

		}

		splitted := strings.Split(entry, "/")

		classifications = append(
			classifications, ClaimClassification{
				File:         file,
				Name:         claim.Name,
				Kind:         claim.Kind,
				Environment:  splitted[2],
				ResourceType: claim.ResourceType,
				Tenant:       splitted[1],
				Platform:     splitted[0],
				App:          m.App,
			})

		fmt.Printf("Classifying claims in %s\n", entry)

	}

	return classifications, nil

}
