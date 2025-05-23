# Hydrate Orchestrator

## Overview

The **Pull Model Rendering System** streamlines the deployment of Kubernetes objects, Terraform resources, and secrets using a unified GitOps pipeline. It leverages Golang modules, Hydrate for orchestration, ArgoCD for deployment, External Secrets for secret management, and the Firestartr operator for Terraform.

![render-dagger-module-Dependencies diagram drawio](https://github.com/user-attachments/assets/3dbb698f-0ebe-4fc7-9471-1d3a98bf1dc1)

#### Core Components
1. **Hydrate Orchestrator**: Manages the workflow, coordinating three rendering modules:
   - **HydateKubenetes**: Renders Kubernetes core objects (e.g., ConfigMaps, Services) and workloads (e.g., Deployments) into a wet repository.
   - **HydateTerraformWorkspaces**: Renders Terraform Workspaces as Kubernetes Custom Resources (CRs) into the wet repository.
   - **HydrateSecrets**: Generates Kubernetes Secrets (e.g., database passwords) using External Secrets, linking them to Terraform Workspaces.

2. **ArgoCD**: Monitors the wet repository in GitHub, deploying Kubernetes objects to clusters and applying Terraform CRs.

3. **Firestartr Operator**: Processes Terraform CRs, resolves secrets (e.g., a PostgreSQL password), and applies Terraform to provision non-Kubernetes resources.

4. **GitHub**: Stores the wet repository with all rendered resources and manages the hydrating system opening pull requests.

#### Workflow
1. **Rendering**:
   - Kubernetes objects, Terraform CRs, and secrets are rendered by their respective modules.
   - Outputs are stored in a wet repository in GitHub.
2. **Deployment**:
   - ArgoCD deploys Kubernetes objects to clusters and applies Terraform CRs.
   - Firestartr operator handles Terraform CRs, using secrets to provision resources (e.g., a PostgreSQL database).
3. **Result**: A unified pipeline for Kubernetes and Terraform deployments with secure secret management.

#### Flow chart
![render-dagger-module-Flow chart drawio](https://github.com/user-attachments/assets/184d5660-2cf5-472f-a66b-e0190f8d2abc)

[diagrams](https://app.diagrams.net/?src=about#G1KiCY57g_N5B6txxDpM-eVw5R-vl9NQCr#%7B%22pageId%22%3A%22BWRnRSUX-oT4-0keMzf5%22%7D)
