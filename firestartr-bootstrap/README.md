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

### 2. Bootstrap File

There are two deployment modes, selected via `deploymentMode`:

| Field | SaaS (default) | Dedicated (Azure) |
|---|---|---|
| `deploymentMode` | `saas` (or omit) | `dedicated` |
| `env` | **required** — `"pre"` or `"pro"` | **omit** — no environment concept |
| `domain` | not used | **required** — base domain (e.g. `"azure-pre.firestartr.dev"`) |

#### 2.1 SaaS Bootstrap File (AWS)

```yaml
# BootstrapFile.yaml (SaaS / AWS)
---
deploymentMode: saas  # optional — saas is the default
org: <github org name> # github org name
customer: <customer name> # customer name used for Firestartr internally
env: <environment> # required for SaaS: "pre" (firestartr-pre) or "pro" (firestartr-pro)
defaultOrgPermissions: <none | view | contribute> # default permissions for the <org>-all group, can be none, view or contribute
defaultBranch: main
defaultBranchStrategy: none
defaultDomainName: <default domain name> # ask customer for a default domain name to be used in the claims, for example "myproduct"
defaultSystemName: <default system name> # ask customer for a default system name to be used in the claims, for example "mysystem"
defaultGroup: <default group> # ask customer for the owner of the default system and domain
defaultFirestartrGroup: firestartr # default group for firestartr users and related components (claims, state-github, state-infra...)

firestartr:
  # Check latest available release at github.com/prefapp/gitops-k8s
  operator: <operator_version> # Ex. v1.56.1
  cli: <cli_version> # Ex. v1.56.1

pushFiles:
  claims:
    push: true # When the process finishes, the generated claims will be pushed to the claims repository.
    repo: "claims" # Normally, the claims repository will be called "claims", but it is possible to change the name.
  dotFirestartr:
    push: true
    repo: ".firestartr"
  crs:
    providers:
      github:
        push: true # When the process finishes, the generated crs will be pushed to the crs repository.
        repo: "state-github" # Normally, the state-github repository will be called "state-github", but it is possible to change the name.
      terraform:
        push: true # When the process finishes, the generated crs will be pushed to the crs repository.
        repo: "state-infra" # Normally, the state-infra repository will be called "state-infra", but it is possible to change the name.
      secrets:
        push: true # When the process finishes, the generated crs will be pushed to the crs repository.
        repo: "state-secrets" # Normally, the state-secrets repository will be called "state-secrets", but it is possible to change the name.
  dotFirestartr:
    push: true # When the process finishes, the generated crs will be pushed to the crs repository.
    repo: ".firestartr" # Normally, the .firestartr repository will be called ".firestartr", but it is possible to change the name.

components:
  - name: "dot-firestartr" # claim name
    description: "Firestartr configuration repository, containing the base configuration for the platform in the organization"
    repoName: ".firestartr" # repository name
    defaultBranch: main
    features: # features that will be provisioned
      - name: firestartr_repo
        version: latest  # Check available versions at github.com/prefapp/features

  - name: "claims"
    description: "Firestartr configuration folders and files"
    defaultBranch: main
    features:
      - name: claims_repo
        version: latest  # You can omit this field to use the latest avaliable version
    secrets:
      - name: FS_IMPORT_PEM_FILE
        value: "ref:secretsclaim:firestartr-secrets:fs-import-pem"
    variables:
      - name: "FS_IMPORT_APP_ID"
        value: "ref:secretsclaim:firestartr-secrets:fs-import-appid"
    labels:
      - automerge

  - name: "catalog"
    description: "Firestartr configuration folders and files"
    defaultBranch: main
    features:
      - name: catalog_repo
        version: latest  # You can also explicitly set "latest" as the version for clarity

  - name: "state-github"
    description: "Firestartr Github wet repository"
    defaultBranch: main
    features:
      - name: state_github
        version: latest  # Check available versions at github.com/prefapp/features
    labels:
      - plan

  - name: "state-infra"
    description: "Firestartr Terraform workspaces wet repository"
    defaultBranch: main
    features:
      - name: state_infra
        version: latest  # Check available versions at github.com/prefapp/features
    labels:
      - plan

  - name: "state-secrets"
    description: "Firestartr Secrets wet repository"
    defaultBranch: main
    features:
      - name: state_secrets
        version: latest  # Check available versions at github.com/prefapp/features
```

