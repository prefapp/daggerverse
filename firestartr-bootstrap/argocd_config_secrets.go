package main

import (
	"context"
	"fmt"
	"log"
	"dagger/firestartr-bootstrap/internal/dagger"
	"gopkg.in/yaml.v3"
)

func (m *FirestartrBootstrap) AddArgoCDSecrets(
	ctx context.Context,
) (*dagger.Directory, error){

	tokenSecret := dag.SetSecret(
		"token",
		m.Creds.GithubApp.OperatorPat,
	)

	argoCDRepo, err := m.CloneRepo(
		
		ctx,
		fmt.Sprintf("firestartr-%s", m.Bootstrap.Env),
		"state-sys-services",
		tokenSecret,
	)
	
	if err != nil {
	
		return nil, fmt.Errorf("Cloning state-sys-services repo: %s", err)
	} 

	secretRef := fmt.Sprintf(
		"/firestartr/%s/fs-%s-argocd/pem",
		m.Bootstrap.Customer,
		m.Bootstrap.Org,
	)

	clientAccess := ClientAccess{
		GithubAppId: m.Creds.ArgoCDCreds.GithubAppId,
		GithubAppInstallationId: m.Creds.ArgoCDCreds.GithubAppInstallationId,
		GithubAppPrivateKey: PrivateKeyReference {
			RemoteRef: secretRef,
		},
	}	

    patchedDir, err := safelyPatchYamlConfig(
		ctx,
		argoCDRepo.Directory("/repo"),
		"kubernetes-sys-services/firestartr-pre/argo-configuration-secrets/values.yaml",
		m.Bootstrap.Customer,
		clientAccess,
	)

	if err != nil {

		return nil, fmt.Errorf("Patching argocd secrets: %s", err)
	}

    err = m.CreatePR(
        ctx,
        "state-sys-services",
        fmt.Sprintf("firestartr-%s", m.Bootstrap.Env),
        patchedDir,
        fmt.Sprintf("automated-add-argocd-secrets-for-%s", m.Bootstrap.Org),
        fmt.Sprintf("feat: add argocd secrets for %s [automated]", m.Bootstrap.Org),
		"",
        tokenSecret,
    )

    if err != nil {
        return nil, fmt.Errorf("Error generating PR for state-sys-services: %s", err)
    }
    

	return patchedDir, nil
}


// SafelyPatchYamlConfig reads a YAML file, modifies only the githubOrgAccess.clients map,
// and writes the entire, modified content back, preserving all other top-level keys.
//
// sourceDirectory: The directory containing the config file.
// fileName: The name of the config file.
// newOrgName: The name of the new organization/client (e.g., "fm-prefapp").
// newClientConfig: The configuration struct for the new client.
func safelyPatchYamlConfig(
    ctx context.Context,
    sourceDirectory *dagger.Directory,
    fileName string,
    newOrgName string,
    newClientConfig ClientAccess,
) (*dagger.Directory, error) {
	
    // --- SLURP (Read File) ---
    yamlContent, err := sourceDirectory.File(fileName).Contents(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to read file %s: %w", fileName, err)
    }

    // Use a generic map to hold the entire configuration, preserving all unknown fields.
    var fullConfig map[string]interface{}
    
    // Unmarshal the YAML content into the generic map
    if err := yaml.Unmarshal([]byte(yamlContent), &fullConfig); err != nil {
        return nil, fmt.Errorf("failed to unmarshal YAML content: %w", err)
    }

    githubAccess, ok := fullConfig["githubOrgAccess"].(map[string]interface{})
    if !ok {
        return nil, fmt.Errorf("failed to find or parse 'githubOrgAccess' as a map")
    }

    clients, ok := githubAccess["clients"].(map[string]interface{})
    if !ok {
        return nil, fmt.Errorf("failed to find or parse 'githubOrgAccess.clients' as a map")
    }
    
    // Check if the new organization already exists
    if _, exists := clients[newOrgName]; exists {
        log.Printf("Organization '%s' already exists in the configuration. Skipping addition.", newOrgName)
    } else {
        log.Printf("Adding new organization: %s", newOrgName)
        
        // 4. Safely add the new client configuration to the 'clients' map.
        // We use the newClientConfig struct (which is marshaled below) directly as the value.
        clients[newOrgName] = newClientConfig
    }

    // Marshal the *full* generic map back into YAML bytes. This preserves all non-modified keys.
    modifiedYAMLBytes, err := yaml.Marshal(&fullConfig)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal modified object back to YAML: %w", err)
    }

    modifiedYAMLString := string(modifiedYAMLBytes)

    // Create a new Directory and write the modified file into it
    outputDirectory := dag.Directory().
        WithNewFile(fileName, modifiedYAMLString)

    return outputDirectory, nil
}
