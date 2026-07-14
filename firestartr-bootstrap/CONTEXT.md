# firestartr-bootstrap

One-time provisioning pipeline that sets up a GitHub organization for Firestartr
management.

## Language

**Bootstrap**:
The one-time process of creating the repositories, secrets, and configuration
needed for a GitHub organization to be managed by Firestartr.
_Avoid_: setup, initialization, provisioning

**Bootstrap file**:
The `Bootstrapfile.yaml` that declares the target org, customer, environment,
and components for a bootstrap run.
_Avoid_: config file, spec file

**Credentials file**:
The `Credentialsfile.yaml` that holds cloud provider and GitHub App credentials
for a bootstrap run.
_Avoid_: secrets file, auth file

**Component**:
A GitHub repository to be created and configured during bootstrap, declared with
its name, features, variables, and secrets.
_Avoid_: repository, resource
⚠ Not a `ComponentClaim` — this is a bootstrap configuration unit, unrelated to
the Backstage/CRD kind defined in `cdk8s_renderer`.

**Customer**:
The identifier of the Firestartr customer, used to namespace secrets in the
cloud provider parameter store.
_Avoid_: org, tenant, client

**Rollback**:
The inverse bootstrap operation that deletes the repositories, groups, and
webhook created by a bootstrap run.
_Avoid_: teardown, undo, cleanup
