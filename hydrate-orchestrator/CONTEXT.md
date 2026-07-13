# hydrate-orchestrator

The pull-model rendering pipeline that coordinates hydration of Kubernetes
objects, Terraform workspaces, and secrets into the wet repository.

## Language

**Image matrix**:
A JSON document listing the container images to update during an image-update
hydration pass, keyed by tenant, app, env, service, and image key.
_Avoid_: image list, image config

**Dry repository**:
The repository containing the source Claims and configuration that hydration
reads from.
_Avoid_: claims repo, source repo
