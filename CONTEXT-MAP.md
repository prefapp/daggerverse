# Context Map

This monorepo holds independent Dagger modules that together form the Firestartr
automation platform. Each module is its own bounded context. Shared terms used
across the `firestartr` / `hydrate` module family are defined once here.

## Shared Language

**Firestartr**:
The GitOps platform that manages GitHub organizations, Kubernetes resources, and
Terraform workspaces through a claims-driven model.
_Avoid_: gitops-k8s, the operator, the system

**Claim**:
A deterministic YAML document describing a desired platform resource
(`ComponentClaim`, `GroupClaim`, `TFWorkspaceClaim`, etc.). Authoritatively
defined in `gitops-k8s/packages/fscli`.
_Avoid_: spec, manifest, resource definition

**Feature**:
A named, versioned (or ref-pinned) Mustache template that renders files
(workflows, configs, etc.) into a managed repository. Features are defined in
the `features` repository and referenced inside Claims.
_Avoid_: template, addon, plugin

**Hydration**:
The process of rendering Claims into Kubernetes manifests or Terraform CRs and
committing them to the wet repository.
_Avoid_: rendering, generation, publishing

**Wet repository**:
The Git repository where hydration commits rendered manifests, monitored by
ArgoCD for deployment.
_Avoid_: state repo, state-infra, rendered repo

**Tenant**:
A team or business unit that owns a set of deployed workloads within a cluster
and environment.
_Avoid_: team, customer, client

## Contexts

- [firestartr-bootstrap](./firestartr-bootstrap/CONTEXT.md) — one-time org setup
- [firestartr-config](./firestartr-config/CONTEXT.md) — `.firestartr` configuration model
- [hydrate-orchestrator](./hydrate-orchestrator/CONTEXT.md) — Claims-to-manifests rendering pipeline
- [notify-and-hydrate-state](./notify-and-hydrate-state/CONTEXT.md) — claim PR analysis and state hydration
- [update-claims-features](./update-claims-features/CONTEXT.md) — feature version management in Claims

Modules with no non-obvious domain terms (intentionally skipped):
`dagger-structure-test`, `gh`, `github`, `k6`, `kind`, `opa`, `validate-crds`.

## Relationships

- **firestartr-bootstrap → firestartr-config**: Bootstrap consumes a `.firestartr`
  directory; `firestartr-config` models its structure.
- **update-claims-features → hydrate-orchestrator**: Feature updates either open
  a claims PR (version pin) or trigger hydration directly (ref pin).
- **notify-and-hydrate-state → hydrate-orchestrator**: Analyzes claim PRs and
  triggers hydration on the resulting state changes.
- **hydrate-orchestrator → gh**: Uses `gh` to commit rendered manifests and open
  PRs in the wet repository.
- **firestartr-bootstrap → kind**: Uses `kind` to run a local Kubernetes cluster
  during bootstrap.
- **validate-crds → kind**: Uses `kind` to validate CRD/CR compatibility in a
  local cluster.
