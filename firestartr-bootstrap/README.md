# Firestartr Bootstrap

## Overview

The firestartr bootstrap is a dagger workflow that can provision the initial repositories, files and org configurations to start Firestartr in a github organization.

## How to launch the bootstrap

### 1. Requirements

#### 1.1 Local machine requirements

You'll need to install on your local machine:
- [**go**](https://go.dev/doc/install) (v1.22+)
- [**docker**](https://docs.docker.com/engine/install/) (v24+)
- [**dagger**](https://docs.dagger.io/install) (v0.18.5+)
- [**kind**](https://kind.sigs.k8s.io/docs/user/quick-start#installation) (v0.24.0+).

#### 1.2 AWS requirements

The following AWS Parameter Store parameters are required:
- `/firestartr/<org-name>/fs-<org-name>-admin/pem`
- `/firestartr/<org-name>/fs-<org-name>-admin/app-id`
- `/firestartr/<org-name>/fs-<org-name>-admin/installation-id`
- `/firestartr/<org-name>/prefapp-bot-pat`

### 2. Bootstrap File

```yaml
# Bootstrapfile.yaml
---
org: <github-org>
defaultBranch: main
defaultSystemName: default-system
defaultDomainName: default-domain
defaultOrgPermissions: view
defaultBranchStrategy: none
defaultFirestartrGroup: firestartr
finalSecretStoreName: <secret-store-name>
webhookUrl: <webhook-url>

firestartr:
  version: <cli-version> # Check latest available at github.com/prefapp/gitops-k8s. Omit the "v" (i.e. use "1.2.3" instead of "v1.2.3")
pushFiles:
  claims:
    push: true # When the process finishes, the generated claims will be pushed to the claims repository.
    repo: "claims" # Normally, the claims repository will be called "claims", but it is possible to change the name.
  crs:
    providers:
      github:
        push: true # When the process finishes, the generated crs will be pushed to the crs repository.
        repo: "state-github" # Normally, the state-github repository will be called "state-github", but it is possible to change the name.
      terraform:
        push: true # When the process finishes, the generated crs will be pushed to the crs repository.
        repo: "state-infra" # Normally, the state-infra repository will be called "state-infra", but it is possible to change the name.

components:
  - name: "dot-firestartr" # claim name
    description: "Repository with the terraform code for manage the multi-tenant infrastructure"
    repoName: ".firestartr" # repository name
    defaultBranch: main
    features: # features that will be provisioned
      - name: firestartr_repo
        version: <feature-version>  # Check latest available at github.com/prefapp/features

  - name: "claims"
    description: "Firestartr configuration folders and files"
    defaultBranch: main
    features:
      - name: claims_repo
        version: <feature-version>  # Check latest available at github.com/prefapp/features
    secrets:
      - name: FS_IMPORT_PEM_FILE
        value: "ref:secretsclaim:firestartr-secrets:fs-import-pem"
    variables:
      - name: "FS_IMPORT_APP_ID"
        value: "ref:secretsclaim:firestartr-secrets:fs-import-appid"

  - name: "catalog"
    description: "Firestartr configuration folders and files"
    defaultBranch: main
    features:
      - name: catalog_repo
        version: <feature-version>  # Check latest available at github.com/prefapp/features

  - name: "state-github"
    description: "Firestartr Github wet repository"
    defaultBranch: main
    features:
      - name: state_github
        version: <feature-version>  # Check latest available at github.com/prefapp/features

  - name: "state-infra"
    description: "Firestartr Terraform workspaces wet repository"
    defaultBranch: main
    features:
      - name: state_infra
        version: <feature-version>  # Check latest available at github.com/prefapp/features
    labels:
      - plan
```

### 3. Credentials File

#### 3.1 AWS terraform backend provider configuration

```yaml
# Credentialsfile.yaml
---
cloudProvider:
  providerConfigName: backend-provider-config-name
  name: aws
  config:
    bucket: "my-bucket"
    region: "eu-west-1"
    access_key: "AKIAXXXXXXXXXXXXXXXX"
    secret_key: "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
    token: "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
  source: hashicorp/aws
  type: aws
  version: ~> 4.0
githubApp:
  providerConfigName: github-app-provider-config-name
  owner: <org>
  botName: "fs-<org>[bot]"
```

#### 3.2 Azure terraform backend provider configuration (currently not supported)

```yaml
# Credentialsfile.yaml
---
cloudProvider:
  providerConfigName: backend-provider-config-name
  name: azurerm
  config:
    use_azuread_auth: true
    tenant_id: "00000000-0000-0000-0000-000000000000"
    client_id: "00000000-0000-0000-0000-000000000000"
    client_secret: "************************************"
    storage_account_name: "abcd1234"
    container_name: "tfstate"
  source: hashicorp/aws
  type: aws
  version: ~> 4.0
githubApp:
  providerConfigName: github-app-provider-config-name
  owner: "firestartr-test"
  botName: "firestartr-local-development-app[bot]"
```

### 4. How to launch the bootstrap

```shell
dagger --bootstrap-file="./Bootstrapfile.yaml" \
       --credentials-secret="file:./Credentialsfile.yaml" \
       call run-bootstrap \
       --docker-socket=/var/run/docker.sock \
       --kind-svc=tcp://127.0.0.1:3000
```

## 5. Flow chart
![BootstrapDiagram drawio](https://github.com/user-attachments/assets/1c824119-b147-47bb-b8f8-8cc17db29c6a)
