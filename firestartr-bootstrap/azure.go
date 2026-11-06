package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"bufio"
	"dagger/firestartr-bootstrap/internal/dagger"
)

// azureTokenResponse is the JSON payload returned by the Azure AD token endpoint.
type azureTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Error       string `json:"error"`
	ErrorDesc   string `json:"error_description"`
}

// kvSecretItem is a single entry in the Key Vault secrets list response.
type kvSecretItem struct {
	ID string `json:"id"`
}

// kvSecretList is a page of results from the Key Vault secrets list endpoint.
type kvSecretList struct {
	Value    []kvSecretItem `json:"value"`
	NextLink string         `json:"nextLink"`
}

// loginAzureKV authenticates the bootstrap App Registration (Service Principal)
// and returns a Bearer token scoped to Azure Key Vault.
// Equivalent to loginAWS for the dedicated deployment path.
func loginAzureKV(ctx context.Context, creds *CredsFile) (string, error) {
	cfg := creds.CloudProvider.Config
	tokenURL := fmt.Sprintf(
		"https://login.microsoftonline.com/%s/oauth2/v2.0/token",
		cfg.TenantId,
	)

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", cfg.BootstrapClientId)
	data.Set("client_secret", cfg.BootstrapClientSecret)
	data.Set("scope", "https://vault.azure.net/.default")

	req, err := http.NewRequestWithContext(
		ctx, http.MethodPost, tokenURL, strings.NewReader(data.Encode()),
	)
	if err != nil {
		return "", fmt.Errorf("building Azure AD token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("calling Azure AD token endpoint: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading Azure AD token response: %w", err)
	}

	var tokenResp azureTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("parsing Azure AD token response: %w", err)
	}

	if tokenResp.Error != "" {
		return "", fmt.Errorf(
			"Azure AD authentication failed: %s — %s",
			tokenResp.Error, tokenResp.ErrorDesc,
		)
	}

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf(
			"Azure AD returned an empty access token (HTTP %d)", resp.StatusCode,
		)
	}

	return tokenResp.AccessToken, nil
}

// ValidateAzureSPCredentials validates that the bootstrap App Registration credentials
// (bootstrap_client_id + bootstrap_client_secret) can successfully authenticate to
// Azure AD and obtain a Key Vault scoped token.
// Equivalent to ValidateSTSCredentials for the dedicated deployment path.
func (m *FirestartrBootstrap) ValidateAzureSPCredentials(ctx context.Context) error {
	log.Println("Attempting to validate Azure bootstrap SP credentials...")

	_, err := loginAzureKV(ctx, m.Creds)
	if err != nil {
		return fmt.Errorf("bootstrap SP credentials are invalid: %w", err)
	}

	log.Printf("✅ Azure bootstrap SP credentials validated successfully.")
	return nil
}

// expectedAzureKVSecrets returns the Key Vault secret names that must exist before
// bootstrap can run. These are the secrets that ESO will pull into the kind cluster
// via the azure_bootstrap_secrets and azure_operator_secrets ExternalSecret CRs.
func (m *FirestartrBootstrap) expectedAzureKVSecrets() []string {
	org := m.GhOrgLowerCase
	return []string{
		"fs-admin-pem",
		"fs-admin-app-id",
		fmt.Sprintf("fs-admin-%s-installation-id", org),
		"fs-pem",
		"fs-app-id",
		fmt.Sprintf("fs-%s-installation-id", org),
	}
}

