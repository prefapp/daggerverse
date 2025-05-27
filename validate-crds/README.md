# ValidateCrds Dagger Module

## Overview

The `ValidateCrds` Dagger module automates the validation of Kubernetes Custom Resource Definitions (CRDs) and Custom Resources (CRs) in a local [Kind](https://kind.sigs.k8s.io/) cluster. It creates a Kind cluster, applies CRDs, and then applies CRs to ensure they are valid and compatible with the defined CRDs.

## Features

- **Kind Cluster Integration**: Leverages the `prefapp/daggerverse/kind` module to create and manage a local Kubernetes cluster using Kind.
- **CRD Application**: Applies CRD manifests from a specified directory to the Kind cluster.
- **CR Validation**: Applies CR manifests and validates them against the deployed CRDs.
- **Customizable Kubernetes Version**: Allows specifying the Kubernetes version for the Kind cluster (optional).
- **Directory-Based Input**: Accepts directories containing CRD and CR manifests, supporting nested directories for CRs.

## Prerequisites

- **CRD and CR Manifests**: Prepare directories with valid CRD and CR YAML/JSON files.

## Usage

The module is initialized with required and optional parameters and provides methods to apply CRDs, create CRs, and validate them.

### Module Initialization

```go
validateCrds := dag.ValidateCrds(
    dockerSocket,
    kindSvc,
    crdsDir,
    crsDir,
    version,
)
```

- **dockerSocket**: A Dagger socket for Docker, required by `prefapp/daggerverse/kind.
- **kindSvc**: A Dagger service for the Kind cluster, required by `prefapp/daggerverse/kind`.
- **crdsDir**: A directory containing CRD manifest files (YAML/JSON).
- **crsDir**: A directory containing CR manifest files, which can include nested directories.
- **version** (optional): Specifies the Kubernetes version for the Kind cluster (e.g., `v1.30`). Defaults to the Kind module's default version if not provided.

### Methods

#### `CreateCRDS() *dagger.Container`

Applies the CRD manifests from the provided `crdsDir` to the Kind cluster.

- **Input**: Uses the `crdsDir` directory mounted to `/crds` in the Kind container.
- **Behavior**: Executes `kubectl apply -f .` to apply all CRD manifests in the directory.
- **Returns**: A Dagger container with the applied CRDs.

#### `CreateCRS(ctr *dagger.Container) (string, error)`

Applies CR manifests from the provided `crsDir` to the Kind cluster.

- **Input**: 
  - `ctr`: A Dagger container with CRDs already applied (typically the one returned by `CreateCRDS`).
  - Uses the `crsDir` directory mounted to `/crs` in the container.
- **Behavior**: Executes `kubectl apply -R -f .` to recursively apply all CR manifests in the directory and its subdirectories.
- **Returns**: The `kubectl` command output as a string and any error encountered.

#### `Validate() (string, error)`

Combines `CreateCRDS` and `CreateCRS` to validate CRs against CRDs.

- **Behavior**: 
  1. Applies CRDs using `CreateCRDS`.
  2. Applies CRs using `CreateCRS` on the resulting container.
- **Returns**: The output of the CR application process and any error encountered.

### Directory Structure Example

```
crds/
├── crd1.yaml
├── crd2.yaml
crs/
├── namespace1/
│   ├── cr1.yaml
│   ├── cr2.yaml
├── namespace2/
│   ├── cr3.yaml
```

- **crdsDir**: Contains CRD manifests (e.g., `crd1.yaml`, `crd2.yaml`).
- **crsDir**: Contains CR manifests, optionally organized in subdirectories (e.g., `namespace1/cr1.yaml`).

### Command to execute

```bash
dagger call \
    --docker-socket=/var/run/docker.sock \
    --kind-svc=tcp://127.0.0.1:3000 \
    --version v1_32 \
    --crds-dir ./crds \
    --crs-dir ./crs \
    validate
```

## Limitations

- Requires a valid Docker socket and Kind service.
- Only supports Kubernetes versions available in the Kind module.
- Does not perform advanced validation beyond applying CRs (e.g., no schema validation).
