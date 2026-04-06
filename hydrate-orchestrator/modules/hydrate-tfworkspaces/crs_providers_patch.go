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

type Context struct {
	Providers []Provider `json:"providers"`
	Backend   Backend    `json:"backend"`
}

type Provider struct {
	Name string `json:"name"`
}

type Backend struct {
	Name string `json:"name"`
}

func (m *HydrateTfworkspaces) PatchClaimWithInferredProviders(

	ctx context.Context,

	claimName string,

	claimsDir *dagger.Directory,

) (*dagger.Directory, error) {

	entries, err := claimsDir.Glob(ctx, "*/*/*/*.yaml")
	//                                  ./platform/tenant/env/claim.yaml

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

			platform = splitted[0]

			tenant = splitted[1]

			env = splitted[2]

			fmt.Printf("Claim found! 🥮 %s\n", claim.Name)

			fmt.Printf("Platform: %s\n", platform)

			fmt.Printf("Tenant: %s\n", tenant)

			fmt.Printf("Env: %s\n", env)

		} else {

			fmt.Println(
				"claim does not match, patching with dummy providers",
			)

			yamlContent, err := m.PatchClaim(
				"/providers/terraform/context",
				`{"providers": [{"name": "dummy"}], "backend": {"name": "dummy"}}`,
				fileContent,
			)

			fmt.Printf("🦖 Patched dummy claim: %s\n", yamlContent)

			if err != nil {

				return nil, err

			}

			claimsDir = claimsDir.
				WithoutFile(entry).
				WithNewFile(entry, yamlContent)

			contents, _ := claimsDir.File(entry).Contents(ctx)

			fmt.Printf("🦖 Contents: %s\n", contents)

		}

	}

	if foundClaim == nil || claimFileContents == "" {

		return nil, fmt.Errorf("claim %s not found", claimName)

	}

	context, err := m.FindProvidersBy(
		ctx,
		foundClaim.ResourceType,
		platform,
		tenant,
		env,
	)
	if err != nil {

		return nil, err

	}

	contextToJson, err := json.Marshal(context)

	if err != nil {

		return nil, err

	}

	fmt.Printf("🦖 Context patch: %s", contextToJson)

	yamlContent, err := m.PatchClaim(
		"/providers/terraform/context",
		string(contextToJson),
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

) (*Context, error) {

	platforms, err := dag.
		FirestartrConfig(m.DotFirestartrDir).
		Platforms(ctx)

	if err != nil {

		return nil, err

	}

	var context *Context

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

		if pName == platform && slices.Contains(pTenants, tenant) && slices.Contains(pEnvs, env) {

			context, err = getContext(ctx, resourceType, pAllowedClaims)

			if err != nil {

				return nil, err

			}

			if len(context.Providers) > 0 {

				return context, nil

			}

		} else {

			fmt.Printf("Skipping platform %s\n", pName)

		}

	}

	return context, nil
}

func getContext(

	ctx context.Context,

	resourceType string,

	allowedClaims []dagger.FirestartrConfigAllowedClaim,

) (*Context, error) {

	context := &Context{}

	fmt.Printf("🦖 Allowed claims: %v\n", allowedClaims)

	for _, allowedClaim := range allowedClaims {

		resourceTypes, err := allowedClaim.ResourceTypes(ctx)

		if err != nil {

			return nil, err

		}

		if slices.Contains(resourceTypes, resourceType) {

			fmt.Printf("Found allowed claim for resource type %s\n", resourceType)

			providers, err := allowedClaim.Providers(ctx)

			if err != nil {

				return nil, err

			}

			for _, provider := range providers {

				context.Providers = append(context.Providers, Provider{Name: provider})

			}

			backend, err := allowedClaim.Backend(ctx)

			if err != nil {

				return nil, err

			}

			context.Backend = Backend{Name: backend}

			return context, nil

		} else {
			fmt.Printf("No allowed claim for resource type %s\n", resourceType)
		}

	}

	return context, nil

}