// listKVSecretNames retrieves all secret names from an Azure Key Vault, following
// pagination via nextLink. Secret values are never fetched — only names.
func listKVSecretNames(ctx context.Context, vaultURL, token string) ([]string, error) {
	names := []string{}
	nextURL := fmt.Sprintf("%s/secrets?api-version=7.4", vaultURL)

	for nextURL != "" {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, nextURL, nil)
		if err != nil {
			return nil, fmt.Errorf("building Key Vault list request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("calling Key Vault list endpoint: %w", err)
		}

		if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
			resp.Body.Close()
			return nil, fmt.Errorf(
				"bootstrap SP lacks Key Vault list permission (HTTP %d). "+
					"Ensure the App Registration has 'Key Vault Secrets User' on vault %q",
				resp.StatusCode, vaultURL,
			)
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf(
				"unexpected HTTP %d from Key Vault list endpoint", resp.StatusCode,
			)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("reading Key Vault list response: %w", err)
		}

		var page kvSecretList
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, fmt.Errorf("parsing Key Vault list response: %w", err)
		}

		for _, item := range page.Value {
			// ID format: https://<vault>.vault.azure.net/secrets/<name>
			parts := strings.Split(item.ID, "/")
			if len(parts) > 0 {
				names = append(names, parts[len(parts)-1])
			}
		}

		nextURL = page.NextLink
	}

	return names, nil
}

// ValidateAzureKeyVaultSecrets validates that the Key Vault is reachable with the
// bootstrap SP credentials and that all secrets required by the bootstrap ExternalSecret
// CRs are present. Equivalent to ValidateParameters for the dedicated deployment path.
func (m *FirestartrBootstrap) ValidateAzureKeyVaultSecrets(ctx context.Context) error {
	log.Println("Validating Azure Key Vault access and required secrets...")

	token, err := loginAzureKV(ctx, m.Creds)
	if err != nil {
		return fmt.Errorf("obtaining Key Vault token for secret validation: %w", err)
	}

	cfg := m.Creds.CloudProvider.Config
	vaultURL := fmt.Sprintf("https://%s.vault.azure.net", cfg.KeyVaultName)

	existing, err := listKVSecretNames(ctx, vaultURL, token)
	if err != nil {
		return fmt.Errorf("listing Key Vault secrets: %w", err)
	}

	existingSet := make(map[string]struct{}, len(existing))
	for _, name := range existing {
		existingSet[name] = struct{}{}
	}

	missing := []string{}
	for _, required := range m.expectedAzureKVSecrets() {
		if _, ok := existingSet[required]; ok {
			log.Printf("✅ Found required Key Vault secret: %s", required)
		} else {
			log.Printf("❌ Missing Key Vault secret: %s", required)
			missing = append(missing, required)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf(
			"Key Vault validation failed. The following secrets are missing from %q:\n - %s",
			cfg.KeyVaultName,
			strings.Join(missing, "\n - "),
		)
	}

	log.Println("✅ All required Key Vault secrets validated successfully.")
	return nil
}

// loginAzureARM authenticates the bootstrap App Registration and returns a
// Bearer token scoped to the Azure Resource Manager API.
// Required to read Managed Identity properties (e.g. client_id) from ARM.
func loginAzureARM(ctx context.Context, creds *CredsFile) (string, error) {
	cfg := creds.CloudProvider.Config
	tokenURL := fmt.Sprintf(
		"https://login.microsoftonline.com/%s/oauth2/v2.0/token",
		cfg.TenantId,
	)

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", cfg.BootstrapClientId)
	data.Set("client_secret", cfg.BootstrapClientSecret)
	data.Set("scope", "https://management.azure.com/.default")

	req, err := http.NewRequestWithContext(
		ctx, http.MethodPost, tokenURL, strings.NewReader(data.Encode()),
	)
	if err != nil {
		return "", fmt.Errorf("building Azure AD ARM token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("calling Azure AD token endpoint for ARM: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading Azure AD ARM token response: %w", err)
	}

	var tokenResp azureTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("parsing Azure AD ARM token response: %w", err)
	}

	if tokenResp.Error != "" {
		return "", fmt.Errorf(
			"Azure AD ARM authentication failed: %s — %s",
			tokenResp.Error, tokenResp.ErrorDesc,
		)
	}

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf(
			"Azure AD returned an empty ARM access token (HTTP %d)", resp.StatusCode,
		)
	}

	return tokenResp.AccessToken, nil
}

// miARMResponse is the subset of the ARM GET response for a User Assigned
// Managed Identity that we care about.
type miARMResponse struct {
	Properties struct {
		ClientId    string `json:"clientId"`
		PrincipalId string `json:"principalId"`
		TenantId    string `json:"tenantId"`
	} `json:"properties"`
}

