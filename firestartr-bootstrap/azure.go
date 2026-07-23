package main

import (
	"context"
	"fmt"
	"log"
	"strings"
)

// ValidateAzureCredentials verifies that the Azure Service Principal credentials are valid
// by attempting a login with the az CLI inside a Dagger container.
func (m *FirestartrBootstrap) ValidateAzureCredentials(
	ctx context.Context,
) error {
	log.Println("Attempting to validate Azure credentials via service principal login...")

	cfg := m.Creds.CloudProvider.Config

	clientSecretArg := cfg.ClientSecret

	output, err := dag.Container().
		From("mcr.microsoft.com/azure-cli").
		WithExec([]string{
			"az", "login",
			"--service-principal",
			"-u", cfg.ClientId,
			"-p", clientSecretArg,
			"--tenant", cfg.TenantId,
		}).
		Stdout(ctx)

	if err != nil {
		return fmt.Errorf("Azure credential validation failed: az login rejected the service principal credentials: %w", err)
	}

	log.Printf("Azure credentials validated successfully. Output: %s", strings.TrimSpace(output))
	return nil
}

// ValidateAzureStorageAccount verifies that the Azure Blob Storage container used for
// Terraform state exists and is accessible.
func (m *FirestartrBootstrap) ValidateAzureStorageAccount(
	ctx context.Context,
) error {
	cfg := m.Creds.CloudProvider.Config

	log.Printf(
		"Validating Azure Storage Account '%s' (container '%s') in resource group '%s'...",
		cfg.StorageAccountName, cfg.ContainerName, cfg.ResourceGroupName,
	)

	_, err := dag.Container().
		From("mcr.microsoft.com/azure-cli").
		WithExec([]string{
			"az", "login",
			"--service-principal",
			"-u", cfg.ClientId,
			"-p", cfg.ClientSecret,
			"--tenant", cfg.TenantId,
		}).
		WithExec([]string{
			"az", "storage", "container", "show",
			"--account-name", cfg.StorageAccountName,
			"--name", cfg.ContainerName,
			"--auth-mode", "login",
		}).
		Stdout(ctx)

	if err != nil {
		return fmt.Errorf(
			"Azure Storage Account validation failed: container '%s' in account '%s' is not accessible: %w",
			cfg.ContainerName, cfg.StorageAccountName, err,
		)
	}

	log.Printf(
		"Azure Storage Account '%s' (container '%s') validated successfully.",
		cfg.StorageAccountName, cfg.ContainerName,
	)
	return nil
}

// ValidateAzureKeyVault verifies that the Azure Key Vault exists and is accessible
// with the provided service principal credentials.
func (m *FirestartrBootstrap) ValidateAzureKeyVault(
	ctx context.Context,
) error {
	cfg := m.Creds.CloudProvider.Config

	log.Printf("Validating Azure Key Vault '%s'...", cfg.KeyVaultName)

	_, err := dag.Container().
		From("mcr.microsoft.com/azure-cli").
		WithExec([]string{
			"az", "login",
			"--service-principal",
			"-u", cfg.ClientId,
			"-p", cfg.ClientSecret,
			"--tenant", cfg.TenantId,
		}).
		WithExec([]string{
			"az", "keyvault", "show",
			"--name", cfg.KeyVaultName,
		}).
		Stdout(ctx)

	if err != nil {
		return fmt.Errorf(
			"Azure Key Vault validation failed: vault '%s' is not accessible: %w",
			cfg.KeyVaultName, err,
		)
	}

	log.Printf("Azure Key Vault '%s' validated successfully.", cfg.KeyVaultName)
	return nil
}
