# Research: Azure Bearer Access Token Auth Support

**Question:** Can an Azure Bearer access token (from `az account get-access-token`) be used
directly as authentication in ESO Azure Key Vault SecretStore, ESO PushSecret, and the
Terraform `azurerm` provider?

**Date:** 2026-07-24  
**Sources:** Official ESO docs, ESO Go types (source), azurerm provider source, Terraform docs.

---

## ESO Version in this codebase

`firestartr-bootstrap/operator.go` installs ESO with:

```go
helm upgrade --install external-secrets external-secrets/external-secrets \
  -n external-secrets --create-namespace
```

**No `--version` flag is set.** ESO is always installed at the latest available Helm chart
version. The `v1beta1` API types analysed below are current as of ESO v0.9+.

---

## Scenario 1: ESO SecretStore — Azure Key Vault (read path)

### Sources
- Docs: https://external-secrets.io/latest/provider/azure-key-vault/
- Source: `apis/externalsecrets/v1beta1/secretstore_azurekv_types.go` (ESO main branch)

### Supported `authType` values (enum, exhaustive)

| Value | Description |
|---|---|
| `ServicePrincipal` (default) | Needs `tenantId` + `clientId` + (`clientSecret` OR `clientCertificate`) |
| `ManagedIdentity` | AAD Pod Identity (deprecated upstream). No secret ref needed. |
| `WorkloadIdentity` | OIDC-based federated identity. Uses a Kubernetes `ServiceAccount`. |

### `AzureKVAuth` struct fields (the `authSecretRef` object)

```go
type AzureKVAuth struct {
    ClientID          *smmeta.SecretKeySelector `json:"clientId,omitempty"`
    TenantID          *smmeta.SecretKeySelector `json:"tenantId,omitempty"`
    ClientSecret      *smmeta.SecretKeySelector `json:"clientSecret,omitempty"`
    ClientCertificate *smmeta.SecretKeySelector `json:"clientCertificate,omitempty"`
}
```

### Does ESO support a raw Bearer access token?

**No.** There is no `accessToken` field in `AzureKVAuth` and no `AccessToken` auth type in the
`AzureAuthType` enum. The struct is exhaustive — only `clientSecret` and `clientCertificate` are
accepted alongside `clientId`.

A token from `az account get-access-token` is a short-lived ARM Bearer token (valid ~1 hour,
audience `https://vault.azure.net/`). ESO has no mechanism to inject a pre-obtained Bearer token
directly into Key Vault HTTP requests.

**Workaround:** The closest option without a permanent SP is `WorkloadIdentity`, which uses OIDC
federation and requires cluster-level setup (`azure.workload.identity/client-id` annotation on a
`ServiceAccount`).

---

## Scenario 2: ESO PushSecret — Azure Key Vault (write path)

### Does PushSecret support different auth methods?

**No difference from the read path.** A `PushSecret` references a `SecretStore` (or
`ClusterSecretStore`) object via `spec.secretStoreRefs`. Authentication is entirely delegated to
the store — the `PushSecret` resource itself has no auth fields.

The same `AzureKVProvider` / `AzureKVAuth` types govern both read (`ExternalSecret`) and write
(`PushSecret`) operations. A raw Bearer access token is equally unsupported on the write path.

The only write-specific requirement is an elevated RBAC role:
- For secrets: `Key Vault Secrets Officer` (or Access Policy `Set`/`Delete`)
- For keys: `Key Vault Crypto Officer`
- For certificates: `Key Vault Certificates Officer`

---

## Scenario 3: Terraform `azurerm` provider

### Sources
- Docs (source): `website/docs/index.html.markdown` (hashicorp/terraform-provider-azurerm main)
- Provider schema: `internal/provider/provider.go` (hashicorp/terraform-provider-azurerm main)

### Supported authentication methods (exhaustive from provider schema)

| Method | Key fields / env vars |
|---|---|
| Azure CLI | `use_cli` / `ARM_USE_CLI` (default `true`) |
| Service Principal + Client Secret | `client_secret` / `ARM_CLIENT_SECRET` |
| Service Principal + Client Certificate | `client_certificate` / `ARM_CLIENT_CERTIFICATE` |
| Managed Service Identity | `use_msi` / `ARM_USE_MSI` |
| AKS Workload Identity | `use_aks_workload_identity` / `ARM_USE_AKS_WORKLOAD_IDENTITY` |
| OIDC | `use_oidc` / `ARM_USE_OIDC` + `oidc_token` / `ARM_OIDC_TOKEN` |

### Does `azurerm` support `access_token` or `ARM_ACCESS_TOKEN`?

**No.** Neither `access_token` nor `ARM_ACCESS_TOKEN` appears anywhere in the provider schema or
documentation.

The `oidc_token` / `ARM_OIDC_TOKEN` field looks superficially similar but is fundamentally
different: it accepts an **OIDC ID token** (a JWT issued by an OIDC provider to be exchanged for
an Azure access token), not a pre-obtained ARM Bearer access token from
`az account get-access-token`.

### `use_cli = true` is the implicit path

When `ARM_USE_CLI` is not overridden (default `true`), the provider delegates to the Azure CLI
credential chain, which calls `az account get-access-token` internally. This is the only way a
CLI-obtained token is consumed — indirectly, with the provider managing token refresh. You cannot
supply the raw token string yourself.

---

## Summary table

| Scenario | Bearer token supported? | Config field/env var | Notes |
|---|---|---|---|
| ESO SecretStore (Azure KV, read) | **No** | — | Only `ServicePrincipal`, `ManagedIdentity`, `WorkloadIdentity` auth types exist in the API |
| ESO PushSecret (Azure KV, write) | **No** | — | Delegates to same SecretStore; no separate auth path |
| Terraform `azurerm` provider | **No** | — | No `access_token`/`ARM_ACCESS_TOKEN` field; CLI token used implicitly via `use_cli=true` |

### Token characteristics (for context)

- `az account get-access-token` returns a token valid for **~1 hour**.
- The default audience is `https://management.azure.com/` (ARM). For Key Vault operations you
  need `--resource https://vault.azure.net/` — a different token.
- Neither ESO nor `azurerm` expose a field to inject either token directly.
