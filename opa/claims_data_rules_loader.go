package main

import (
	"context"
	"dagger/opa/internal/dagger"
	"fmt"

	"gopkg.in/yaml.v3"
)

func (m *Opa) LoadDataRules(ctx context.Context, validationsDir *dagger.Directory, app string) ([]ClaimsDataRules, error) {

	var data []ClaimsDataRules

	for _, ext := range []string{".yml", ".yaml"} {

		entries, err := validationsDir.Glob(ctx, fmt.Sprintf("apps/%s/tfworkspaces/**/*%s", app, ext))

		if err != nil {

			return nil, err

		}

		for _, entry := range entries {

			file := validationsDir.File(entry)

			contents, err := file.Contents(ctx)

			if err != nil {

				return nil, err

			}

			claimsData := &ClaimsDataRules{}

			err = yaml.Unmarshal([]byte(contents), claimsData)

			if err != nil {

				return nil, err

			}

			claimsData.File = file

			data = append(data, *claimsData)

		}
	}

	return data, nil
}
