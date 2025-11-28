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

- `/firestartr/<customer>/fs-<customer>/pem`
- `/firestartr/<customer>/fs-<customer>/app-id`
- `/firestartr/<customer>/fs-<customer>/<org>/installation-id`
- `/firestartr/<customer>/fs-<customer>-admin/pem`
- `/firestartr/<customer>/fs-<customer>-admin/app-id`
- `/firestartr/<customer>/fs-<customer>-admin/<org>/installation-id`
- `/firestartr/<customer>/fs-<customer>-checks/pem`
- `/firestartr/<customer>/fs-<customer>-checks/app-id`
- `/firestartr/<customer>/fs-<customer>-checks/<org>/installation-id`
- `/firestartr/<customer>/fs-<customer>-state/pem`
- `/firestartr/<customer>/fs-<customer>-state/app-id`
- `/firestartr/<customer>/fs-<customer>-state/<org>/installation-id`
- `/firestartr/<customer>/fs-<customer>-import/pem`
- `/firestartr/<customer>/fs-<customer>-import/app-id`
- `/firestartr/<customer>/fs-<customer>-import/<org>/installation-id`
- `/firestartr/<customer>/fs-<customer>-argocd/pem`
- `/firestartr/<customer>/fs-<customer>-argocd/app-id`
- `/firestartr/<customer>/fs-<customer>-argocd/<org>/installation-id`
- `/firestartr/<customer>/prefapp-bot-pat`: Personal Access Token for the Prefapp Bot user
- `/firestartr/<customer>/firestartr-cli-version`: Version of the Firestartr CLI to set as the default in the organization
- `/firestartr/<customer-name>/github-webhook/secret`: Secret for the GitHub Webhook

### 2. Bootstrap File

