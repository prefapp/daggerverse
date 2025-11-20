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

Create a kind cluster using the following command:

```shell
kind create cluster
```

Using `docker ps`, note the port that kind is using to expose the Kubernetes API server. You will need it later when launching the bootstrap.

#### 1.2 AWS requirements

The following AWS Parameter Store parameters are required:

- `/firestartr/<org-name>/fs-<org-name>-admin/pem`
- `/firestartr/<org-name>/fs-<org-name>-admin/app-id`
- `/firestartr/<org-name>/fs-<org-name>-admin/installation-id`
- `/firestartr/<org-name>/fs-<org-name>-checks/pem`
- `/firestartr/<org-name>/fs-<org-name>-checks/app-id`
- `/firestartr/<org-name>/fs-<org-name>-state/pem`
- `/firestartr/<org-name>/fs-<org-name>-state/app-id`
- `/firestartr/<org-name>/fs-<org-name>-import/pem`
- `/firestartr/<org-name>/fs-<org-name>-import/app-id`
- `/firestartr/<org-name>/prefapp-bot-pat`: Personal Access Token for the Prefapp Bot user
- `/firestartr/<org-name>/firestartr-cli-version`: Version of the Firestartr CLI to set as the default in the organization
- `/firestartr/<org-name>/github-webhook/secret`: Secret for the GitHub Webhook

### 2. Bootstrap File

```yaml
# Bootstrapfile.yaml
---
org: <github-org>
env: <env>  # either "pre" or "pro"
customer: <customer-name>  # name used for the org internally, within the parameter store
defaultBranch: main
defaultSystemName: default-system
defaultDomainName: default-domain
defaultOrgPermissions: view
defaultBranchStrategy: none
defaultFirestartrGroup: firestartr

firestartr:
  # Check latest available release at github.com/prefapp/gitops-k8s
  operator: <operator-version>
  cli: <cli-version>
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
  providerConfigName: <your-backend-provider-config-name>
  name: aws
  config:
    bucket: <your-bucket-name>
    region: "eu-west-1"
    access_key: "AKIAXXXXXXXXXXXXXXXX"
    secret_key: "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
    token: "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
  source: hashicorp/aws
  type: aws
  version: ~> 4.0
githubApp:
  providerConfigName: <your-github-app-provider-config-name>
  owner: <org>
  botName: "fs-<org>[bot]"
  prefappBotPat: "<bot-pat>"  # Prefapp Bot's PAT
  operatorPat: "<operator-pat>"  # Operator's PAT, used to commit to the org firestartr-<env>
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

`<your-kind-port>`: Replace with the port that kind is using to expose the Kubernetes API server (noted in step 1.1).

Main command:

```shell
dagger --bootstrap-file="./Bootstrapfile.yaml" \
       --credentials-secret="file:./Credentialsfile.yaml" \
       call cmd-run-bootstrap \
       --kubeconfig="${HOME}/.kube" \
       --kind-svc=tcp://localhost:<your-kind-port>
```

This will launch the whole bootstrapping process. It will:

- Validate your configuration files
- Populate your kind cluster with the needed resources
- Import the org's existing groups and users
- Create the repositories specified in `Bootstrapfile.yaml`
- Upload the claims and crs files created to their respective repositories
- Create a deployment PR in `firestartr-<env>/app-firestartr`
- Create an application PR in `firestartr-<env>/state-argocd`

All of these commands can be run separately, as described in step 5.


Rollback command:

```shell
dagger --bootstrap-file="./Bootstrapfile.yaml" \
       --credentials-secret="file:./Credentialsfile.yaml" \
       call cmd-run-bootstrap \
       --kubeconfig="${HOME}/.kube" \
       --kind-svc=tcp://localhost:<your-kind-port>
```

This will rollback the changes done by the bootstrap process. It will:

- Delete the repositories created, along with their features and secrets
- Delete the groups created by the bootstrap process (not any that where imported)
- Delete the GitHub org's webhook created by the bootstrap process

Note that the rollback process may fail to delete a resource if it is in an error state. In that case, you will need to manually delete the resource. The process will output all changes done and failed deletions when it's finished.

## 5. Individual commands

You can run the individual commands that compose the bootstrap process separately. This is useful for debugging or if you want to run only a part of the process. They are:

Create persistent volume:

```shell
dagger --bootstrap-file="./Bootstrapfile.yaml" \
       --credentials-secret="file:./Credentialsfile.yaml" \
       call cmd-create-persistent-volume \
       --volume-name "firestartr-init"
```

This will create a persistent volume in dagger that will be used to cache resources between commands. Note the volume ID returned, as it will be needed in the commands that need it (it will be marked as `<your-volume-id>`, and will be the SHA outputed by this command).


Validate bootstrap configuration:

```shell
dagger --bootstrap-file="./Bootstrapfile.yaml" \
       --credentials-secret="file:./Credentialsfile.yaml" \
       call cmd-validate-bootstrap
```

This will validate the bootstrap configuration files.


Initialize secrets machinery:

```shell
dagger --bootstrap-file="./Bootstrapfile.yaml" \
       --credentials-secret="file:./Credentialsfile.yaml" \
       call cmd-init-secrets-machinery \
       --kubeconfig="${HOME}/.kube" \
       --kind-svc=tcp://localhost:<your-kind-port>
```
This will initialize the secrets machinery in the kind cluster, installing Helm and creating the secrets necesary for the bootstrap process.

Initialize GitHub Apps machinery:

```shell
dagger --bootstrap-file="./Bootstrapfile.yaml" \
      --credentials-secret="file:./Credentialsfile.yaml" \
      call cmd-init-github-apps-machinery \
      --kubeconfig="${HOME}/.kube" \
      --kind-svc=tcp://localhost:<your-kind-port>
```

This will initialize the GitHub Apps machinery in the kind cluster, populating the variables needed for the bootstrap process to work, as well as check the org's plan and if the `<org>-all` group already exists.

Import and create resources:

```shell
dagger --bootstrap-file="./Bootstrapfile.yaml" \
       --credentials-secret="file:./Credentialsfile.yaml" \
       call cmd-import-resources \
       --kubeconfig="${HOME}/.kube" \
       --kind-svc=tcp://localhost:<your-kind-port> \
       --cache-volume=<your-volume-id>
```

Import existing org resources (groups, users) and create the ones needed by Firestartr (groups, repositories and webhooks).

Push created resources:

```shell
dagger --bootstrap-file="./Bootstrapfile.yaml" \
      --credentials-secret="file:./Credentialsfile.yaml" \
      call cmd-push-resources \
      --kubeconfig="${HOME}/.kube" \
      --kind-svc=tcp://localhost:<your-kind-port> \
      --cache-volume=<your-volume-id>
```

Push the created claims and crs files to their respective repositories.

Create deployment PR:

```shell
dagger --bootstrap-file="./Bootstrapfile.yaml" \
       --credentials-secret="file:./Credentialsfile.yaml" \
       call cmd-push-deployment
```

Creates a deployment PR in `firestartr-<env>/app-firestartr`.


Create ArgoCD application PR:

```shell
dagger --bootstrap-file="./Bootstrapfile.yaml" \
       --credentials-secret="file:./Credentialsfile.yaml" \
       call cmd-push-argo
```

Creates an application PR in `firestartr-<env>/state-argocd`.


## 6. Flow chart
![BootstrapDiagram drawio](https://github.com/user-attachments/assets/1c824119-b147-47bb-b8f8-8cc17db29c6a)
