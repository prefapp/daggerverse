# firestartr-config

Models the `.firestartr` configuration directory that drives Firestartr deployments and validates it against the Firestartr-owned dot-firestartr contracts.

## Schema validation

Configuration under `apps/`, `platforms/`, `docker_registries/`, and `helm_repositories/` is validated **before** decoding/using it against the corresponding contract schemas:

- `apps.schema.json`
- `platforms.schema.json`
- `docker_registries.schema.json`
- `helm_repositories.schema.json`

Runtime schemas are resolved from the #506 docs-repository raw artifacts by default, or from an explicit alternative source (Dagger directory or npm package artifact).

### Source selection, versioning and pinning policy

- `schemaVersion` (default: `v3.2.0`) selects the schema release/path segment. Pass an explicit version (e.g. `v3.2.0`) to pin deterministically, or `latest` to deliberately opt into the latest published schemas (non-deterministic). The requested value is used as-is for docs/raw and directory paths.
- `schemaDirectory` (optional, `*dagger.Directory`): **portable offline/pre-fetched schema source**. When provided, schemas are read via the Dagger API from `<schemaVersion>/<contract>.schema.json` inside the directory and it takes precedence over `schemaBaseURL`. This is the supported way to use schemas available on the caller's machine (pass them in as a directory), since the module runtime cannot read caller-host filesystem paths.
- `schemaBaseURL` (optional) selects the remote schema source when `schemaDirectory` is not provided:
  - empty (default): docs raw base `https://raw.githubusercontent.com/firestartr-pro/docs/refs/heads/main/site/raw/core/dot-firestartr` (the #506-specified source), resolved as `<base>/<schemaVersion>/<contract>.schema.json`.
  - `http(s)` URL: treated as a raw schema file base and resolved as `<base>/<schemaVersion>/<contract>.schema.json` (docs-repository layout).
  - `"npm"` / `"npm:"`: resolve the npm package `@firestartr/firestartr-claims_schemas` from the default npm registry (explicit artifact source).
  - `"npm+http(s)://..."`: resolve the npm package from the given registry/mirror base.
- The module pins `v3.2.0` by default. If the selected remote source cannot resolve/fetch the requested schemas, resolution fails fast with an actionable error; provide `schemaDirectory` with pre-fetched schemas to validate offline.

### Resolution, caching and failure behavior

- raw HTTP/docs source: each contract schema is fetched over HTTP(S) from `<base>/<version>/<contract>.schema.json`.
- npm source (explicit): package metadata and the version tarball are resolved over HTTP(S), tarball integrity is verified (`dist.integrity`, with `dist.shasum` fallback), and `package/dot-firestartr/*.schema.json` contracts are extracted in-memory (pure Go, no Node/npm runtime dependency).
- directory source: contract schemas are read from `schemaDirectory` via the Dagger API at `<schemaVersion>/<contract>.schema.json`.
- Each provider load compiles contract schemas once per `New` invocation and caches compiled schemas in-memory for that load; contract bytes are fetched/read once per contract per load (no per-contract recompilation/refetch).
- Fetch/read/resolve/parse/compile/verification failures and validation failures are surfaced as actionable diagnostics including the file (when applicable), contract, requested/resolved version and schema source (raw URL/base, package/registry/metadata/tarball, or schema directory and path), and cause/reason. Invalid configuration (including registries) fails loading; invalid entries are never silently dropped.
- Runtime validation uses the selected `schemaVersion`/source, then decodes into the generated pinned contract types. If a selected schema (e.g. `latest` or a different version) diverges incompatibly from the pinned/generated shape, validation/decoding may fail with an actionable diagnostic rather than silently dropping required data.

## Schema-derived types

Contract model types (`FirestartrApp`, `FirestartrService`, `Registry`, `BasePaths`, `Repository`, `Platform`, `AllowedClaim`, `AuthStrategy`) are **generated from the pinned `v3.2.0` schema snapshot** and must not be hand-edited.

- Source snapshot: `testdata/dot-firestartr/v3.2.0/*.schema.json`
- Generator: `tools/gentypes` (invoked via `generate.go` / `go:generate`)
- Generated output: `types_gen.go`

To refresh derived types from the pinned schemas:

```sh
cd firestartr-config
go generate ./...
dagger develop
```

When bumping the default pinned schema version, update the snapshot in `testdata/dot-firestartr/<version>/`, update `DefaultSchemaVersion`/generator inputs, regenerate types, and run tests. Runtime validation always uses the explicitly selected `schemaVersion`/`schemaBaseURL`; decoding targets the generated (pinned) types after successful validation. If a selected schema (e.g. `latest`) diverges incompatibly from the pinned/generated shape, validation/decoding may fail with an actionable diagnostic rather than silently dropping required data.

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

## Compatibility impact

- Constructor is schema-aware: `New(dotFirestartr, schemaVersion?, schemaBaseURL?, schemaDirectory?)` with default pinned `schemaVersion` `v3.2.0`.
- Runtime schema source defaults to the #506 docs-repository raw artifacts (`<base>/<schemaVersion>/<contract>.schema.json`, default base `https://raw.githubusercontent.com/firestartr-pro/docs/refs/heads/main/site/raw/core/dot-firestartr`). `schemaBaseURL` selects the remote source when `schemaDirectory` is not provided: empty = default docs raw base, `http(s)` = raw schema file base, `"npm"` = npm package `@firestartr/firestartr-claims_schemas` (explicit artifact source), `"npm+http(s)://..."` = npm registry/mirror. For offline/pre-fetched schemas, pass `schemaDirectory` (`*dagger.Directory`) containing `<schemaVersion>/<contract>.schema.json` (read via the Dagger API); it takes precedence over `schemaBaseURL`. Caller-host filesystem paths in `schemaBaseURL` are not a supported offline fallback (the module runtime cannot read them).
- Apps: service definitions use `services` (no longer `image_types`); valid service lists deserialize as `[{repo, service_names, platforms?}]`.
- Docker registries: required `name`, `registry` (host only, no scheme/port), `image_types`, `default`, `auth_strategy` (`aws_oidc|azure_oidc|generic|ghcr|dockerhub|none`), and `base_paths` with required `services` (optional `charts`). Invalid registries now fail validation instead of being silently dropped.
- Helm repositories: require non-empty `name` and `url` beginning with `http://`, `https://`, or `oci://`.
- Platforms: `platforms/` remains canonical and both `.yaml` and `.yml` files are supported/validated.
- Downstream Dagger consumers must regenerate bindings/migrate to the updated API and config shapes/semantics (separate consumer migration). If a selected runtime schema (e.g. `latest`) diverges incompatibly from the pinned/generated types, validation/decoding may fail with an actionable diagnostic.