#### 2.2 Dedicated Bootstrap File (Azure)

For dedicated (non-SaaS) deployments, replace `env` with `deploymentMode: dedicated` and `domain`:

```yaml
# BootstrapFile.yaml (dedicated / Azure)
---
deploymentMode: dedicated
domain: "azure-pre.firestartr.dev"  # fully-qualified base domain; no env suffix needed
org: <github org name>
customer: <customer name>
# NOTE: 'env' is absent — dedicated deployments have no environment concept
defaultOrgPermissions: <none | view | contribute>
defaultBranch: main
defaultBranchStrategy: none
defaultDomainName: <default domain name>
defaultSystemName: <default system name>
defaultGroup: <default group>
defaultFirestartrGroup: firestartr

firestartr:
  operator: <operator_version>
  cli: <cli_version>

pushFiles:
  claims:
    push: true
    repo: "claims"
  dotFirestartr:
    push: true
    repo: ".firestartr"
  crs:
    providers:
      github:
        push: true
        repo: "state-github"

components:
  # state-sys-services and state-argocd are auto-injected; no need to declare them here
  - name: "dot-firestartr"
    ...
```

All the parameters must be filled. When copy pasting this file, `<placeholders>` must be replaced, but any other values can be treated as defaults and changed if needed:

- `org`: name of the GitHub organization where Firestartr will be installed.
- `deploymentMode`: deployment topology. `saas` (default, can be omitted) targets the shared `firestartr-<env>` org on AWS. `dedicated` targets the customer's own GitHub org on Azure.
- `env`: **SaaS only.** Environment where the deployment and ArgoCD application will be created. Can be either `pre` or `pro`, resulting in commits to the `firestartr-<env>` organization. **Omit for dedicated deployments.**
- `domain`: **Dedicated only.** Fully-qualified base domain for the deployment (e.g. `azure-pre.firestartr.dev`). The webhook URL will be `https://<customer>.events.<domain>`. **Omit for SaaS deployments.**
- `customer`: name used for the org internally, to compose the parameter store paths (e.g. `/firestartr/fs-<customer>-admin/app-id`). Must be set even if it matches the org name.
- `defaultBranch`: default branch name to set in the `defaults` config file, `claims_defaults.yaml`. Usually `main` or `master`.
- `defaultSystemName`: the name of the system that will be created by the bootstrapping process and set in the `claims_defaults.yaml` configuration file. Though any name can be used, it's recommended the bootstrap operator asks the client which system name they want to use as default.
- `defaultDomainName`: the name of the domain that will be created by the bootstrapping process and set in the `claims_defaults.yaml` configuration file. Though any name can be used, it's recommended the bootstrap operator asks the client which domain name they want to use as default.
- `defaultOrgPermissions`: default permissions for the organization members. Can be: `none`, `view` or `contribute`.
- `defaultBranchStrategy`: default branch strategy for the organization repositories. These are defined in the `branch_strategies.yaml` and `expander_branch_strategies.yaml` files. Currently, the bootstrap creates only a definition for `gitflow`, though more can be added after bootstrapping if needed. Allowed values: `none`, `gitflow` or `custom`.
- `defaultFirestartrGroup`: name of the group that will be used by Firestartr by default. It can be an already existing group, which will be imported and used in the bootstrapping process, or a new group that will be created by it.
- `defaultGroup`: slug of the group that will be used by Firestartr by default. It can be an already existing group, which will be imported and used as the owner of the system and domain created by the bootstrap process.
- `firestartr.operator`: Firestartr version to be used by the operator. Must be the name of an image tag, without the flavor (i.e., `v1.53.0` instead of `v1.53.0_full-aws` or `v1.53.0_slim`). You can check the latest available image version [here](https://github.com/prefapp/gitops-k8s/pkgs/container/gitops-k8s).
- `firestartr.cli`: Firestartr CLI version to be used in the importation process. You can check the latest available CLI version [here](https://github.com/prefapp/gitops-k8s/blob/main/.release-please-manifest.json#L2). Note that this CLI version **won't** be the version set as the `FIRESTARTR_CLI_VERSION` organization variable, which is set from the parameter store instead (`/firestartr/<customer-name>/firestartr-cli-version`).
- `pushFiles`: whether or not to push the files create to their respective repositories once the bootstrap process finishes. Each section has two parameters: `push`, which if `true` will push those files to `repo`, whose value should be the name of the repository where those files will be pushed to.
- `components`: list of repositories to create during the bootstrap process. The values of each component will be explained in section 2.3. For a default bootstrap installation, it's recommended to leave them as is and update only the `<feature-version>` placeholders. This section should only be updated on special cases (e.g., the client already has a `claims` repository created). **For dedicated deployments, `state-sys-services` and `state-argocd` are auto-injected and do not need to be listed here.**


#### 2.3 Components

Each component represents a repository that will be created in the organization. All fields are mandatory. The parameters are:

- `name`: name of the repository claim.
- `description`: description of the repository.
- `repoName`: name of the repository. If not specified, it will be the same as `name`.
- `defaultBranch`: default branch name for the repository (usually `main` or `master`).
- `features`: list of features that will be installed in the repository. Each feature must have a `name` and, optionally, a `version`. If `version` is omitted or equals `latest`, the latest avaliable version for the feature will be used. The complete list of available features can be found in the [here](https://github.com/prefapp/features/blob/0e4e2ddac1b9afa83dc207a23d4abe8123e19ade/.release-please-manifest.json) (when setting a feature name from that list, omit the `packages/` prefix, i.e. `name: tech_docs` instead of `name: packages/tech_docs`).
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
    bucket: tfstate-<customer_name>
    region: "eu-west-1"
    # You need to generate a temporary access key and secret key for the AWS credentials, using for example AWS CLI or AWS Console, and provide them in the fields below
    # from the AWS CLI, you can generate the temporary credentials
    # with the command `aws sts get-session-token --duration-seconds 3600`, which will give you credentials valid for 1 hour
    access_key: "<your-access-key>"
    secret_key: "<your-secret-key>"
    token: "<your-session-token>"
  source: hashicorp/aws
  type: aws
  version: ~> 4.0
github:
  prefappBotPat: "<your-prefapp-bot-pat>"  # Prefapp Bot's PAT for the customer, to access prefapp/features repository and use the features within the claims
  operatorPat: "<your-operator-pat>"  # GitOps operator's PAT for the customer, used to push to the git repositories with the generated claims and CRs
```

All the parameters must be filled. When copy pasting this file, `<placeholders>` must be replaced

The rest of the parameters of the `cloudProvider` section are the AWS S3 bucket credentials that will be used as the terraform backend for the `state-infra` repository.

- `github.prefappBotPat`: Personal Access Token for the Prefapp Bot user, used to download the features from the features repository.
- `github.operatorPat`: Personal Access Token for the Operator user, used to commit the deployment and ArgoCD application PRs to the `firestartr-<env>` organization.

#### 3.2 Azure dedicated deployment configuration

For dedicated (non-SaaS) deployments on Azure, set `deploymentMode: dedicated` in your `Bootstrapfile.yaml` and use the following `Credentialsfile.yaml` format:

```yaml
# Credentialsfile.yaml
---
cloudProvider:
  name: azure
  config:
    tenant_id: "<azure-tenant-id>"
    subscription_id: "<azure-subscription-id>"
    # Runtime identity: firestartr-mi User-Assigned Managed Identity.
    # Used in the deployed AKS cluster via Workload Identity (no client secret).
    client_id: "<firestartr-mi-client-id>"
    # Bootstrap identity: dedicated App Registration (Service Principal).
    # Used only during bootstrap in the local kind cluster by ESO and Terraform.
    # Delete the entire App Registration after bootstrap completes.
    bootstrap_client_id: "<bootstrap-sp-application-id>"
    bootstrap_client_secret: "<bootstrap-sp-client-secret>"
    storage_account_name: "tfstate<customer>"   # Azure Storage Account name for Terraform state
    container_name: "tfstate"
    resource_group_name: "rg-firestartr"        # Also used for DNS zone resource group
    key_vault_name: "firestartr-kv"             # Azure Key Vault name (no slashes in secret names)
    aks_cluster_name: "<aks-cluster-name>"      # Name of the target AKS cluster
    location: "<azure-region>"                  # Azure region, e.g. "westeurope" — must match resource group
  source: hashicorp/azurerm
  type: azurerm
  version: "~> 3.0"
github:
  prefappBotPat: "<prefapp-bot-pat>"
  operatorPat: "<operator-pat>"
```

And your `Bootstrapfile.yaml` must include `deploymentMode` and `domain`:

```yaml
# Bootstrapfile.yaml (dedicated Azure example)
---
deploymentMode: dedicated
domain: "azure-pre.firestartr.dev"   # Fully-qualified base domain; encodes env if needed
org: "<your-github-org>"
customer: "<customer-name>"
# (no 'env' field for dedicated deployments)
...
```

**Pre-flight requirements for dedicated Azure deployments:**

1. A `firestartr-mi` User-Assigned Managed Identity with OIDC federation configured on the AKS cluster.
2. A bootstrap App Registration (Service Principal) in Entra ID with:
   - `Key Vault Secrets Officer` on the Azure Key Vault
   - `Storage Blob Data Contributor` on the Terraform state storage account
   - A client secret (populate `bootstrap_client_id` and `bootstrap_client_secret` in the credentials file)
   - **Delete the entire App Registration after bootstrap completes.**
3. Azure Key Vault populated with all required secrets using the `[a-zA-Z0-9-]` naming convention:
   - `fs-pem`, `fs-app-id`, `fs-<org>-installation-id`
   - `fs-admin-pem`, `fs-admin-app-id`, `fs-admin-<org>-installation-id`
   - `fs-argocd-pem`, `fs-argocd-app-id`, `fs-argocd-<org>-installation-id`
   - `fs-state-pem`, `fs-state-app-id`, `fs-state-<org>-installation-id`
   - `fs-checks-pem`, `fs-checks-app-id`, `fs-checks-<org>-installation-id`
   - `fs-import-pem`, `fs-import-app-id`, `fs-import-<org>-installation-id`
   - `github-webhook-secret`, `prefapp-bot-pat`, `firestartr-cli-version`
4. Delegated DNS zone matching `domain` in the Azure resource group.

**What the dedicated bootstrap does differently from SaaS:**

| Concern | SaaS (AWS) | Dedicated (Azure) |
|---|---|---|
| Repo targeting | `firestartr-<env>/<repo>` | `<bootstrap.Org>/<repo>` |
| Cluster services | ESO only (pre-provisioned) | Installs ESO + nginx + cert-manager + external-dns + ArgoCD |
| Webhook URL | `<customer>.events[.<env>].firestartr.dev` | `<customer>.events.<domain>` |
| Secret refs | AWS Parameter Store paths | Azure Key Vault dash-delimited names |
| Deployment values | `values.tmpl` (AWS/ALB/IAM) | `azure_values.tmpl` (Azure/nginx/MI) |

> **Important:** The `bootstrap_client_id` and `bootstrap_client_secret` belong to a dedicated App Registration created solely for this bootstrap run. The deployed state uses `firestartr-mi` via Workload Identity and never needs a client secret. **Delete the entire App Registration after bootstrap completes** — not just the secret.

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

Also, the following AWS Parameter Store are going to be generated and/or uploaded (pushed):

- `/firestartr/<customer>/prefapp-bot-pat`: Personal Access Token for the Prefapp Bot user
- `/firestartr/<customer>/firestartr-cli-version`: Version of the Firestartr CLI to set as the default in the organization
- `/firestartr/<customer>/github-webhook/secret`: Secret for the GitHub Webhook


Rollback command:

```shell
dagger --bootstrap-file="./Bootstrapfile.yaml" \
       --credentials-secret="file:./Credentialsfile.yaml" \
       call cmd-rollback \
       --kubeconfig="${HOME}/.kube" \
       --kind-svc=tcp://localhost:<your-kind-port>
```

This will rollback the changes done by the bootstrap process. It will:

- Delete the repositories created, along with their features and secrets
- Delete the groups created by the bootstrap process (not any that where imported)
- Delete the GitHub org's webhook created by the bootstrap process

⚠️  WARNING: This process will only delete the resources mentioned above. Any other resources created by the process, such as the deployment and ArgoCD applications PRs, the files created by merging them, or the Terraform state stored in the S3 bucket will not be deleted. Please make sure to manually delete those resources if needed.

Note that the rollback process may fail to delete a resource if it is in an error state. In that case, you will need to manually delete the resource. The process will output all changes done and failed deletions when it's finished.

All of these commands can be run separately, as described in step 6.

## 5. Step by step script

It is provided a ```step_by_step.sh``` script to help with the bootstrap process.

This script executes the provisioning pipeline in sequential stages, ensuring that prerequisite tasks are completed before moving to deployment steps.
At critical junctures, the pipeline will pause and require explicit user input to determine the next action, especially if a previous step failed.

🛑 User Intervention Required
When the script encounters a non-fatal error during a setup stage (e.g., resources already exist or a validation warning),
it will halt and prompt you to decide how to proceed.

```sh
# copy the script
cp step_by_step.sh my-steps.sh

# edit the file and fill the needed data at the top
vim my-steps.sh

# run the script
bash my-steps.sh

```

### 5.1 Step by step script flags

- `-k` or `--kind-cluster-name`: the name of the kind cluster to use. If not provided, the script will prompt the user to create a new kind cluster, with an automatically generated cluster name.
- `-d` or `--delete-cluster-on-failure`: if set, the kind cluster will be deleted if any step fails.
- `--auto-execute-script`: if set, the script will not prompt the user for input and will automatically proceed to the next step. This option is not recommended unless the script is being executed in a testing environment.
- `-w` or `--wait-time`: time in seconds to wait between steps when `--auto-execute-script` is set. Default is 5 seconds.
- `-e` or `--extract-claim`: the name of a claim to extract from the cache volume after rendering. The claim YAML and any associated custom resources are exported to `./boot/extracted/<claim-name>`. Requires a valid `VOLUME_ID` set in the script header.
- `-h` or `--help`: display a help message showing all the previous commands.

## 6. Individual commands

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
This will initialize the secrets machinery in the kind cluster, installing Helm and creating the secrets necesary for the bootstrap process. The following AWS Parameter Store parameters will also be generated and/or uploaded (pushed):
- `/firestartr/<customer>/prefapp-bot-pat`: Personal Access Token for the Prefapp Bot user
- `/firestartr/<customer>/firestartr-cli-version`: Version of the Firestartr CLI to set as the default in the organization
- `/firestartr/<customer>/github-webhook/secret`: Secret for the GitHub Webhook

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

Import existing org resources (groups, users) and create the ones needed by Firestartr (groups, repositories and webhooks). These are saved to `--cache-volume` for later use.

Push created resources:

```shell
dagger --bootstrap-file="./Bootstrapfile.yaml" \
      --credentials-secret="file:./Credentialsfile.yaml" \
      call cmd-push-resources \
      --kubeconfig="${HOME}/.kube" \
      --kind-svc=tcp://localhost:<your-kind-port> \
      --cache-volume=<your-volume-id>
```

Push the created claims and crs files stored in `--cache-volume` to their respective repositories.

Create deployment PR:

```shell
dagger --bootstrap-file="./Bootstrapfile.yaml" \
       --credentials-secret="file:./Credentialsfile.yaml" \
       call cmd-push-deployment
```

- **SaaS:** Creates a deployment PR in `firestartr-<env>/app-firestartr`.
- **Dedicated:** Creates a deployment PR in `<org>/state-sys-services` with all sys-service release descriptors and values.

Apply sys-services with values (**dedicated only**):

```shell
dagger --bootstrap-file="./Bootstrapfile.yaml" \
       --credentials-secret="file:./Credentialsfile.yaml" \
       call cmd-apply-sys-services \
       --docker-socket=/var/run/docker.sock \
       --kind-svc=tcp://localhost:<your-kind-port> \
       --kind-cluster-name=<your-kind-cluster-name>
```

This step is **only needed for dedicated deployments**. It renders the Azure-specific Helm values and applies them directly to the AKS cluster via `helm upgrade --install --values`, ensuring all cluster services (nginx, cert-manager, external-dns, ArgoCD, argo-events, argo-workflows) are correctly configured from the start.

The AKS cluster name is read from `aks_cluster_name` in the credentials file. If the external-dns Managed Identity client ID cannot be resolved automatically, the command prompts for the client ID on the terminal.

> Merge the `state-sys-services` PR **before** running this step so the desired state is recorded in git first.

Create ArgoCD application PR:

```shell
dagger --bootstrap-file="./Bootstrapfile.yaml" \
       --credentials-secret="file:./Credentialsfile.yaml" \
       call cmd-push-argo
```

- **SaaS:** Creates an application PR in `firestartr-<env>/state-argocd`.
- **Dedicated:** Creates an application PR in `<org>/state-argocd` and patches `<org>/state-sys-services` with the ArgoCD secrets entry.

## 7. Troubleshooting

The `kubectl apply` commands have a timeout of 10 hours. This is done to allow time for debugging and fixing the issue. If you see any `kubectl apply` command executing for two minutes or more without finishing, this probably indicates that an error happened. Without stopping Dagger, note the CR that's being applied and enter the cluster via `k9s`. Then:

- Check the CR actually has an error status. Maybe the process is taking a while to finish.
- If the CR has an error status, the actual error message won't appear in it (instead a generic error message will be shown). To see the actual error message, go to the `tfresults` CRD and select the record whose name matches the one in the error status of the CR you're debugging.
- You can also check the logs of the pod running the controller. It will be in the namespace `default` and its name will be similar to `firestartr-init-firestartr-init-<random-string>`.

## 8. Updating an already existing repository

In some cases, we might want to use this module to update an already bootstrapped repository: when we want to update a claims repository with the `claims_repo` feature from `1.X.Y` to `2.X.Y`, as version `2.X.Y` requires creating the `state-repo` repository and adding it to the ArgoCD ecosystem. In those cases, this module can be used to simplify that process, though it also can be done manually. The steps to do it are:

- Add the following to the Bootstrapfile.yaml configuration file:

```
createWebhook: false

pushFiles:
  providers:
    ....
    secrets:
      push: true  # Set all other <provider>.push values to false
      repo: "state-secrets"

components:
  - name: "state-secrets"
    description: "Firestartr secrets wet repository"
    defaultBranch: main
    features:
      - name: state_secrets
```

- Execute the Bootstrap process:
1. The first five steps will create the new repository and push the ExternalSecrets.firestartr-secrets.yaml file to it.
2. Skip the Push organization state secrets step, as this creates the secrets in the GitHub org and that was already done by the first Bootstrap process.
3. The two ArgoCD steps should create new PRs in app-firestartr and state-argocd. state-sys-services will be skipped as no changes are necessary and a warning will be shown on screen, but this is normal and can be safely ignored.
- Give fs-<org>-argocd and fs-<org>-state access to the newly created state-secrets
- The ExternalSecrets.firestartr-secrets.yaml file stored in state-infra must be manually deleted. Since a new one already exists in state-secrets and these files have no tfStateKey nothing should happen

[More info](https://github.com/prefapp/gitops-k8s/issues/1999#issuecomment-4519670990)

## 9. Flow chart
![BootstrapDiagram drawio](https://github.com/user-attachments/assets/1c824119-b147-47bb-b8f8-8cc17db29c6a)