// aksARMResponse is the subset of the ARM GET response for a Managed Cluster
// that we need to read the OIDC issuer URL.
type aksARMResponse struct {
	Properties struct {
		OidcIssuerProfile struct {
			IssuerURL string `json:"issuerURL"`
		} `json:"oidcIssuerProfile"`
	} `json:"properties"`
}

// fetchAksOidcIssuerUrl calls the Azure ARM API to retrieve the OIDC issuer
// URL of the target AKS cluster.  This avoids requiring the user to look up
// and copy the URL into the credentials file manually.
//
// The bootstrap App Registration credentials in creds are used for
// authentication (they are available at RenderInitialCrs time).
func fetchAksOidcIssuerUrl(ctx context.Context, creds *CredsFile) (string, error) {
	cfg := creds.CloudProvider.Config

	token, err := loginAzureARM(ctx, creds)
	if err != nil {
		return "", fmt.Errorf("fetchAksOidcIssuerUrl: %w", err)
	}

	armURL := fmt.Sprintf(
		"https://management.azure.com/subscriptions/%s/resourceGroups/%s/providers/Microsoft.ContainerService/managedClusters/%s?api-version=2024-02-01",
		cfg.SubscriptionId,
		cfg.ResourceGroupName,
		cfg.AksClusterName,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, armURL, nil)
	if err != nil {
		return "", fmt.Errorf("fetchAksOidcIssuerUrl: building request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetchAksOidcIssuerUrl: ARM GET: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("fetchAksOidcIssuerUrl: reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf(
			"fetchAksOidcIssuerUrl: ARM returned HTTP %d for cluster %q: %s",
			resp.StatusCode, cfg.AksClusterName, string(body),
		)
	}

	var aksResp aksARMResponse
	if err := json.Unmarshal(body, &aksResp); err != nil {
		return "", fmt.Errorf("fetchAksOidcIssuerUrl: parsing response: %w", err)
	}

	issuer := aksResp.Properties.OidcIssuerProfile.IssuerURL
	if issuer == "" {
		return "", fmt.Errorf(
			"fetchAksOidcIssuerUrl: OIDC issuer URL is empty for cluster %q — "+
				"ensure OIDC issuer is enabled on the AKS cluster "+
				"(az aks update --enable-oidc-issuer)",
			cfg.AksClusterName,
		)
	}

	log.Printf("✅ AKS OIDC issuer URL for %q: %s", cfg.AksClusterName, issuer)
	return issuer, nil
}

// GetExternalDnsMIClientId reads the external-dns Managed Identity resource ID
// from the TFWorkspace output secret created in the kind cluster, then calls the
// Azure ARM API to resolve the corresponding client_id.
//
// The kind cluster is accessed via the Kind Dagger module (which requires the
// Docker socket used when the cluster was created).
//
// This function is called internally by CmdApplySysServices to obtain the
// dedicated external-dns MI client_id before running the Helm install on the
// target AKS cluster.  It can also be called as a standalone step when the
// client_id is needed for other purposes (e.g. updating state-sys-services).
//
func (m *FirestartrBootstrap) GetExternalDnsMIClientId(
	ctx context.Context,
	// Docker socket needed to access the Kind cluster container.
	dockerSocket *dagger.Socket,
	kindSvc *dagger.Service,
	kindClusterName string,
) (string, error) {

	if !m.isDedicatedDeployment() {
		return "", nil
	}
	return promptForExternalDnsClientID(ctx)
}

func promptForExternalDnsClientID(ctx context.Context) (string, error) {
	if fi, err := os.Stdin.Stat(); err != nil || (fi.Mode()&os.ModeCharDevice) == 0 {
		return "", fmt.Errorf("external-dns Managed Identity client ID must be provided interactively")
	}

	fmt.Print("external-dns Managed Identity client ID: ")
	reader := bufio.NewReader(os.Stdin)
	clientID, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("reading external-dns Managed Identity client ID: %w", err)
	}

	clientID = strings.TrimSpace(clientID)
	if clientID == "" {
		return "", fmt.Errorf("external-dns Managed Identity client ID is required")
	}

	return clientID, nil
}