```yaml
# BootstrapFile.yaml
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

All the parameters must be filled. When copy pasting this file, `<placeholders>` must be replaced, but any other values can be treated as defaults and changed if needed:

- `org`: name of the GitHub organization where Firestartr will be installed.
- `env`: environment where the deployment and ArgoCD application will be created. Can be either `pre` or `pro`, and will result in commits being created to the necessary repositories in the `firestartr-<env>` organization.
- `customer`: name used for the org internally, to compose the parameter store paths (e.g. `/firestartr/fs-<customer>-admin/app-id`). Must be set even if it matches the org name.
- `defaultBranch`: default branch name to set in the `defaults` config file, `claims_defaults.yaml`. Usually `main` or `master`.
- `defaultSystemName`: the name of the system that will be created by the bootstrapping process and set in the `claims_defaults.yaml` configuration file. Though any name can be used, it's recommended the bootstrap operator asks the client which system name they want to use as default.
- `defaultDomainName`: the name of the domain that will be created by the bootstrapping process and set in the `claims_defaults.yaml` configuration file. Though any name can be used, it's recommended the bootstrap operator asks the client which domain name they want to use as default.
- `defaultOrgPermissions`: default permissions for the organization members. Can be: `none`, `view` or `contribute`.
- `defaultBranchStrategy`: default branch strategy for the organization repositories. These are defined in the `branch_strategies.yaml` and `expander_branch_strategies.yaml` files. Currently, the bootstrap creates only a definition for `gitflow`, though more can be added after bootstrapping if needed. Allowed values: `none`, `gitflow` or `custom`.
- `defaultFirestartrGroup`: name of the group that will be used by Firestartr by default. It can be an already existing group, which will be imported and used in the bootstrapping process, or a new group that will be created by it.
- `firestartr.operator`: Firestartr version to be used by the operator. Must be the name of an image tag, without the flavor (i.e., `v1.53.0` instead of `v1.53.0_full-aws` or `v1.53.0_slim`). You can check the latest available image version [here](https://github.com/prefapp/gitops-k8s/pkgs/container/gitops-k8s).
- `firestartr.cli`: Firestartr CLI version to be used in the importation process. You can check the latest available CLI version [here](https://github.com/prefapp/gitops-k8s/blob/main/.release-please-manifest.json#L2). Note that this CLI version **won't** be the version set as the `FIRESTARTR_CLI_VERSION` organization variable, which is set from the parameter store instead (`/firestartr/<customer-name>/firestartr-cli-version`).
- `pushFiles`: whether or not to push the files create to their respective repositories once the bootstrap process finishes. Each section has two parameters: `push`, which if `true` will push those files to `repo`, whose value should be the name of the repository where those files will be pushed to.
- `components`: list of repositories to create during the bootstrap process. The values of each component will be explained in section 2.1. For a default bootstrap installation, it's recommended to leave them as is and update only the `<feature-version>` placeholders. This section should only be updated on special cases (e.g., the client already has a `claims` repository created).


#### 2.1 Components

Each component represents a repository that will be created in the organization. All fields are mandatory. The parameters are:

- `name`: name of the repository claim.
- `description`: description of the repository.
- `repoName`: name of the repository. If not specified, it will be the same as `name`.
- `defaultBranch`: default branch name for the repository (usually `main` or `master`).
- `features`: list of features that will be installed in the repository. Each feature must have a `name` and a `version`.The complete list of available features can be found in the [here](https://github.com/prefapp/features/blob/0e4e2ddac1b9afa83dc207a23d4abe8123e19ade/.release-please-manifest.json) (when setting a feature name from that list, omit the `packages/` prefix, i.e. `name: tech_docs` instead of `name: packages/tech_docs`).
- `secrets`: (optional) list of secrets that will be created in the repository. Each secret must have a `name` and a `value`. `name` will be the name of the secret in the repository, and `value` should be a reference to a secret in the [`SecretsClaim`](https://github.com/prefapp/daggerverse/blob/main/firestartr-bootstrap/templates/initial_claims.tmpl#L60-L77) (the link provided goes to the `main` branch version of the template file. Please select the appropriate version if needed). The format for referencing a secret from that file is: `ref:secretsclaim:firestartr-secrets:<secretName>`
- `variables`: (optional) list of variables that will be created in the repository. Each variable must have a `name` and a `value`. `name` will be the name of the variable in the repository, and `value` should be a reference to a secret in the [`SecretsClaim`](https://github.com/prefapp/daggerverse/blob/main/firestartr-bootstrap/templates/initial_claims.tmpl#L60-L77) (the link provided goes to the `main` branch version of the template file. Please select the appropriate version if needed). The format for referencing a secret from that file is: `ref:secretsclaim:firestartr-secrets:<secretName>`
- `labels`: (optional) list of labels that will be created in the repository. In this case, used to create the `plan` label needed for the workflows of the `state_infra` feature to work.


### 3. Credentials File

#### 3.1 AWS terraform backend provider configuration

```yaml
# Credentialsfile.yaml
---
cloudProvider:
  name: aws
  config:
    bucket: <your-bucket-name>
    region: "eu-west-1"
    access_key: "<your-access-key>"
    secret_key: "<your-secret-key>"
    token: "<your-token>"
  source: hashicorp/aws
  type: aws
  version: ~> 4.0
argoCD:
  githubAppId: "<id of the github app for argocd>"
  githubAppInstallationId: "<id of the github app for argocd installed on the org>"
githubApp:
  owner: <org>
  botName: "fs-<org>[bot]"
  prefappBotPat: "<bot-pat>"  # Prefapp Bot's PAT
  operatorPat: "<operator-pat>"  # Operator's PAT, used to commit to the org firestartr-<env>
```

All the parameters must be filled. When copy pasting this file, `<placeholders>` must be replaced

The rest of the parameters of the `cloudProvider` section are the AWS S3 bucket credentials that will be used as the terraform backend for the `state-infra` repository.

- `githubApp.owner`: name of the GitHub organization where Firestartr will be installed.
- `githubApp.botName`: name of the GitHub App bot user.
- `githubApp.prefappBotPat`: Personal Access Token for the Prefapp Bot user, used to download the features from the features repository.
- `githubApp.operatorPat`: Personal Access Token for the Operator user, used to commit the deployment and ArgoCD application PRs to the `firestartr-<env>` organization.

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
