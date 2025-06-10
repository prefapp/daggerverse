# Kind Dagger Module

## Overview

The `Kind` Dagger module simplifies the creation and management of a local Kubernetes cluster using [Kind](https://kind.sigs.k8s.io/). It automates the setup of a Kind cluster within a Dagger pipeline, allowing developers to test Kubernetes workloads, apply manifests, or debug clusters locally or in CI/CD environments. The module supports loading container images into the cluster, interacting with the cluster via a `k9s` terminal, and inspecting the cluster's container environment.

## Features

- **Kind Cluster Creation**: Sets up a local Kubernetes cluster using Kind (v0.25.0) in a Docker container.
- **Customizable Configuration**: Supports specifying the Kubernetes version and cluster name.
- **Container Image Loading**: Loads container images into the Kind cluster with proper annotations for Kubernetes compatibility.
- **Interactive Debugging**: Provides a `k9s` terminal for exploring the cluster.
- **Cluster Inspection**: Allows inspection of the container running the Kind cluster.
- **Flexible Networking**: Configures the Kind cluster's API server port via a provided service endpoint.

## Usage

The module is initialized with required and optional parameters and provides methods to manage the Kind cluster, load images, and interact with the cluster.

### Module Initialization

```go
kind := dag.Kind(
    dockerSocket,
    kindSvc,
    version,
    clusterName,
)
```

- **dockerSocket** (required): A Dagger socket for Docker (e.g., `/var/run/docker.sock`).
- **kindSvc** (required): A Dagger service with a TCP endpoint (e.g., `tcp://127.0.0.1:3000`) for the Kind cluster's API server.
- **version** (optional): The Kubernetes version in `vx.y` format (e.g., `v1.25`). Must match an available Kind image for Kind v0.25.0. Defaults to Kind's default version if not specified.
- **clusterName** (optional): The name of the Kind cluster. Defaults to `dagger-kubernetes-cluster`.

### Methods

#### `LoadContainerOnKind(container *dagger.Container, tag string) *dagger.Container`

Loads a container image into the Kind cluster.

- **Inputs**:
  - `container`: The Dagger container to load into the cluster.
  - `tag`: The tag for the image (e.g., `my-image`).
- **Behavior**:
  - Annotates the container with Kind-compatible metadata (`org.opencontainers.image.ref.name` and `io.containerd.image.name`).
  - Exports the container as a tarball.
  - Loads the tarball into the Kind cluster using `kind load image-archive`.
- **Returns**: The updated Kind container.
- **Note**: Use `imagePullPolicy: Never` in Kubernetes manifests for images loaded this way, as they are not hosted on a registry.

#### `Knines() *dagger.Container`

Launches a `k9s` terminal for interactive exploration of the Kind cluster.

- **Behavior**: Opens a terminal session in the Kind container running `k9s`.
- **Returns**: The Kind container with the `k9s` terminal session.
- **Usage Example**: `dagger call --docker-socket=/var/run/docker.sock --kind-svc=tcp://127.0.0.1:3000 knines`

#### `Inspect() *dagger.Container`

Provides access to the Kind container's terminal for inspection.

- **Behavior**: Opens a terminal session in the Kind container for debugging or inspection.
- **Returns**: The Kind container with a terminal session.
- **Usage Example**: `dagger call --docker-socket=/var/run/docker.sock --kind-svc=tcp://127.0.0.1:3000 inspect`

#### `Test() *dagger.Container`

A placeholder method for testing purposes.

- **Behavior**: Returns a new container based on `node:20-alpine`. (Note: This method appears incomplete and may be intended for future development.)
- **Returns**: A new Dagger container.

### Configuration Details

- **Kind Cluster Setup**:
  - Uses Kind v0.25.0.
  - Creates a cluster with a YAML configuration specifying the API server port.
  - Installs `docker`, `kubectl`, `k9s`, and `curl` in an Alpine-based container.
  - Deletes any existing cluster with the same name before creating a new one.
  - Configures `kubectl` to connect to the cluster at `https://localhost:<port>`.

- **Networking**:
  - The `kindSvc` endpoint must use `tcp://127.0.0.1:<port>`, where `<port>` is between 1024 and 65535.
  - The `/etc/hosts` file on the host must map `127.0.0.1` to `localhost`.

- **Image Loading**:
  - Images are annotated with Kind-compatible metadata and loaded as tarballs.
  - The `docker.io/library/<tag>:latest` prefix is added to satisfy Kind's requirements, but the image is not pulled from a registry.

### Command to execute

By executing the following command, you get in return a container with kind installed.

`dagger call --docker-socket=/var/run/docker.sock --kind-svc=tcp://127.0.0.1:3000 container`

### Notes

- Ensure the host's `/etc/hosts` includes `127.0.0.1 localhost` to avoid connection issues.
- The port specified in `kindSvc` must be valid (1024–65535).
- The `Test` method is currently a placeholder and may not provide meaningful functionality.
- Use `imagePullPolicy: Never` for manifests using images loaded via `LoadContainerOnKind`.
- Errors during cluster creation or image loading will cause the module to panic with descriptive messages.

## Limitations

- Requires a valid Docker socket and a properly configured `kindSvc` endpoint.
- Only supports Kubernetes versions compatible with Kind v0.25.0.
- The `Test` method is incomplete and may not function as expected.
- Assumes the host environment has `/etc/hosts` configured correctly.
