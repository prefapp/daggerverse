package main

import (
	"context"
	"dagger/hydrate-tfworkspaces/internal/dagger"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"sigs.k8s.io/yaml"
)

func (m *HydrateTfworkspaces) PatchClaimWithInferredProviders(

	ctx context.Context,

	claimName string,

	claimsDir *dagger.Directory,

) (*dagger.Directory, error) {

	entries, err := claimsDir.Glob(ctx, "tfworkspaces/*/*/*/*.yaml")
	//                                   tfworkspaces/platform/tenant/env/claim.yaml

	if err != nil {

		return nil, err

	}

	var foundClaim Claim

	var claimFileContents string

	var platform string

	var tenant string

	var env string

	for _, entry := range entries {

		fileContent, err := claimsDir.File(entry).Contents(ctx)

		if err != nil {

			return nil, err

		}

		claim := Claim{}

		err = yaml.Unmarshal([]byte(fileContent), claim)

		if err != nil {

			return nil, err

		}

		if claim.Name == claimName {

			foundClaim = claim

			claimFileContents = fileContent

			splitted := strings.Split(entry, "/")

			platform = splitted[1]

			tenant = splitted[2]

			env = splitted[3]

			break

		}

	}

	if foundClaim == (Claim{}) || claimFileContents == "" {

		return nil, fmt.Errorf("claim not found")

	}

	if err != nil {

		return nil, err

	}

	return claimsDir, nil

}

// type: tfworkspaces
// name: example-platform
// tenants: [test-tenant]
// envs: [test-env]
// allowedClaims:
//   - resourceTypes:
//       - az-vmss
//     providers:
//       - azure-provider-corpme
//     backend: azure-backend-terraform

func (m *HydrateTfworkspaces) FindProvidersBy(

	ctx context.Context,

	resourceType string,

	platform string,

	tenant string,

	env string,

) (string, error) {

	platforms, err := dag.
		FirestartrConfig(m.DotFirestartrDir).
		Platforms(ctx)

	if err != nil {

		return "", err

	}

	var providers []string

	for _, p := range platforms {

		pName, err := p.Name(ctx)

		if err != nil {

			return "", err

		}

		pTenants, err := p.Tenants(ctx)

		if err != nil {

			return "", err

		}

		pEnvs, err := p.Envs(ctx)

		if err != nil {

			return "", err

		}

		pAllowedClaims, err := p.AllowedClaims(ctx)

		if err != nil {

			return "", err

		}

		if pName == platform && slices.Contains(pTenants, tenant) && slices.Contains(pEnvs, env) {

			providers, err = getProviders(ctx, resourceType, pAllowedClaims)

			if err != nil {

				return "", err

			}

			if len(providers) > 0 {

				marshaledJson, err := json.Marshal(providers)

				return string(marshaledJson), err

			}

		}

	}

	return "", nil
}

func getProviders(ctx context.Context, resourceType string, allowedClaims []dagger.FirestartrConfigAllowedClaim) ([]string, error) {

	for _, allowedClaim := range allowedClaims {

		resourceTypes, err := allowedClaim.ResourceTypes(ctx)

		if err != nil {

			return nil, err

		}

		if slices.Contains(resourceTypes, resourceType) {

			providers, err := allowedClaim.Providers(ctx)

			if err != nil {

				return nil, err

			}

			return providers, nil

		}

	}

	return []string{}, nil

}
