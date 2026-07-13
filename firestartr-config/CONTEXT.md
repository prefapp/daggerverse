# firestartr-config

Models the `.firestartr` configuration directory that drives Firestartr deployments.

## Language

**Platform**:
A named deployment target (e.g. a Kubernetes cluster) with a type, a set of
tenants and environments, and constraints on which Claim resource types and
providers are allowed.
_Avoid_: cluster, environment, target

**App**:
A Firestartr-managed application grouping services for a given tenant and
environment.
_Avoid_: workload, service, deployment

**Registry**:
A container image registry configuration used by Firestartr-managed deployments.
_Avoid_: image repository, image source

**dotFirestartr directory**:
The `.firestartr/` directory containing the platform, app, registry, and
repository YAML configurations consumed by this module.
_Avoid_: config dir, values dir
