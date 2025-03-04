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

type Providers struct {
	Providers []Provider `json:"providers"`
}

type Provider struct {
	Name string `json:"name"`
}

func (m *HydrateTfworkspaces) PatchClaimWithInferredProviders(

	ctx context.Context,

	claimName string,

	claimsDir *dagger.Directory,

) (*dagger.Directory, error) {

	entries, err := claimsDir.Glob(ctx, "tfworkspaces/*/*/*/*.yaml")
	//                                   tfworkspaces/platform/tenant/env/claim.yaml

	fmt.Printf("🦖 Entries: %s\n", entries)

	if err != nil {

		return nil, err

	}

	var foundClaim *Claim

	var claimFileContents string

	var platform string

	var tenant string

	var env string

	var trappedEntry string

	for _, entry := range entries {

		fileContent, err := claimsDir.File(entry).Contents(ctx)

		if err != nil {

			return nil, err

		}

		claim := &Claim{}

		err = yaml.Unmarshal([]byte(fileContent), claim)

		if err != nil {

			return nil, err

		}

		fmt.Printf("🦖 Claim name %s from entry: %s\n", claim.Name, entry)
		fmt.Printf("Claim name from input: %s\n", claimName)
		if claim.Name == claimName {

			foundClaim = claim

			claimFileContents = fileContent

			trappedEntry = entry

			splitted := strings.Split(entry, "/")

			platform = splitted[1]

			tenant = splitted[2]

			env = splitted[3]

			fmt.Printf("Claim found! 🥮 %s\n", claim.Name)

			fmt.Printf("Platform: %s\n", platform)
			fmt.Printf("Tenant: %s\n", tenant)
			fmt.Printf("Env: %s\n", env)
			break

		}

	}

	if foundClaim == nil || claimFileContents == "" {

		return nil, fmt.Errorf("claim not found")

	}

	providers, err := m.FindProvidersBy(
		ctx,
		foundClaim.ResourceType,
		platform,
		tenant,
		env,
	)
	if err != nil {

		return nil, err

	}

	fmt.Printf("🦖 Providers TO PATCH: %s\n", providers)

	// providers : [{name: "hola"}]

	providersValue := Providers{}

	providersValue.Providers = []Provider{}

	for _, pv := range providers {

		providersValue.Providers = append(providersValue.Providers, Provider{Name: pv})

	}

	providersToJson, err := json.Marshal(providersValue)

	if err != nil {

		return nil, err

	}

	yamlContent, err := m.PatchClaim(
		"/providers/terraform/context",
		string(providersToJson),
		claimFileContents,
	)

	if err != nil {

		return nil, err

	}

	claimsDir = claimsDir.
		WithoutFile(trappedEntry).
		WithNewFile(trappedEntry, yamlContent)

	return claimsDir, nil

}

func (m *HydrateTfworkspaces) FindProvidersBy(

	ctx context.Context,

	resourceType string,

	platform string,

	tenant string,

	env string,

) ([]string, error) {

	platforms, err := dag.
		FirestartrConfig(m.DotFirestartrDir).
		Platforms(ctx)

	if err != nil {

		return nil, err

	}

	var providers []string

	for _, p := range platforms {

		pName, err := p.Name(ctx)

		if err != nil {

			return nil, err

		}

		pTenants, err := p.Tenants(ctx)

		if err != nil {

			return nil, err

		}

		pEnvs, err := p.Envs(ctx)

		if err != nil {

			return nil, err

		}

		pAllowedClaims, err := p.AllowedClaims(ctx)

		if err != nil {

			return nil, err

		}

		fmt.Printf("Coordinates: %s %s %s\n", pName, pTenants, pEnvs)

		if pName == platform && slices.Contains(pTenants, tenant) && slices.Contains(pEnvs, env) {

			providers, err = getProviders(ctx, resourceType, pAllowedClaims)

			fmt.Printf("Providers: %s\n", providers)

			if err != nil {

				return nil, err

			}

			if len(providers) > 0 {

				if err != nil {

					return nil, err

				}

				return providers, nil

			}

		} else {

			fmt.Printf("Skipping platform %s\n", pName)

		}

	}

	return []string{}, nil
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
